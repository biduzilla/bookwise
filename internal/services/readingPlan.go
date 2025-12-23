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

type readingPlanService struct {
	readingPlan repositories.ReadingPlanRepository
	db          *sql.DB
}

func NewReadingPlanService(
	readingPlan repositories.ReadingPlanRepository,
	db *sql.DB,
) *readingPlanService {
	return &readingPlanService{
		readingPlan: readingPlan,
		db:          db,
	}
}

type ReadingPlanService interface {
	FindAll(
		status *models.ReadingStatus,
		startDate *time.Time,
		targetDate *time.Time,
		userID, bookID int64,
		f filters.Filters,
	) ([]*models.ReadingPlan, filters.Metadata, error)
	Save(model *models.ReadingPlan, userID int64, v *validator.Validator) error
	FindByID(id, userID int64) (*models.ReadingPlan, error)
	Update(model *models.ReadingPlan, userID int64, v *validator.Validator) error
	Delete(id, userID int64) error
}

func (s *readingPlanService) FindAll(
	status *models.ReadingStatus,
	startDate *time.Time,
	targetDate *time.Time,
	userID, bookID int64,
	f filters.Filters,
) ([]*models.ReadingPlan, filters.Metadata, error) {
	return s.readingPlan.GetAll(status, startDate, targetDate, userID, bookID, f)
}

func (s *readingPlanService) FindByID(id, userID int64) (*models.ReadingPlan, error) {
	return s.readingPlan.GetByID(id, userID)
}

func (s *readingPlanService) Save(model *models.ReadingPlan, userID int64, v *validator.Validator) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		if model.ValidateReadingPlan(v); !v.Valid() {
			return errors.ErrInvalidData
		}

		if model.User == nil {
			model.User = &models.User{
				ID: userID,
			}
		}

		return s.readingPlan.Insert(tx, model)
	})
}

func (s *readingPlanService) Update(model *models.ReadingPlan, userID int64, v *validator.Validator) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		if model.ValidateReadingPlan(v); !v.Valid() {
			return errors.ErrInvalidData
		}

		return s.readingPlan.Update(tx, model, userID)
	})
}

func (s *readingPlanService) Delete(id, userID int64) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		return s.readingPlan.Delete(tx, id, userID)
	})
}
