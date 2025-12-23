package handlers

import (
	"bookwise/internal/contexts"
	"bookwise/internal/models"
	"bookwise/internal/models/filters"
	"bookwise/internal/services"
	"bookwise/utils"
	e "bookwise/utils/errors"
	"bookwise/utils/validator"
	"net/http"
)

type bookHandler struct {
	book   services.BookService
	errRsp e.ErrorResponseInterface
	GenericHandlerInterface[models.Book, models.BookDTO]
}

func NewBookHandler(
	book services.BookService,
	errRsp e.ErrorResponseInterface,
) *bookHandler {
	return &bookHandler{
		book:                    book,
		errRsp:                  errRsp,
		GenericHandlerInterface: NewGenericHandler(book, errRsp),
	}
}

type BookHandler interface {
	FindAll(w http.ResponseWriter, r *http.Request)
	GenericHandlerInterface[
		models.Book,
		models.BookDTO,
	]
}

func (h *bookHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	var input struct {
		title,
		author string
		filters.Filters
	}

	v := validator.New()

	qs := r.URL.Query()
	input.title = utils.ReadString(qs, "title", "")
	input.author = utils.ReadString(qs, "author", "")
	input.Filters.Page = utils.ReadInt(qs, "page", 1, v)
	input.Filters.PageSize = utils.ReadInt(qs, "page_size", 20, v)
	input.Filters.Sort = utils.ReadString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "-id", "-name"}

	if filters.ValidateFilters(v, input.Filters); !v.Valid() {
		h.errRsp.HandlerErrorResponse(w, r, e.ErrInvalidData, v)
		return
	}

	user := contexts.ContextGetUser(r)
	books, metadata, err := h.book.FindAll(
		input.title,
		input.author,
		user.ID,
		input.Filters,
	)

	if err != nil {
		h.errRsp.HandlerErrorResponse(w, r, err, v)
		return
	}

	dtos := make([]*models.BookDTO, 0, len(books))

	for _, book := range books {
		dtos = append(dtos, book.ToDTO())
	}

	respond(w, r, http.StatusOK, utils.Envelope{"books": dtos, "metadata": metadata}, nil, h.errRsp)
}
