package routers

import (
	"bookwise/internal/handlers"
	"bookwise/internal/middleware"

	"github.com/go-chi/chi"
)

type readingSessionRouter struct {
	readingSession handlers.ReadingSessionHandler
	m              middleware.MiddlewareInterface
}

type ReadingSessionRouter interface {
	ReadingSessionRoutes(r chi.Router)
}

func NewReadingSessionRouter(
	readingSession handlers.ReadingSessionHandler,
	m middleware.MiddlewareInterface,

) *readingSessionRouter {
	return &readingSessionRouter{
		readingSession: readingSession,
		m:              m,
	}
}

func (b *readingSessionRouter) ReadingSessionRoutes(r chi.Router) {
	r.Route("/reading-session", func(r chi.Router) {
		r.Use(b.m.RequireActivatedUser)

		r.Get("/{id}", b.readingSession.FindByID)
		r.Get("/", b.readingSession.FindAll)
		r.Post("/", b.readingSession.Save)
		r.Put("/", b.readingSession.Update)
		r.Delete("/{id}", b.readingSession.Delete)
	})
}
