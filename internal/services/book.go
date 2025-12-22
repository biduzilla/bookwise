package services

import (
	"bookwise/internal/models"
	"bookwise/internal/models/filters"
	"bookwise/internal/repositories"
	"bookwise/utils"
	e "bookwise/utils/errors"
	"bookwise/utils/validator"
	"database/sql"
)

type bookService struct {
	book repositories.BookRepository
	db   *sql.DB
}

type BookService interface {
	FindAll(
		title, author string,
		userID int64,
		f filters.Filters,
	) ([]*models.Book, filters.Metadata, error)
	Save(book *models.Book, userID int64, v *validator.Validator) error
	FindByID(id, userID int64) (*models.Book, error)
	Update(book *models.Book, userID int64, v *validator.Validator) error
	Delete(id, userID int64) error
}

func NewBookService(
	book repositories.BookRepository,
	db *sql.DB,
) *bookService {
	return &bookService{
		book: book,
		db:   db,
	}
}

func (s *bookService) FindAll(
	title, author string,
	userID int64,
	f filters.Filters,
) ([]*models.Book, filters.Metadata, error) {
	return s.book.GetAll(title, author, userID, f)
}

func (s *bookService) Save(book *models.Book, userID int64, v *validator.Validator) error {
	if book.User == nil {
		book.User = &models.User{
			ID: userID,
		}
	}

	if book.ValidateBook(v); !v.Valid() {
		return e.ErrInvalidData
	}

	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		return s.book.Insert(tx, book)
	})
}

func (s *bookService) FindByID(id, userID int64) (*models.Book, error) {
	return s.book.GetByID(id, userID)
}

func (s *bookService) Update(book *models.Book, userID int64, v *validator.Validator) error {
	if book.ValidateBook(v); !v.Valid() {
		return e.ErrInvalidData
	}

	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		return s.book.Update(tx, book, userID)
	})
}

func (s *bookService) Delete(id, userID int64) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		return s.book.Delete(tx, id, userID)
	})
}
