package models

import (
	"bookwise/utils/validator"
	"time"
)

type ReadingStatus int
type ReadingPriority int

const (
	ReadingStatusPlanned ReadingStatus = iota + 1
	ReadingStatusReading
	ReadingStatusCompleted
	ReadingStatusPaused
)

const (
	ReadingPriorityLow ReadingPriority = iota + 1
	ReadingPriorityMedium
	ReadingPriorityHigh
)

type ReadingPlan struct {
	ID            int64           `db:"id"`
	Status        ReadingStatus   `db:"status"`
	StartDate     *time.Time      `db:"start_date"`
	TargetDate    *time.Time      `db:"target_date"`
	Priority      ReadingPriority `db:"priority"`
	PagesPerDay   int             `db:"pages_per_day"`
	MinutesPerDay int             `db:"minutes_per_day"`
	Book          *Book           `db:"book_id"`
	User          *User           `db:"user_id"`

	BaseModel
}

type ReadingSession struct {
	ID          string
	ReadingPlan ReadingPlan
	PagesRead   int
	Minutes     int
	Notes       string
	Date        time.Time
	BaseModel
}

type ReadingPlanDTO struct {
	ID            *int64           `json:"id" dto:"ID"`
	Status        *ReadingStatus   `json:"status" dto:"Status"`
	StartDate     *time.Time       `json:"startDate" dto:"StartDate"`
	TargetDate    *time.Time       `json:"targetDate" dto:"TargetDate"`
	Priority      *ReadingPriority `json:"priority" dto:"Priority"`
	PagesPerDay   *int             `json:"pagesPerDay" dto:"PagesPerDay"`
	MinutesPerDay *int             `json:"minutesPerDay" dto:"MinutesPerDay"`
	Book          *BookDTO         `json:"book" dto:"Book"`
	User          *UserDTO         `json:"user" dto:"User"`
}

type ReadingSessionDTO struct {
	ID          *string         `json:"id" dto:"ID"`
	ReadingPlan *ReadingPlanDTO `json:"readingPlan" dto:"ReadingPlan"`
	PagesRead   *int            `json:"pagesRead" dto:"PagesRead"`
	Minutes     *int            `json:"minutes" dto:"Minutes"`
	Notes       *string         `json:"notes" dto:"Notes"`
	Date        *time.Time      `json:"date" dto:"Date"`
}

func (dto ReadingSessionDTO) ToModel() *ReadingSession {
	var model ReadingSession

	if dto.ID != nil {
		model.ID = *dto.ID
	}

	if dto.ReadingPlan != nil {
		model.ReadingPlan = *dto.ReadingPlan.ToModel()
	}

	if dto.PagesRead != nil {
		model.PagesRead = *dto.PagesRead
	}

	if dto.Minutes != nil {
		model.Minutes = *dto.Minutes
	}

	if dto.Notes != nil {
		model.Notes = *dto.Notes
	}

	if dto.Date != nil {
		model.Date = *dto.Date
	}

	return &model
}

func (m ReadingSession) ToDTO() *ReadingSessionDTO {
	return &ReadingSessionDTO{
		ID:          &m.ID,
		ReadingPlan: m.ReadingPlan.ToDTO(),
		PagesRead:   &m.PagesRead,
		Minutes:     &m.Minutes,
		Notes:       &m.Notes,
		Date:        &m.Date,
	}
}

func (dto ReadingPlanDTO) ToModel() *ReadingPlan {
	var model ReadingPlan

	if dto.ID != nil {
		model.ID = *dto.ID
	}

	if dto.Status != nil {
		model.Status = *dto.Status
	}

	model.StartDate = dto.StartDate
	model.TargetDate = dto.TargetDate

	if dto.Priority != nil {
		model.Priority = *dto.Priority
	}

	if dto.PagesPerDay != nil {
		model.PagesPerDay = *dto.PagesPerDay
	}

	if dto.MinutesPerDay != nil {
		model.MinutesPerDay = *dto.MinutesPerDay
	}

	if dto.Book != nil {
		model.Book = dto.Book.ToModel()
	}

	if dto.User != nil {
		model.User = dto.User.ToModel()
	}

	return &model
}

func (m ReadingPlan) ToDTO() *ReadingPlanDTO {
	return &ReadingPlanDTO{
		ID:            &m.ID,
		Status:        &m.Status,
		StartDate:     m.StartDate,
		TargetDate:    m.TargetDate,
		Priority:      &m.Priority,
		PagesPerDay:   &m.PagesPerDay,
		MinutesPerDay: &m.MinutesPerDay,
		Book:          m.Book.ToDTO(),
		User:          m.User.ToDTO(),
	}
}

func (m *ReadingSession) ValidateReadingSession(v *validator.Validator) {
	v.Check(m.ReadingPlan.ID != 0, "ReadingPlan", "must be provided")

	if m.PagesRead == 0 && m.Minutes == 0 {
		v.Check(false, "Session", "either pagesRead or minutes must be provided")
	}

	v.Check(!m.Date.IsZero(), "Date", "must be provided")
}

func (m *ReadingPlan) ValidateReadingPlan(v *validator.Validator) {
	v.Check(m.Status > 0, "Status", "must be provided")
	v.Check(m.Priority > 0, "Priority", "must be provided")

	v.Check(m.Book.ID != 0, "Book", "must be provided")
	v.Check(m.User.ID != 0, "User", "must be provided")

	if m.PagesPerDay == 0 && m.MinutesPerDay == 0 {
		v.Check(false, "Plan", "either pagesPerDay or minutesPerDay must be provided")
	}

	if m.StartDate != nil && m.TargetDate != nil {
		v.Check(
			m.TargetDate.After(*m.StartDate),
			"TargetDate",
			"must be after startDate",
		)
	}
}

func (s ReadingStatus) String() string {
	switch s {
	case ReadingStatusPlanned:
		return "PLANNED"
	case ReadingStatusReading:
		return "READING"
	case ReadingStatusCompleted:
		return "COMPLETED"
	case ReadingStatusPaused:
		return "PAUSED"
	default:
		return "UNKNOWN"
	}
}

func ReadingStatusFromString(str string) ReadingStatus {
	switch str {
	case "PLANNED":
		return ReadingStatusPlanned
	case "READING":
		return ReadingStatusReading
	case "COMPLETED":
		return ReadingStatusCompleted
	case "PAUSED":
		return ReadingStatusPaused
	default:
		return 0
	}
}
