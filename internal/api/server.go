package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"mini-fhir/internal/fhir/dstu3"
	"mini-fhir/internal/search"
	"mini-fhir/internal/store"
	"mini-fhir/internal/validation"
)

type Server struct {
	Registry  *dstu3.Registry
	Validator *validation.Validator
	Store     *store.Store
	Searcher  *search.Searcher
}

func RegisterRoutes(e *echo.Echo, registry *dstu3.Registry, validator *validation.Validator, store *store.Store, searcher *search.Searcher) {
	s := &Server{
		Registry:  registry,
		Validator: validator,
		Store:     store,
		Searcher:  searcher,
	}

	e.GET("/metadata", s.handleMetadata)
	e.POST("/$validate", s.handleValidate)
	e.POST("/:type/$validate", s.handleValidate)

	e.POST("/", s.handleBatchTransaction)
	e.POST("/:type", s.handleCreate)
	e.GET("/:type/:id", s.handleRead)
	e.PUT("/:type/:id", s.handleUpdate)
	e.DELETE("/:type/:id", s.handleDelete)
	e.GET("/:type/:id/_history", s.handleHistory)
	e.GET("/_history", s.handleSystemHistory)
	e.GET("/:type", s.handleSearch)

	e.GET("/healthz", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
}
