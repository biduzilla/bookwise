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

type readingPlanHandler struct {
	readingPlan services.ReadingPlanService
	errRsp      e.ErrorResponseInterface
	GenericHandlerInterface[models.ReadingPlan, models.ReadingPlanDTO]
}

func NewReadingPlanHandler(
	readingPlan services.ReadingPlanService,
	errRsp e.ErrorResponseInterface,
) *readingPlanHandler {
	return &readingPlanHandler{
		readingPlan:             readingPlan,
		errRsp:                  errRsp,
		GenericHandlerInterface: NewGenericHandler(readingPlan, errRsp),
	}
}

type ReadingPlanHandler interface {
	FindAll(w http.ResponseWriter, r *http.Request)
	GenericHandlerInterface[models.ReadingPlan, models.ReadingPlanDTO]
}

func (h *readingPlanHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	bookID, ok := parseID(w, r, h.errRsp)
	if !ok {
		return
	}

	var input struct {
		status     models.ReadingStatus
		startDate  *time.Time
		targetDate *time.Time
		filters.Filters
	}

	v := validator.New()

	qs := r.URL.Query()
	input.status = models.ReadingStatusFromString(utils.ReadString(qs, "status", ""))
	input.startDate = utils.ReadDate(qs, "start_date", "2006-01-02")
	input.targetDate = utils.ReadDate(qs, "target_date", "2006-01-02")
	input.Filters.Page = utils.ReadInt(qs, "page", 1, v)
	input.Filters.PageSize = utils.ReadInt(qs, "page_size", 20, v)
	input.Filters.Sort = utils.ReadString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "description", "-id", "-description"}

	if filters.ValidateFilters(v, input.Filters); !v.Valid() {
		h.errRsp.FailedValidationResponse(w, r, v.Errors)
		return
	}

	user := contexts.ContextGetUser(r)

	objects, m, err := h.readingPlan.FindAll(
		&input.status,
		input.startDate,
		input.targetDate,
		user.ID,
		bookID,
		input.Filters,
	)

	if err != nil {
		h.errRsp.HandlerErrorResponse(w, r, err, v)
		return
	}

	dtos := make([]*models.ReadingPlanDTO, 0, len(objects))

	for _, o := range objects {
		dtos = append(dtos, o.ToDTO())
	}

	respond(w, r, http.StatusOK, utils.Envelope{"reading_plans": dtos, "metadata": m}, nil, h.errRsp)
}
