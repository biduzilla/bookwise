package models

import "time"

type BaseModel struct {
	Version   int        `db:"version"`
	Deleted   bool       `db:"deleted"`
	CreatedAt time.Time  `db:"created_at"`
	CreatedBy *int64     `db:"created_by"`
	UpdatedAt *time.Time `db:"updated_at"`
	UpdatedBy *int64     `db:"updated_by"`
}
