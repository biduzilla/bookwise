package routers

import (
	"bookwise/internal/handlers"

	"github.com/go-chi/chi"
)

type UserRouter struct {
	User handlers.UserHandlerInterface
}

func NewUserRouter(userHandler handlers.UserHandlerInterface) *UserRouter {
	return &UserRouter{
		User: userHandler,
	}
}

type UserRoutesInterface interface {
	UserRoutes(r chi.Router)
}

func (u *UserRouter) UserRoutes(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/activate", u.User.ActivateUserHandler)
		r.Post("/", u.User.CreateUserHandler)
	})
}
