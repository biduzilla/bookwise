package services

import (
	"bookwise/internal/models"
	"bookwise/internal/models/filters"
	"bookwise/internal/repositories"
	"bookwise/utils"
	"bookwise/utils/errors"
	"bookwise/utils/validator"
	"database/sql"
	"time"
)

type readingSessionService struct {
	readingSession repositories.ReadingSessionRepository
	db             *sql.DB
}

func NewReadingSessionService(
	readingSession repositories.ReadingSessionRepository,
	db *sql.DB,
) *readingSessionService {
	return &readingSessionService{
		readingSession: readingSession,
		db:             db,
	}
}

type ReadingSessionService interface {
	FindAll(
		date *time.Time,
		userID int64,
		readingPlanID int64,
		f filters.Filters,
	) ([]*models.ReadingSession, filters.Metadata, error)
	Save(model *models.ReadingSession, userID int64, v *validator.Validator) error
	FindByID(id, userID int64) (*models.ReadingSession, error)
	Update(model *models.ReadingSession, userID int64, v *validator.Validator) error
	Delete(id, userID int64) error
}

func (s *readingSessionService) FindAll(
	date *time.Time,
	userID int64,
	readingPlanID int64,
	f filters.Filters,
) ([]*models.ReadingSession, filters.Metadata, error) {
	return s.readingSession.GetAll(date, userID, readingPlanID, f)
}

func (s *readingSessionService) Save(model *models.ReadingSession, userID int64, v *validator.Validator) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		if model.ValidateReadingSession(v); !v.Valid() {
			return errors.ErrInvalidData
		}

		return s.readingSession.Insert(tx, model, userID)
	})
}

func (s *readingSessionService) FindByID(id, userID int64) (*models.ReadingSession, error) {
	return s.readingSession.GetByID(id, userID)
}

func (s *readingSessionService) Update(model *models.ReadingSession, userID int64, v *validator.Validator) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		if model.ValidateReadingSession(v); !v.Valid() {
			return errors.ErrInvalidData
		}

		return s.readingSession.Update(tx, model, userID)
	})
}

func (s *readingSessionService) Delete(id, userID int64) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		return s.readingSession.Delete(tx, id, userID)
	})
}
