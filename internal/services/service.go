package services

import (
	"bookwise/internal/config"
	"bookwise/internal/jsonlog"
	"bookwise/internal/models"
	"bookwise/internal/repositories"
	"bookwise/utils/validator"
	"database/sql"
)

type GenericServiceInterface[
	T models.ModelInterface[D],
	D any,
] interface {
	Save(entity *T, userID int64, v *validator.Validator) error
	FindByID(id, userID int64) (*T, error)
	Update(entity *T, userID int64, v *validator.Validator) error
	Delete(id, userID int64) error
}

type Services struct {
	User UserService
	Auth AuthServiceInterface
	Book BookService
}

func NewServices(logger jsonlog.Logger, db *sql.DB, config config.Config) *Services {
	r := repositories.NewRepository(logger, db)
	userService := NewUserService(r.User, db)

	return &Services{
		User: userService,
		Auth: NewAuthService(userService, config),
		Book: NewBookService(r.Book, db),
	}

}
