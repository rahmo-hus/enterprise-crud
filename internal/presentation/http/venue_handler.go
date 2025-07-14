package http

import (
	"net/http"
	"time"

	"enterprise-crud/internal/domain/venue"
	venueDto "enterprise-crud/internal/dto/venue"
	"enterprise-crud/internal/infrastructure/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// VenueHandler handles HTTP requests for venue operations
type VenueHandler struct {
	venueService venue.Service
	jwtService   *auth.JWTService
}

// NewVenueHandler creates a new instance of VenueHandler
func NewVenueHandler(venueService venue.Service, jwtService *auth.JWTService) *VenueHandler {
	return &VenueHandler{
		venueService: venueService,
		jwtService:   jwtService,
	}
}

// CreateVenue creates a new venue
// @Summary Create a new venue
// @Description Create a new venue (requires ORGANIZER or ADMIN role)
// @Tags venues
// @Accept json
// @Produce json
// @Param venue body venueDto.CreateVenueRequest true "Venue data"
// @Success 201 {object} venueDto.VenueResponse
// @Failure 400 {object} venueDto.ErrorResponse
// @Failure 401 {object} venueDto.ErrorResponse
// @Failure 403 {object} venueDto.ErrorResponse
// @Failure 500 {object} venueDto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/venues [post]
func (h *VenueHandler) CreateVenue(c *gin.Context) {
	var req venueDto.CreateVenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, venueDto.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid input data: " + err.Error(),
		})
		return
	}

	// Create venue entity
	newVenue := &venue.Venue{
		ID:          uuid.New(),
		Name:        req.Name,
		Address:     req.Address,
		Capacity:    req.Capacity,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create the venue
	if err := h.venueService.CreateVenue(c.Request.Context(), newVenue); err != nil {
		if venue.IsVenueError(err) {
			c.JSON(http.StatusBadRequest, venueDto.ErrorResponse{
				Error:   venue.GetVenueErrorCode(err),
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, venueDto.ErrorResponse{
				Error:   "creation_error",
				Message: "Failed to create venue: " + err.Error(),
			})
		}
		return
	}

	// Return created venue
	response := mapVenueToResponse(newVenue)
	c.JSON(http.StatusCreated, response)
}

// GetVenue retrieves a venue by ID
// @Summary Get venue by ID
// @Description Get venue details by ID
// @Tags venues
// @Accept json
// @Produce json
// @Param id path string true "Venue ID"
// @Success 200 {object} venueDto.VenueResponse
// @Failure 400 {object} venueDto.ErrorResponse
// @Failure 404 {object} venueDto.ErrorResponse
// @Failure 500 {object} venueDto.ErrorResponse
// @Router /api/v1/venues/{id} [get]
func (h *VenueHandler) GetVenue(c *gin.Context) {
	venueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, venueDto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid venue ID format",
		})
		return
	}

	foundVenue, err := h.venueService.GetVenueByID(c.Request.Context(), venueID)
	if err != nil {
		if venue.IsVenueNotFoundError(err) {
			c.JSON(http.StatusNotFound, venueDto.ErrorResponse{
				Error:   venue.GetVenueErrorCode(err),
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, venueDto.ErrorResponse{
				Error:   "retrieval_error",
				Message: "Failed to retrieve venue: " + err.Error(),
			})
		}
		return
	}

	response := mapVenueToResponse(foundVenue)
	c.JSON(http.StatusOK, response)
}

// GetAllVenues retrieves all venues
// @Summary Get all venues
// @Description Get list of all venues
// @Tags venues
// @Accept json
// @Produce json
// @Success 200 {object} venueDto.VenueListResponse
// @Failure 500 {object} venueDto.ErrorResponse
// @Router /api/v1/venues [get]
func (h *VenueHandler) GetAllVenues(c *gin.Context) {
	venues, err := h.venueService.GetAllVenues(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, venueDto.ErrorResponse{
			Error:   "retrieval_error",
			Message: "Failed to retrieve venues: " + err.Error(),
		})
		return
	}

	response := venueDto.VenueListResponse{
		Venues: make([]venueDto.VenueResponse, len(venues)),
		Count:  len(venues),
	}

	for i, v := range venues {
		response.Venues[i] = mapVenueToResponse(v)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateVenue updates an existing venue
// @Summary Update venue
// @Description Update an existing venue (requires ORGANIZER or ADMIN role)
// @Tags venues
// @Accept json
// @Produce json
// @Param id path string true "Venue ID"
// @Param venue body venueDto.UpdateVenueRequest true "Venue data"
// @Success 200 {object} venueDto.VenueResponse
// @Failure 400 {object} venueDto.ErrorResponse
// @Failure 401 {object} venueDto.ErrorResponse
// @Failure 403 {object} venueDto.ErrorResponse
// @Failure 404 {object} venueDto.ErrorResponse
// @Failure 500 {object} venueDto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/venues/{id} [put]
func (h *VenueHandler) UpdateVenue(c *gin.Context) {
	venueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, venueDto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid venue ID format",
		})
		return
	}

	var req venueDto.UpdateVenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, venueDto.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid input data: " + err.Error(),
		})
		return
	}

	// Update venue entity
	updatedVenue := &venue.Venue{
		ID:          venueID,
		Name:        req.Name,
		Address:     req.Address,
		Capacity:    req.Capacity,
		Description: req.Description,
		UpdatedAt:   time.Now(),
	}

	// Update the venue
	if err := h.venueService.UpdateVenue(c.Request.Context(), updatedVenue); err != nil {
		if venue.IsVenueNotFoundError(err) {
			c.JSON(http.StatusNotFound, venueDto.ErrorResponse{
				Error:   venue.GetVenueErrorCode(err),
				Message: err.Error(),
			})
		} else if venue.IsVenueError(err) {
			c.JSON(http.StatusBadRequest, venueDto.ErrorResponse{
				Error:   venue.GetVenueErrorCode(err),
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, venueDto.ErrorResponse{
				Error:   "update_error",
				Message: "Failed to update venue: " + err.Error(),
			})
		}
		return
	}

	// Return updated venue
	response := mapVenueToResponse(updatedVenue)
	c.JSON(http.StatusOK, response)
}

// DeleteVenue deletes a venue
// @Summary Delete venue
// @Description Delete a venue (requires ADMIN role)
// @Tags venues
// @Accept json
// @Produce json
// @Param id path string true "Venue ID"
// @Success 200 {object} venueDto.SuccessResponse
// @Failure 400 {object} venueDto.ErrorResponse
// @Failure 401 {object} venueDto.ErrorResponse
// @Failure 403 {object} venueDto.ErrorResponse
// @Failure 404 {object} venueDto.ErrorResponse
// @Failure 500 {object} venueDto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/venues/{id} [delete]
func (h *VenueHandler) DeleteVenue(c *gin.Context) {
	venueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, venueDto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid venue ID format",
		})
		return
	}

	// Delete the venue
	if err := h.venueService.DeleteVenue(c.Request.Context(), venueID); err != nil {
		if venue.IsVenueNotFoundError(err) {
			c.JSON(http.StatusNotFound, venueDto.ErrorResponse{
				Error:   venue.GetVenueErrorCode(err),
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, venueDto.ErrorResponse{
				Error:   "deletion_error",
				Message: "Failed to delete venue: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, venueDto.SuccessResponse{
		Message: "Venue deleted successfully",
	})
}

// RegisterRoutes registers venue routes with the gin router
func (h *VenueHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Create JWT middleware
	jwtMiddleware := auth.NewJWTMiddleware(h.jwtService)

	// Venue routes group
	venueRoutes := router.Group("/venues")
	{
		// Public routes
		venueRoutes.GET("", h.GetAllVenues) // Get all venues
		venueRoutes.GET("/:id", h.GetVenue) // Get venue by ID

		// Organizer routes (require ORGANIZER or ADMIN role)
		venueRoutes.POST("",
			jwtMiddleware.AuthRequired(),
			auth.RequireOrganizer(),
			h.CreateVenue)

		venueRoutes.PUT("/:id",
			jwtMiddleware.AuthRequired(),
			auth.RequireOrganizer(),
			h.UpdateVenue)

		// Admin routes (require ADMIN role)
		venueRoutes.DELETE("/:id",
			jwtMiddleware.AuthRequired(),
			auth.RequireAdmin(),
			h.DeleteVenue)
	}
}

// mapVenueToResponse converts venue entity to response DTO
func mapVenueToResponse(v *venue.Venue) venueDto.VenueResponse {
	return venueDto.VenueResponse{
		ID:          v.ID,
		Name:        v.Name,
		Address:     v.Address,
		Capacity:    v.Capacity,
		Description: v.Description,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}
