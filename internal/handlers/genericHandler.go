package handlers

import (
	"bookwise/internal/contexts"
	"bookwise/internal/models"
	"bookwise/internal/services"
	"bookwise/utils"
	"bookwise/utils/errors"
	"bookwise/utils/validator"
	"net/http"
)

type genericHandler[
	T models.ModelInterface[D],
	D models.DTOInterface[T],
] struct {
	service services.GenericServiceInterface[T, D]
	errRsp  errors.ErrorResponseInterface
}

type GenericHandlerInterface[
	T models.ModelInterface[D],
	D models.DTOInterface[T],
] interface {
	FindByID(w http.ResponseWriter, r *http.Request)
	Save(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

func NewGenericHandler[
	T models.ModelInterface[D],
	D models.DTOInterface[T],
](
	service services.GenericServiceInterface[T, D],
	errRsp errors.ErrorResponseInterface,
) *genericHandler[T, D] {
	return &genericHandler[T, D]{
		service: service,
		errRsp:  errRsp,
	}
}

func (h *genericHandler[T, D]) FindByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, h.errRsp)
	if !ok {
		return
	}

	user := contexts.ContextGetUser(r)
	model, err := h.service.FindByID(id, user.ID)
	if err != nil {
		h.errRsp.HandlerErrorResponse(w, r, err, nil)
		return
	}

	respond(
		w, r,
		http.StatusOK,
		utils.Envelope{utils.GetTypeName(model): (*model).ToDTO()},
		nil,
		h.errRsp,
	)
}

func (h *genericHandler[T, D]) Save(w http.ResponseWriter, r *http.Request) {
	var dto D
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		h.errRsp.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	user := contexts.ContextGetUser(r)
	model := dto.ToModel()

	if err := h.service.Save(model, user.ID, v); err != nil {
		h.errRsp.HandlerErrorResponse(w, r, err, v)
		return
	}

	respond(w, r, http.StatusCreated, utils.Envelope{utils.GetTypeName(model): (*model).ToDTO()}, nil, h.errRsp)
}

func (h *genericHandler[T, D]) Update(w http.ResponseWriter, r *http.Request) {
	var dto D
	if err := utils.ReadJSON(w, r, &dto); err != nil {
		h.errRsp.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	user := contexts.ContextGetUser(r)
	model := dto.ToModel()

	if err := h.service.Update(model, user.ID, v); err != nil {
		h.errRsp.HandlerErrorResponse(w, r, err, v)
		return
	}

	respond(w, r, http.StatusOK, utils.Envelope{utils.GetTypeName(model): (*model).ToDTO()}, nil, h.errRsp)
}

func (h *genericHandler[T, D]) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, h.errRsp)
	if !ok {
		return
	}

	user := contexts.ContextGetUser(r)
	if err := h.service.Delete(id, user.ID); err != nil {
		h.errRsp.HandlerErrorResponse(w, r, err, nil)
		return
	}

	respond(
		w,
		r,
		http.StatusNoContent,
		nil,
		nil,
		h.errRsp,
	)
}
