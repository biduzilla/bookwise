package repositories

import (
	"bookwise/internal/jsonlog"
	"bookwise/internal/models"
	"bookwise/utils"
	e "bookwise/utils/errors"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type UserRepository struct {
	db     *sql.DB
	logger jsonlog.Logger
}

type UserRepositoryInterface interface {
	GetByCodAndEmail(cod int, email string) (*models.User, error)
	GetByID(id int64) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Insert(tx *sql.Tx, user *models.User) error
	UpdateCodByEmail(tx *sql.Tx, user *models.User) error
	Update(tx *sql.Tx, user *models.User) error
	Delete(tx *sql.Tx, idUser int64) error
}

func NewUserRepository(
	db *sql.DB,
	logger jsonlog.Logger,
) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepository) getUserByQuery(query string, args ...any) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User
	row := r.db.QueryRowContext(ctx, query, args...)
	err := scanStruct(row, &user)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, e.ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func parseUserConstraintError(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Constraint {
		case "users_email_key":
			return e.ErrDuplicateEmail
		case "users_phone_key":
			return e.ErrDuplicatePhone
		}
	}
	return err
}

func (r *UserRepository) GetByCodAndEmail(cod int, email string) (*models.User, error) {
	query := `
	select u.* 
	from users u
	WHERE
		email = :email
		AND deleted = false
		AND cod = :cod
	`
	params := map[string]any{
		"email": email,
		"cod":   cod,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)
	return getByQuery[models.User](r.db, query, args)
}

func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	query := `
	select u.* 
	from users u
	WHERE
		id = :id
		AND deleted = false
	`

	params := map[string]any{
		"id": id,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)
	return getByQuery[models.User](r.db, query, args)
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
	select u.* 
	from users u
	WHERE
		email = :email
		AND deleted = false
	`
	params := map[string]any{
		"email": email,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	return getByQuery[models.User](r.db, query, args)
}

func (r *UserRepository) Insert(tx *sql.Tx, user *models.User) error {
	query := `
	INSERT INTO users (name, email, phone,cod, password_hash, activated,deleted)
	VALUES ($1, $2, $3, $4, $5, $6,false)
	RETURNING id, created_at, version
	`
	args := []any{
		user.Name,
		user.Email,
		user.Phone,
		user.Cod,
		user.Password.Hash,
		user.Activated,
	}

	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrEditConflict
		}

		return parseUserConstraintError(err)
	}

	return nil
}

func (r *UserRepository) UpdateCodByEmail(tx *sql.Tx, user *models.User) error {
	query := `
	UPDATE users SET
	cod = $1
	WHERE id = $2 AND version = $3
	RETURNING version`

	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, user.Cod, user.ID, user.Version).Scan(
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return e.ErrRecordNotFound
		default:
			return err
		}
	}
	return nil

}

func (r *UserRepository) Update(tx *sql.Tx, user *models.User) error {
	query := `
	UPDATE users SET
		name = $1,
		email = $2,
		cod = $3,
		phone = $4,
		password_hash = $5,
		activated = $6,
		version = version + 1
	WHERE
		id = $7
		AND version = $8
	RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Cod,
		user.Phone,
		user.Password.Hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&user.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrEditConflict
		}

		return parseUserConstraintError(err)
	}
	return nil
}

func (r *UserRepository) Delete(tx *sql.Tx, idUser int64) error {
	query := `
	UPDATE users set
	deleted = true
	where id = $1
	`

	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := tx.ExecContext(ctx, query, idUser)
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
