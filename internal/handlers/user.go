package handlers

import (
	"bookwise/internal/models"
	"bookwise/internal/services"
	"bookwise/utils"
	e "bookwise/utils/errors"
	"bookwise/utils/validator"
	"net/http"
)

type UserHandler struct {
	user   services.UserService
	errRsp e.ErrorResponseInterface
}

type UserHandlerInterface interface {
	ActivateUserHandler(w http.ResponseWriter, r *http.Request)
	CreateUserHandler(w http.ResponseWriter, r *http.Request)
}

func NewUserHandler(
	user services.UserService,
	errRsp e.ErrorResponseInterface,
) *UserHandler {
	return &UserHandler{
		user:   user,
		errRsp: errRsp,
	}
}

func (h *UserHandler) ActivateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Cod   int    `json:"cod"`
		Email string `json:"email"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		h.errRsp.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	user, err := h.user.ActivateUser(
		input.Cod,
		input.Email,
		v,
	)
	if err != nil {
		h.errRsp.HandlerErrorResponse(w, r, err, v)
		return
	}

	dto, err := utils.ConverterByTag[models.UserDTO](user, "dto")

	if err != nil {
		h.errRsp.ServerErrorResponse(w, r, err)
		return
	}

	respond(
		w,
		r,
		http.StatusOK,
		utils.Envelope{"user": dto},
		nil,
		h.errRsp,
	)
}

func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var userDTO models.UserSaveDTO
	if err := utils.ReadJSON(w, r, &userDTO); err != nil {
		h.errRsp.BadRequestResponse(w, r, err)
		return
	}

	user, err := userDTO.ToModel()
	if err != nil {
		h.errRsp.ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	err = h.user.Save(user, v)
	if err != nil {
		h.errRsp.HandlerErrorResponse(w, r, err, v)
		return
	}

	dto, err := utils.ConverterByTag[models.UserDTO](user, "dto")

	if err != nil {
		h.errRsp.ServerErrorResponse(w, r, err)
		return
	}

	respond(
		w,
		r,
		http.StatusCreated,
		utils.Envelope{"user": dto},
		nil,
		h.errRsp,
	)
}
