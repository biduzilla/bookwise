package services

import (
	"bookwise/internal/models"
	"bookwise/internal/repositories"
	"bookwise/utils"
	e "bookwise/utils/errors"
	"bookwise/utils/validator"
	"database/sql"
	"errors"
)

type userService struct {
	user repositories.UserRepositoryInterface
	db   *sql.DB
}

type UserService interface {
	GetUserByEmail(email string, v *validator.Validator) (*models.User, error)
	ActivateUser(cod int, email string, v *validator.Validator) (*models.User, error)
	Update(user *models.User, v *validator.Validator) error
	GetUserByCodAndEmail(cod int, email string, v *validator.Validator) (*models.User, error)
	Save(user *models.User, v *validator.Validator) error
}

func NewUserService(
	userRepository repositories.UserRepositoryInterface,
	db *sql.DB,
) *userService {
	return &userService{
		user: userRepository,
		db:   db,
	}
}

func (s *userService) GetUserByEmail(email string, v *validator.Validator) (*models.User, error) {
	user, err := s.user.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) ActivateUser(cod int, email string, v *validator.Validator) (*models.User, error) {
	if models.ValidateEmail(v, email); !v.Valid() {
		return nil, e.ErrInvalidData
	}

	user, err := s.user.GetByCodAndEmail(cod, email)

	if err != nil {
		return nil, err
	}

	user.Activated = true
	user.Cod = 0

	if err = s.Update(user, v); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Update(user *models.User, v *validator.Validator) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		err := s.user.Update(tx, user)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *userService) GetUserByCodAndEmail(cod int, email string, v *validator.Validator) (*models.User, error) {
	user, err := s.user.GetByCodAndEmail(cod, email)
	if err != nil {
		switch {
		case errors.Is(err, e.ErrRecordNotFound):
			v.AddError("code", "invalid validation code or email")
			return nil, e.ErrInvalidData
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *userService) Save(user *models.User, v *validator.Validator) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		if user.ValidateUser(v); !v.Valid() {
			return e.ErrInvalidData
		}
		user.Cod = utils.GenerateRandomCode()
		return s.user.Insert(tx, user)
	})
}

func (s *userService) Delete(idUser int64) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		return s.user.Delete(tx, idUser)
	})
}
