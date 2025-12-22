package routers

import (
	"bookwise/internal/handlers"
	"bookwise/internal/middleware"

	"github.com/go-chi/chi"
)

type bookRouter struct {
	book handlers.BookHandler
	m    middleware.MiddlewareInterface
}

type BookRouter interface {
	BookRoutes(r chi.Router)
}

func NewBookRouter(
	book handlers.BookHandler,
	m middleware.MiddlewareInterface,

) *bookRouter {
	return &bookRouter{
		book: book,
		m:    m,
	}
}

func (b *bookRouter) BookRoutes(r chi.Router) {
	r.Route("/books", func(r chi.Router) {
		r.Use(b.m.RequireActivatedUser)

		r.Get("/{id}", b.book.FindByID)
		r.Get("/", b.book.FindAll)
		r.Post("/", b.book.Save)
		r.Put("/", b.book.Update)
		r.Delete("/{id}", b.book.Delete)
	})
}
