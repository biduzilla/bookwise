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

	"github.com/lib/pq"
)

type bookRepository struct {
	db     *sql.DB
	logger jsonlog.Logger
}

func NewBookRepository(
	db *sql.DB,
	logger jsonlog.Logger,
) *bookRepository {
	return &bookRepository{
		db:     db,
		logger: logger,
	}
}

type BookRepository interface {
	GetAll(title, author string,
		userID int64,
		f filters.Filters,
	) ([]*models.Book, filters.Metadata, error)
	GetByID(bookID, userID int64) (*models.Book, error)
	Insert(tx *sql.Tx, book *models.Book) error
	Update(tx *sql.Tx, book *models.Book, userID int64) error
	Delete(tx *sql.Tx, bookID, userID int64) error
}

func parseBookConstraintError(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Constraint {
		case "unique_title_per_user":
			return e.ErrBookTitle
		case "chk_books_pages_positive":
			return e.ErrBookPages
		}
	}
	return err
}

func (r *bookRepository) GetAll(title, author string,
	userID int64,
	f filters.Filters,
) ([]*models.Book, filters.Metadata, error) {
	cols := strings.Join([]string{
		selectColumns(models.Book{}, "b"),
		selectColumns(models.User{}, "u"),
	}, ", ")
	query := fmt.Sprintf(`
        SELECT
            count(*) OVER(),
           	%s
        FROM books b
        LEFT JOIN users u ON u.id = b.user_id
        WHERE
            (to_tsvector('simple', b.title) @@ plainto_tsquery('simple', :title) OR :title = '')
            AND (to_tsvector('simple', b.author) @@ plainto_tsquery('simple', :author) OR :author = '')
            AND b.deleted = false
			and b.user_id = :userID
        ORDER BY
            b.%s %s,
            b.id ASC
        LIMIT :limit
        OFFSET :offset
    `, cols, f.SortColumn(), f.SortDirection())

	params := map[string]any{
		"title":  title,
		"author": author,
		"userID": userID,
		"limit":  f.Limit(),
		"offset": f.Offset(),
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	return paginatedQuery(
		r.db,
		query,
		args,
		f,
		func() *models.Book {
			return &models.Book{
				User: &models.User{},
			}
		},
	)
}

func (r *bookRepository) GetByID(bookID, userID int64) (*models.Book, error) {
	cols := strings.Join([]string{
		selectColumns(models.Book{}, "b"),
		selectColumns(models.User{}, "u"),
	}, ", ")

	query := fmt.Sprintf(`
    select
        %s
    from books b
    left join users u on u.id = b.user_id
    where
        b.id = :bookID
        and b.user_id = :userID
        and b.deleted = false
`, cols)

	params := map[string]any{
		"bookID": bookID,
		"userID": userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)
	return getByQuery[models.Book](r.db, query, args)
}

func (r *bookRepository) Insert(tx *sql.Tx, book *models.Book) error {
	query := `
	insert into books(
		title,
		author,
		pages,
		description,
		user_id,
		created_by
	)
	values (
		:title,
		:author,
		:pages,
		:description,
		:user_id,
		:user_id
	)
	returning id, created_at, version
	`

	params := map[string]any{
		"title":       book.Title,
		"author":      book.Author,
		"pages":       book.Pages,
		"description": book.Description,
		"user_id":     book.User.ID,
	}
	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&book.ID,
		&book.CreatedAt,
		&book.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrRecordNotFound
		}

		return parseBookConstraintError(err)
	}

	return nil
}

func (r *bookRepository) Update(tx *sql.Tx, book *models.Book, userID int64) error {
	query := `
	update books set
		title = :title,
		author = :author,
		pages = :pages,
		description = :description,
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
		"title":       book.Title,
		"author":      book.Author,
		"pages":       book.Pages,
		"description": book.Description,
		"user_id":     book.User.ID,
		"version":     book.Version,
		"id":          book.ID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&book.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrEditConflict
		}

		return parseBookConstraintError(err)
	}
	return nil
}

func (r *bookRepository) Delete(tx *sql.Tx, bookID, userID int64) error {
	query := `
	update books set
		deleted = false
	where 
		id = :id
		and user_id = :userID
	`

	params := map[string]any{
		"id":     bookID,
		"userID": userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := tx.ExecContext(ctx, query, args)
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
