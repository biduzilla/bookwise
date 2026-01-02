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
	"time"
)

type readingSessionHandler struct {
	readingSession services.ReadingSessionService
	errRsp         e.ErrorResponseInterface
	GenericHandlerInterface[models.ReadingPlan, models.ReadingPlanDTO]
}

func NewReadingSessionHandler(
	readingSession services.ReadingSessionService,
	errRsp e.ErrorResponseInterface,
) *readingSessionHandler {
	return &readingSessionHandler{
		readingSession:          readingSession,
		errRsp:                  errRsp,
		GenericHandlerInterface: NewGenericHandler(readingSession, errRsp),
	}
}

type ReadingSessionHandler interface {
	FindAll(w http.ResponseWriter, r *http.Request)
	GenericHandlerInterface[models.ReadingSession, models.ReadingSessionDTO]
}

func (h *readingSessionHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	readingPlanID, ok := parseID(w, r, h.errRsp)
	if !ok {
		return
	}

	var input struct {
		date *time.Time
		filters.Filters
	}

	v := validator.New()

	qs := r.URL.Query()
	input.date = utils.ReadDate(qs, "date", "2006-01-02")
	input.Filters.Page = utils.ReadInt(qs, "page", 1, v)
	input.Filters.PageSize = utils.ReadInt(qs, "page_size", 20, v)
	input.Filters.Sort = utils.ReadString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "description", "-id", "-description"}

	if filters.ValidateFilters(v, input.Filters); !v.Valid() {
		h.errRsp.FailedValidationResponse(w, r, v.Errors)
		return
	}

	user := contexts.ContextGetUser(r)

	objects, m, err := h.readingSession.FindAll(
		input.date,
		user.ID,
		readingPlanID,
		input.Filters,
	)

	if err != nil {
		h.errRsp.HandlerErrorResponse(w, r, err, v)
		return
	}

	dtos := make([]*models.ReadingSessionDTO, 0, len(objects))

	for _, o := range objects {
		dtos = append(dtos, o.ToDTO())
	}

	respond(w, r, http.StatusOK, utils.Envelope{"reading_sessions": dtos, "metadata": m}, nil, h.errRsp)
}
