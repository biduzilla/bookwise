package repositories

import (
	"bookwise/internal/jsonlog"
	"bookwise/internal/models"
	"bookwise/internal/models/filters"
	"bookwise/utils"
	e "bookwise/utils/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type readingSessionRepository struct {
	db     *sql.DB
	logger jsonlog.Logger
}

func NewReadingSessionRepository(db *sql.DB,
	logger jsonlog.Logger,
) *readingSessionRepository {
	return &readingSessionRepository{
		db:     db,
		logger: logger,
	}
}

type ReadingSessionRepository interface {
	GetAll(
		date *time.Time,
		userID, readingPlanID int64,
		f filters.Filters,
	) ([]*models.ReadingSession, filters.Metadata, error)

	GetByID(id, userID int64) (*models.ReadingSession, error)
	Insert(tx *sql.Tx, session *models.ReadingSession, userID int64) error
	Update(tx *sql.Tx, session *models.ReadingSession, userID int64) error
	Delete(tx *sql.Tx, sessionID, userID int64) error
}

func (r *readingSessionRepository) GetAll(
	date *time.Time,
	userID, readingPlanID int64,
	f filters.Filters,
) ([]*models.ReadingSession, filters.Metadata, error) {
	cols := strings.Join([]string{
		selectColumns(models.ReadingPlan{}, "r"),
		selectColumns(models.ReadingSession{}, "rs"),
	}, ", ")

	query := fmt.Sprintf(`
        SELECT
            count(*) OVER(),
           	%s
        FROM reading_session rs
        LEFT JOIN reading_plans r ON r.id = rs.reading_plan_id
        WHERE
			(:date::timestamptz IS NULL OR r.start_date >= :date::timestamptz)
            AND rs.deleted = false
			and r.user_id = :userID
			and rs.reading_plan_id = :readingPlanID
        ORDER BY
            r.%s %s,
            r.id ASC
        LIMIT :limit
        OFFSET :offset
    `, cols, f.SortColumn(), f.SortDirection())

	dateSql := sql.NullTime{}
	if date != nil {
		dateSql.Valid = true
		dateSql.Time = *date
	}

	params := map[string]any{
		"date":          date,
		"readingPlanID": readingPlanID,
		"userID":        userID,
		"limit":         f.Limit(),
		"offset":        f.Offset(),
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	return paginatedQuery(
		r.db,
		query,
		args,
		f,
		func() *models.ReadingSession {
			return &models.ReadingSession{
				ReadingPlan: &models.ReadingPlan{},
			}
		},
	)
}

func (r *readingSessionRepository) GetByID(id, userID int64) (*models.ReadingSession, error) {
	cols := strings.Join([]string{
		selectColumns(models.ReadingPlan{}, "r"),
		selectColumns(models.ReadingSession{}, "rs"),
	}, ", ")

	query := fmt.Sprintf(`
	select
		%s
	FROM reading_session rs
    LEFT JOIN reading_plans r ON r.id = rs.reading_plan_id	
	where
		rs.id = :id,
		and r.user_id = :userID
		and rs.deleted = false
	`, cols)

	params := map[string]any{
		"id":     id,
		"userID": userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)
	return getByQuery[models.ReadingSession](r.db, query, args)
}

func (r *readingSessionRepository) Insert(
	tx *sql.Tx,
	session *models.ReadingSession,
	userID int64,
) error {
	query := `
	insert into reading_session (
		pages_read,
		minutes,
		notes,
		date,
		reading_plan_id,
		created_by
	)
	values (
		:pages_read,
		:minutes,
		:notes,
		:date,
		:reading_plan_id,
		:user_id
	)
	returning id, created_at, version
	`

	params := map[string]any{
		"pages_read":      session.PagesRead,
		"minutes":         session.Minutes,
		"notes":           session.Notes,
		"date":            session.Date,
		"reading_plan_id": session.ReadingPlan.ID,
		"user_id":         userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&session.ID,
		&session.CreatedAt,
		&session.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrRecordNotFound
		}
		return err
	}

	return nil
}

func (r *readingSessionRepository) Update(
	tx *sql.Tx,
	session *models.ReadingSession,
	userID int64,
) error {
	query := `
	update reading_session set
		pages_read = :pages_read,
		minutes = :minutes,
		notes = :notes,
		date = :date,
		updated_at = now(),
		updated_by = :user_id,
		version = version + 1
	where
		id = :id
		and version = :version
		and deleted = false
		and reading_plan_id in (
			select id from reading_plans
			where user_id = :user_id
		)
	returning version
	`

	params := map[string]any{
		"id":         session.ID,
		"pages_read": session.PagesRead,
		"minutes":    session.Minutes,
		"notes":      session.Notes,
		"date":       session.Date,
		"version":    session.Version,
		"user_id":    userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&session.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrEditConflict
		}
		return err
	}

	return nil
}

func (r *readingSessionRepository) Delete(
	tx *sql.Tx,
	sessionID, userID int64,
) error {
	query := `
	update reading_session set
		deleted = true,
		updated_at = now(),
		updated_by = :user_id
	where
		id = :id
		and deleted = false
		and reading_plan_id in (
			select id from reading_plans
			where user_id = :user_id
		)
	`

	params := map[string]any{
		"id":      sessionID,
		"user_id": userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return e.ErrRecordNotFound
	}

	return nil
}
