package models

import "bookwise/utils/validator"

type Book struct {
	ID          int64  `db:"id" dto:"ID"`
	Title       string `db:"title" dto:"Title" `
	Author      string `db:"author" dto:"Author"`
	Pages       int    `db:"pages" dto:"Pages"`
	Description string `db:"description" dto:"Description"`
	BaseModel
	User *User `db:"-" dto:"User"`
}

type BookDTO struct {
	ID          *int64   `json:"id" dto:"ID"`
	Title       *string  `json:"title" dto:"Title"`
	Author      *string  `json:"author" dto:"Author"`
	Pages       *int     `json:"pages" dto:"Pages"`
	Description *string  `json:"description" dto:"Description"`
	User        *UserDTO `json:"user" dto:"User"`
}

func (m BookDTO) ToModel() *Book {
	var model Book

	if m.ID != nil {
		model.ID = *m.ID
	}

	if m.Title != nil {
		model.Title = *m.Title
	}

	if m.Author != nil {
		model.Author = *m.Author
	}

	if m.Pages != nil {
		model.Pages = *m.Pages
	}

	if m.Description != nil {
		model.Description = *m.Description
	}

	if m.User != nil {
		model.User = m.User.ToModel()
	}

	return &model
}

func (m Book) ToDTO() *BookDTO {
	return &BookDTO{
		ID:          &m.ID,
		Title:       &m.Title,
		Author:      &m.Author,
		Pages:       &m.Pages,
		Description: &m.Description,
		User:        m.User.ToDTO(),
	}
}

func (m *Book) ValidateBook(v *validator.Validator) {
	v.Check(m.Title != "", "Title", "must be provided")
	v.Check(m.Author != "", "Author", "must be provided")
	v.Check(m.Pages != 0, "Pages", "must be provided")
	v.Check(m.Description != "", "Description", "must be provided")
}
