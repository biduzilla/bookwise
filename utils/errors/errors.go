package errors

import (
	"bookwise/internal/jsonlog"
	"bookwise/utils"
	"bookwise/utils/validator"
	"errors"
	"fmt"
	"net/http"
)

type ValidationFieldError struct {
	Field   string
	Message string
}

func (e ValidationFieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

var (
	ErrRecordNotFound        = errors.New("record not found")
	ErrEditConflict          = errors.New("edit conflict")
	ErrInvalidData           = errors.New("invalid data")
	ErrInvalidCredentials    = errors.New("invalid authentication credentials")
	ErrInactiveAccount       = errors.New("your user account must be activated to access this resource")
	ErrStartDateAfterEndDate = errors.New("start date must be before end date")
	ErrInvalidRole           = errors.New("invalid role")
	ErrScanModel             = errors.New("dest must be a pointer")

	ErrDuplicateEmail = ValidationFieldError{"email", "a register with this email address already exists"}
	ErrDuplicateName  = ValidationFieldError{"name", "a register with this name already exists"}
	ErrDuplicatePhone = ValidationFieldError{"phone", "a register with this phone number already exists"}
	ErrBookPages      = ValidationFieldError{"pages", "pages must be a positive number"}
	ErrBookTitle      = ValidationFieldError{"title", "book with this title already exists for this user"}
)

type errorResponse struct {
	logger jsonlog.Logger
}

type ErrorResponseInterface interface {
	NotPermittedResponse(w http.ResponseWriter, r *http.Request)
	AuthenticationRequiredResponse(w http.ResponseWriter, r *http.Request)
	InactiveAccountResponse(w http.ResponseWriter, r *http.Request)
	InvalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request)
	InvalidCredentialsResponse(w http.ResponseWriter, r *http.Request)
	InvalidRoleResponse(w http.ResponseWriter, r *http.Request)
	RateLimitExceededResponse(w http.ResponseWriter, r *http.Request)
	ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error)
	NotFoundResponse(w http.ResponseWriter, r *http.Request)
	MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request)
	BadRequestResponse(w http.ResponseWriter, r *http.Request, err error)
	FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string)
	EditConflictResponse(w http.ResponseWriter, r *http.Request)
	HandlerErrorResponse(w http.ResponseWriter, r *http.Request, err error, v *validator.Validator)
}

func NewErrorResponse(logger jsonlog.Logger) *errorResponse {
	return &errorResponse{logger: logger}
}

func (e *errorResponse) HandlerErrorResponse(w http.ResponseWriter, r *http.Request, err error, v *validator.Validator) {
	var dupErr ValidationFieldError
	if v != nil && errors.As(err, &dupErr) {
		v.AddError(dupErr.Field, dupErr.Message)
		e.FailedValidationResponse(w, r, v.Errors)
		return
	}

	switch {
	case errors.Is(err, ErrInvalidData):
		e.FailedValidationResponse(w, r, v.Errors)

	case errors.Is(err, ErrRecordNotFound):
		e.NotFoundResponse(w, r)

	case errors.Is(err, ErrEditConflict):
		e.EditConflictResponse(w, r)

	case errors.Is(err, ErrInactiveAccount):
		e.InactiveAccountResponse(w, r)

	case errors.Is(err, ErrInvalidRole):
		e.InvalidRoleResponse(w, r)

	default:
		e.ServerErrorResponse(w, r, err)
	}
}

func (e *errorResponse) NotPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	e.errorResponse(w, r, http.StatusForbidden, message)
}

func (e *errorResponse) AuthenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	e.errorResponse(w, r, http.StatusUnauthorized, message)
}
func (e *errorResponse) InactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	e.errorResponse(w, r, http.StatusForbidden, message)
}

func (e *errorResponse) InvalidRoleResponse(w http.ResponseWriter, r *http.Request) {
	message := "Your user account does not have access to this feature."
	e.errorResponse(w, r, http.StatusForbidden, message)
}

func (e *errorResponse) InvalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	message := "invalid or missing authentication token"
	e.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (e *errorResponse) InvalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	e.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (e *errorResponse) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceed"
	e.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (e *errorResponse) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logError(r, err)
	message := "the server encountered a problem and could not process your request"
	e.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (e *errorResponse) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	e.errorResponse(w, r, http.StatusNotFound, message)
}

func (e *errorResponse) MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	e.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (e *errorResponse) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (e *errorResponse) FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	e.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (e *errorResponse) EditConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	e.errorResponse(w, r, http.StatusConflict, message)
}

func (e *errorResponse) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := utils.Envelope{"error": message}
	err := utils.WriteJSON(w, status, env, nil)
	if err != nil {
		e.logError(r, err)
		w.WriteHeader(500)
	}
}

func (e *errorResponse) logError(r *http.Request, err error) {
	e.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}
