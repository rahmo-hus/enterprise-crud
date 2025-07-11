package http

import (
	"net/http"

	"enterprise-crud/internal/domain/event"
	eventDto "enterprise-crud/internal/dto/event"
	"enterprise-crud/internal/infrastructure/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EventHandler handles HTTP requests for event operations
type EventHandler struct {
	eventService event.Service
	jwtService   *auth.JWTService
}

// NewEventHandler creates a new instance of EventHandler
func NewEventHandler(eventService event.Service, jwtService *auth.JWTService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
		jwtService:   jwtService,
	}
}

// CreateEvent creates a new event
// @Summary Create a new event
// @Description Create a new event (requires ORGANIZER or ADMIN role)
// @Tags events
// @Accept json
// @Produce json
// @Param event body event.CreateEventRequest true "Event data"
// @Success 201 {object} event.EventResponse
// @Failure 400 {object} event.ErrorResponse
// @Failure 401 {object} event.ErrorResponse
// @Failure 403 {object} event.ErrorResponse
// @Failure 500 {object} event.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/events [post]
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req eventDto.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, eventDto.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid input data: " + err.Error(),
		})
		return
	}

	// Get user ID from context
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, eventDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	claims, ok := userClaims.(*auth.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, eventDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid authentication credentials",
		})
		return
	}

	// Create event entity
	newEvent := &event.Event{
		VenueID:      req.VenueID,
		OrganizerID:  claims.UserID,
		Title:        req.Title,
		Description:  req.Description,
		EventDate:    req.EventDate,
		TicketPrice:  req.TicketPrice,
		TotalTickets: req.TotalTickets,
	}

	// Create the event
	if err := h.eventService.CreateEvent(c.Request.Context(), newEvent); err != nil {
		// Handle different types of errors appropriately
		if event.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, eventDto.ErrorResponse{
				Error:   event.GetEventErrorCode(err),
				Message: err.Error(),
			})
		} else if event.IsVenueNotFoundError(err) {
			c.JSON(http.StatusNotFound, eventDto.ErrorResponse{
				Error:   event.GetEventErrorCode(err),
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, eventDto.ErrorResponse{
				Error:   "creation_error",
				Message: "Failed to create event: " + err.Error(),
			})
		}
		return
	}

	// Return created event
	response := mapEventToResponse(newEvent)
	c.JSON(http.StatusCreated, response)
}

// GetEvent retrieves an event by ID
// @Summary Get event by ID
// @Description Get event details by ID
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} event.EventResponse
// @Failure 400 {object} event.ErrorResponse
// @Failure 404 {object} event.ErrorResponse
// @Failure 500 {object} event.ErrorResponse
// @Router /api/v1/events/{id} [get]
func (h *EventHandler) GetEvent(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, eventDto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid event ID format",
		})
		return
	}

	foundEvent, err := h.eventService.GetEventByID(c.Request.Context(), eventID)
	if err != nil {
		if event.IsEventNotFoundError(err) {
			c.JSON(http.StatusNotFound, eventDto.ErrorResponse{
				Error:   event.GetEventErrorCode(err),
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, eventDto.ErrorResponse{
				Error:   "retrieval_error",
				Message: "Failed to retrieve event: " + err.Error(),
			})
		}
		return
	}

	response := mapEventToResponse(foundEvent)
	c.JSON(http.StatusOK, response)
}

// GetAllEvents retrieves all events
// @Summary Get all events
// @Description Get list of all events
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {object} event.EventListResponse
// @Failure 500 {object} event.ErrorResponse
// @Router /api/v1/events [get]
func (h *EventHandler) GetAllEvents(c *gin.Context) {
	events, err := h.eventService.GetAllEvents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, eventDto.ErrorResponse{
			Error:   event.GetEventErrorCode(err),
			Message: err.Error(),
		})
		return
	}

	response := eventDto.EventListResponse{
		Events: make([]eventDto.EventResponse, len(events)),
		Count:  len(events),
	}

	for i, e := range events {
		response.Events[i] = mapEventToResponse(e)
	}

	c.JSON(http.StatusOK, response)
}

// GetMyEvents retrieves events created by the current organizer
// @Summary Get my events
// @Description Get events created by the current organizer
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {object} event.EventListResponse
// @Failure 401 {object} event.ErrorResponse
// @Failure 500 {object} event.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/events/my-events [get]
func (h *EventHandler) GetMyEvents(c *gin.Context) {
	// Get user ID from context
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, eventDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	claims, ok := userClaims.(*auth.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, eventDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid authentication credentials",
		})
		return
	}

	events, err := h.eventService.GetEventsByOrganizer(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, eventDto.ErrorResponse{
			Error:   event.GetEventErrorCode(err),
			Message: err.Error(),
		})
		return
	}

	response := eventDto.EventListResponse{
		Events: make([]eventDto.EventResponse, len(events)),
		Count:  len(events),
	}

	for i, e := range events {
		response.Events[i] = mapEventToResponse(e)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateEvent updates an existing event
// @Summary Update event
// @Description Update an existing event (only by organizer)
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param event body eventDto.UpdateEventRequest true "Event data"
// @Success 200 {object} event.EventResponse
// @Failure 400 {object} event.ErrorResponse
// @Failure 401 {object} event.ErrorResponse
// @Failure 403 {object} event.ErrorResponse
// @Failure 404 {object} event.ErrorResponse
// @Failure 500 {object} event.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/events/{id} [put]
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, eventDto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid event ID format",
		})
		return
	}

	var req eventDto.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, eventDto.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid input data: " + err.Error(),
		})
		return
	}

	// Get user ID from context
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, eventDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	claims, ok := userClaims.(*auth.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, eventDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid authentication credentials",
		})
		return
	}

	// Get existing event to check ownership
	existingEvent, err := h.eventService.GetEventByID(c.Request.Context(), eventID)
	if err != nil {
		if event.IsEventNotFoundError(err) {
			c.JSON(http.StatusNotFound, eventDto.ErrorResponse{
				Error:   event.GetEventErrorCode(err),
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, eventDto.ErrorResponse{
				Error:   "retrieval_error",
				Message: "Failed to retrieve event: " + err.Error(),
			})
		}
		return
	}

	// Check if user is the organizer (unless they're admin)
	if existingEvent.OrganizerID != claims.UserID && !auth.HasRole(c, "ADMIN") {
		c.JSON(http.StatusForbidden, eventDto.ErrorResponse{
			Error:   "forbidden",
			Message: "You can only update your own events",
		})
		return
	}

	// Update event entity
	updatedEvent := &event.Event{
		ID:           eventID,
		VenueID:      req.VenueID,
		OrganizerID:  existingEvent.OrganizerID,
		Title:        req.Title,
		Description:  req.Description,
		EventDate:    req.EventDate,
		TicketPrice:  req.TicketPrice,
		TotalTickets: req.TotalTickets,
		Status:       existingEvent.Status,
		CreatedAt:    existingEvent.CreatedAt,
	}

	// Update the event
	if err := h.eventService.UpdateEvent(c.Request.Context(), updatedEvent); err != nil {
		// Handle different types of errors appropriately
		if event.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, eventDto.ErrorResponse{
				Error:   event.GetEventErrorCode(err),
				Message: err.Error(),
			})
		} else if event.IsVenueNotFoundError(err) {
			c.JSON(http.StatusNotFound, eventDto.ErrorResponse{
				Error:   event.GetEventErrorCode(err),
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, eventDto.ErrorResponse{
				Error:   "update_error",
				Message: "Failed to update event: " + err.Error(),
			})
		}
		return
	}

	// Return updated event
	response := mapEventToResponse(updatedEvent)
	c.JSON(http.StatusOK, response)
}

// CancelEvent cancels an event
// @Summary Cancel event
// @Description Cancel an event (only by organizer)
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} event.SuccessResponse
// @Failure 400 {object} event.ErrorResponse
// @Failure 401 {object} event.ErrorResponse
// @Failure 403 {object} event.ErrorResponse
// @Failure 404 {object} event.ErrorResponse
// @Failure 500 {object} event.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/events/{id}/cancel [patch]
func (h *EventHandler) CancelEvent(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, eventDto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid event ID format",
		})
		return
	}

	// Get user ID from context
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, eventDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	claims, ok := userClaims.(*auth.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, eventDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid authentication credentials",
		})
		return
	}

	// Cancel the event
	if err := h.eventService.CancelEvent(c.Request.Context(), eventID, claims.UserID); err != nil {
		// Handle different types of errors appropriately
		if event.IsEventNotFoundError(err) {
			c.JSON(http.StatusNotFound, eventDto.ErrorResponse{
				Error:   event.GetEventErrorCode(err),
				Message: err.Error(),
			})
		} else if event.IsUnauthorizedError(err) {
			c.JSON(http.StatusForbidden, eventDto.ErrorResponse{
				Error:   event.GetEventErrorCode(err),
				Message: err.Error(),
			})
		} else if event.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, eventDto.ErrorResponse{
				Error:   event.GetEventErrorCode(err),
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, eventDto.ErrorResponse{
				Error:   "cancel_error",
				Message: "Failed to cancel event: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, eventDto.SuccessResponse{
		Message: "Event cancelled successfully",
	})
}

// DeleteEvent deletes an event
// @Summary Delete event
// @Description Delete an event (only by organizer, only if no tickets sold)
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} event.SuccessResponse
// @Failure 400 {object} event.ErrorResponse
// @Failure 401 {object} event.ErrorResponse
// @Failure 403 {object} event.ErrorResponse
// @Failure 404 {object} event.ErrorResponse
// @Failure 500 {object} event.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/events/{id} [delete]
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, eventDto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid event ID format",
		})
		return
	}

	// Get user ID from context
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, eventDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	claims, ok := userClaims.(*auth.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, eventDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid authentication credentials",
		})
		return
	}

	// Delete the event
	if err := h.eventService.DeleteEvent(c.Request.Context(), eventID, claims.UserID); err != nil {
		// Handle different types of errors appropriately
		if event.IsEventNotFoundError(err) {
			c.JSON(http.StatusNotFound, eventDto.ErrorResponse{
				Error:   event.GetEventErrorCode(err),
				Message: err.Error(),
			})
		} else if event.IsUnauthorizedError(err) {
			c.JSON(http.StatusForbidden, eventDto.ErrorResponse{
				Error:   event.GetEventErrorCode(err),
				Message: err.Error(),
			})
		} else if event.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, eventDto.ErrorResponse{
				Error:   event.GetEventErrorCode(err),
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, eventDto.ErrorResponse{
				Error:   "delete_error",
				Message: "Failed to delete event: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, eventDto.SuccessResponse{
		Message: "Event deleted successfully",
	})
}

// RegisterRoutes registers event routes with the gin router
func (h *EventHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Create JWT middleware
	jwtMiddleware := auth.NewJWTMiddleware(h.jwtService)

	// Event routes group
	eventRoutes := router.Group("/events")
	{
		// Public routes
		eventRoutes.GET("", h.GetAllEvents)        // Get all events
		eventRoutes.GET("/:id", h.GetEvent)        // Get event by ID

		// Organizer routes (require ORGANIZER or ADMIN role)
		eventRoutes.POST("",
			jwtMiddleware.AuthRequired(),
			auth.RequireOrganizer(),
			h.CreateEvent)

		eventRoutes.PUT("/:id",
			jwtMiddleware.AuthRequired(),
			auth.RequireOrganizer(),
			h.UpdateEvent)

		eventRoutes.PATCH("/:id/cancel",
			jwtMiddleware.AuthRequired(),
			auth.RequireOrganizer(),
			h.CancelEvent)

		eventRoutes.DELETE("/:id",
			jwtMiddleware.AuthRequired(),
			auth.RequireOrganizer(),
			h.DeleteEvent)

		// My events route (require authentication)
		eventRoutes.GET("/my-events",
			jwtMiddleware.AuthRequired(),
			auth.RequireOrganizer(),
			h.GetMyEvents)
	}
}

// mapEventToResponse converts event entity to response DTO
func mapEventToResponse(e *event.Event) eventDto.EventResponse {
	return eventDto.EventResponse{
		ID:               e.ID,
		VenueID:          e.VenueID,
		OrganizerID:      e.OrganizerID,
		Title:            e.Title,
		Description:      e.Description,
		EventDate:        e.EventDate,
		TicketPrice:      e.TicketPrice,
		AvailableTickets: e.AvailableTickets,
		TotalTickets:     e.TotalTickets,
		Status:           e.Status,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
}