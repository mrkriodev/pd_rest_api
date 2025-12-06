package http

import (
	"context"
	"net/http"
	"strconv"

	"pdrest/internal/interfaces/services"

	"github.com/labstack/echo/v4"
)

type HTTPHandler struct {
	clientService *services.ClientService
	eventService  *services.EventService
}

func NewHTTPHandler(e *echo.Echo, clientService *services.ClientService, eventService *services.EventService) {
	h := &HTTPHandler{
		clientService: clientService,
		eventService:  eventService,
	}

	api := e.Group("/api")
	api.GET("/status", h.Status)
	api.GET("/client_status/:id", h.ClientStatus)
	api.GET("/available_events", h.AvailableEvents)
}

func (h *HTTPHandler) Status(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HTTPHandler) ClientStatus(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	status, err := h.clientService.GetStatus(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, status)
}

func (h *HTTPHandler) AvailableEvents(c echo.Context) error {
	if h.eventService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for events"})
	}

	ctx := context.Background()
	events, err := h.eventService.GetAvailableEvents(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"events": events,
	})
}
