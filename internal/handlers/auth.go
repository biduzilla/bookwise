package handlers

import (
	"bookwise/internal/services"
	"bookwise/utils"
	"bookwise/utils/errors"
	"bookwise/utils/validator"
	"net/http"
)

type AuthHandler struct {
	auth          services.AuthServiceInterface
	errorResponse errors.ErrorResponseInterface
}

type AuthHandlerInterface interface {
	LoginHandler(w http.ResponseWriter, r *http.Request)
}

func NewAuthHandler(authService services.AuthServiceInterface, errResp errors.ErrorResponseInterface) *AuthHandler {
	return &AuthHandler{
		auth:          authService,
		errorResponse: errResp,
	}
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		h.errorResponse.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	token, err := h.auth.Login(v, input.Email, input.Password)
	if err != nil {
		h.errorResponse.HandlerErrorResponse(w, r, err, v)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"authentication_token": token}, nil)
	if err != nil {
		h.errorResponse.ServerErrorResponse(w, r, err)
	}
}
