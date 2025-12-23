package repositories

import (
	"bookwise/internal/jsonlog"
	"database/sql"
)

type readingPlanRepository struct {
	db     *sql.DB
	logger jsonlog.Logger
}
