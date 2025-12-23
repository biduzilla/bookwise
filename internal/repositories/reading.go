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

type readingPlanRepository struct {
	db     *sql.DB
	logger jsonlog.Logger
}

type readingSession struct {
	db     *sql.DB
	logger jsonlog.Logger
}

func NewReadingPlanRepository(db *sql.DB,
	logger jsonlog.Logger,
) *readingPlanRepository {
	return &readingPlanRepository{
		db:     db,
		logger: logger,
	}
}

func NewReadingSessionRepository(db *sql.DB,
	logger jsonlog.Logger,
) *readingSession {
	return &readingSession{
		db:     db,
		logger: logger,
	}
}

type ReadingPlanRepository interface {
	GetAll(
		status *models.ReadingStatus,
		startDate *time.Time,
		targetDate *time.Time,
		userID, bookID int64,
		f filters.Filters,
	) ([]*models.ReadingPlan, filters.Metadata, error)
	GetByID(id, userID int64) (*models.ReadingPlan, error)
	Insert(
		tx *sql.Tx,
		plan *models.ReadingPlan,
	) error
	Update(
		tx *sql.Tx,
		plan *models.ReadingPlan,
		userID int64,
	) error
	Delete(
		tx *sql.Tx,
		planID int64,
		userID int64,
	) error
}

func (r *readingPlanRepository) GetAll(
	status *models.ReadingStatus,
	startDate *time.Time,
	targetDate *time.Time,
	userID, bookID int64,
	f filters.Filters,
) ([]*models.ReadingPlan, filters.Metadata, error) {
	cols := strings.Join([]string{
		selectColumns(models.ReadingPlan{}, "r"),
		selectColumns(models.Book{}, "b"),
		selectColumns(models.User{}, "u"),
	}, ", ")

	query := fmt.Sprintf(`
        SELECT
            count(*) OVER(),
           	%s
        FROM reading_plans r
        LEFT JOIN users u ON u.id = r.user_id
        LEFT JOIN books b ON b.id = r.book_id
        WHERE
            (:status is null or r.status = :status)
			AND (:startDate::timestamptz IS NULL OR r.start_date >= :startDate::timestamptz)
			AND (:targetDate::timestamptz IS NULL OR r.target_date <= :targetDate::timestamptz)
            AND b.deleted = false
			and r.user_id = :userID
			and r.book_id = :bookID
        ORDER BY
            r.%s %s,
            r.id ASC
        LIMIT :limit
        OFFSET :offset
    `, cols, f.SortColumn(), f.SortDirection())

	start := sql.NullTime{}
	if startDate != nil {
		start.Valid = true
		start.Time = *startDate
	}

	target := sql.NullTime{}
	if targetDate != nil {
		target.Valid = true
		target.Time = targetDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	params := map[string]any{
		"startDate":  startDate,
		"targetDate": target,
		"userID":     userID,
		"bookID":     bookID,
		"status":     status,
		"limit":      f.Limit(),
		"offset":     f.Offset(),
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	return paginatedQuery(
		r.db,
		query,
		args,
		f,
		func() *models.ReadingPlan {
			return &models.ReadingPlan{
				User: &models.User{},
				Book: &models.Book{},
			}
		},
	)
}

func (r *readingPlanRepository) GetByID(id, userID int64) (*models.ReadingPlan, error) {
	cols := strings.Join([]string{
		selectColumns(models.ReadingPlan{}, "r"),
		selectColumns(models.Book{}, "b"),
		selectColumns(models.User{}, "u"),
	}, ", ")
	query := fmt.Sprintf(`
	select
		%s
	FROM reading_plans r
    LEFT JOIN users u ON u.id = r.user_id
    LEFT JOIN books b ON b.id = r.book_id	
	where
		r.id = :id,
		and r.user_id = :userID
		and r.deleted = false
	`, cols)

	params := map[string]any{
		"id":     id,
		"userID": userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)
	return getByQuery[models.ReadingPlan](r.db, query, args)
}

func (r *readingPlanRepository) Insert(
	tx *sql.Tx,
	plan *models.ReadingPlan,
) error {
	query := `
	insert into reading_plans (
		status,
		start_date,
		target_date,
		priority,
		pages_per_day,
		minutes_per_day,
		book_id,
		user_id,
		created_by
	)
	values (
		:status,
		:start_date,
		:target_date,
		:priority,
		:pages_per_day,
		:minutes_per_day,
		:book_id,
		:user_id,
		:user_id
	)
	returning id, created_at, version
	`

	params := map[string]any{
		"status":          plan.Status,
		"start_date":      plan.StartDate,
		"target_date":     plan.TargetDate,
		"priority":        plan.Priority,
		"pages_per_day":   plan.PagesPerDay,
		"minutes_per_day": plan.MinutesPerDay,
		"book_id":         plan.Book.ID,
		"user_id":         plan.User.ID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&plan.ID,
		&plan.CreatedAt,
		&plan.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrRecordNotFound
		}
		return err
	}

	return nil
}

func (r *readingPlanRepository) Update(
	tx *sql.Tx,
	plan *models.ReadingPlan,
	userID int64,
) error {
	query := `
	update reading_plans set
		status = :status,
		start_date = :start_date,
		target_date = :target_date,
		priority = :priority,
		pages_per_day = :pages_per_day,
		minutes_per_day = :minutes_per_day,
		book_id = :book_id,
		updated_at = now(),
		updated_by = :user_id,
		version = version + 1
	where
		id = :id
		and version = :version
		and deleted = false
		and user_id = :user_id
	returning version
	`

	params := map[string]any{
		"id":              plan.ID,
		"status":          plan.Status,
		"start_date":      plan.StartDate,
		"target_date":     plan.TargetDate,
		"priority":        plan.Priority,
		"pages_per_day":   plan.PagesPerDay,
		"minutes_per_day": plan.MinutesPerDay,
		"book_id":         plan.Book.ID,
		"user_id":         userID,
		"version":         plan.Version,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, args...).Scan(&plan.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrEditConflict
		}
		return err
	}

	return nil
}

func (r *readingPlanRepository) Delete(
	tx *sql.Tx,
	planID int64,
	userID int64,
) error {
	query := `
	update reading_plans set
		deleted = true,
		updated_at = now(),
		updated_by = :user_id
	where
		id = :id
		and user_id = :user_id
		and deleted = false
	`

	params := map[string]any{
		"id":      planID,
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
