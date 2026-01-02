package routers

import (
	"bookwise/internal/handlers"
	"bookwise/internal/middleware"

	"github.com/go-chi/chi"
)

type readingPlanRouter struct {
	readingPlan handlers.ReadingPlanHandler
	m           middleware.MiddlewareInterface
}

type ReadingPlanRouter interface {
	ReadingPlanRoutes(r chi.Router)
}

func NewReadingPlanRouter(
	readingPlan handlers.ReadingPlanHandler,
	m middleware.MiddlewareInterface,

) *readingPlanRouter {
	return &readingPlanRouter{
		readingPlan: readingPlan,
		m:           m,
	}
}

func (b *readingPlanRouter) ReadingPlanRoutes(r chi.Router) {
	r.Route("/reading-plan", func(r chi.Router) {
		r.Use(b.m.RequireActivatedUser)

		r.Get("/{id}", b.readingPlan.FindByID)
		r.Get("/", b.readingPlan.FindAll)
		r.Post("/", b.readingPlan.Save)
		r.Put("/", b.readingPlan.Update)
		r.Delete("/{id}", b.readingPlan.Delete)
	})
}
