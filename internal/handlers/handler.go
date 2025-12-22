package handlers

import (
	"bookwise/internal/config"
	"bookwise/internal/jsonlog"
	"bookwise/internal/services"
	"bookwise/utils"
	"bookwise/utils/errors"
	"database/sql"
	"net/http"
)

type Handler struct {
	User    UserHandlerInterface
	Auth    AuthHandlerInterface
	Book    BookHandler
	Service *services.Services
}

func NewHandler(
	db *sql.DB,
	errRsp errors.ErrorResponseInterface,
	config config.Config,
	logger jsonlog.Logger,
) *Handler {
	s := services.NewServices(logger, db, config)

	return &Handler{
		Service: s,
		User:    NewUserHandler(s.User, errRsp),
		Auth:    NewAuthHandler(s.Auth, errRsp),
		Book:    NewBookHandler(s.Book, errRsp),
	}
}

func parseID(
	w http.ResponseWriter,
	r *http.Request,
	errRsp errors.ErrorResponseInterface,
) (int64, bool) {
	id, err := utils.ReadIntPathVariable(r, "id")
	if err != nil {
		errRsp.BadRequestResponse(w, r, err)
		return 0, false
	}
	return id, true
}

func respond(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	data utils.Envelope,
	headers http.Header,
	errRsp errors.ErrorResponseInterface,
) {
	err := utils.WriteJSON(w, status, data, headers)
	if err != nil {
		errRsp.ServerErrorResponse(w, r, err)
	}
}
