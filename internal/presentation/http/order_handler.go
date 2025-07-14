package http

import (
	"net/http"

	"enterprise-crud/internal/domain/order"
	orderDto "enterprise-crud/internal/dto/order"
	"enterprise-crud/internal/infrastructure/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OrderHandler handles HTTP requests for order operations
type OrderHandler struct {
	orderService order.Service
	jwtService   *auth.JWTService
}

// NewOrderHandler creates a new instance of OrderHandler
func NewOrderHandler(orderService order.Service, jwtService *auth.JWTService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		jwtService:   jwtService,
	}
}

// CreateOrder creates a new order
// @Summary Create a new order
// @Description Create a new order (requires USER role)
// @Tags orders
// @Accept json
// @Produce json
// @Param order body orderDto.CreateOrderRequest true "Order data"
// @Success 201 {object} orderDto.OrderResponse
// @Failure 400 {object} orderDto.ErrorResponse
// @Failure 401 {object} orderDto.ErrorResponse
// @Failure 403 {object} orderDto.ErrorResponse
// @Failure 500 {object} orderDto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req orderDto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, orderDto.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid input data: " + err.Error(),
		})
		return
	}

	// Get user ID from context
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, orderDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	claims, ok := userClaims.(*auth.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, orderDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid authentication credentials",
		})
		return
	}

	// Create the order
	createdOrder, err := h.orderService.CreateOrder(c.Request.Context(), claims.UserID, req.EventID, req.Quantity)
	if err != nil {
		// Handle different types of errors appropriately
		if order.IsInvalidQuantityError(err) {
			c.JSON(http.StatusBadRequest, orderDto.ErrorResponse{
				Error:   order.GetOrderErrorCode(err),
				Message: err.Error(),
			})
		} else if order.IsEventNotFoundError(err) {
			c.JSON(http.StatusNotFound, orderDto.ErrorResponse{
				Error:   order.GetOrderErrorCode(err),
				Message: err.Error(),
			})
		} else if order.IsEventNotActiveError(err) {
			c.JSON(http.StatusBadRequest, orderDto.ErrorResponse{
				Error:   order.GetOrderErrorCode(err),
				Message: err.Error(),
			})
		} else if order.IsInsufficientTicketsError(err) {
			c.JSON(http.StatusBadRequest, orderDto.ErrorResponse{
				Error:   order.GetOrderErrorCode(err),
				Message: err.Error(),
			})
		} else if order.IsOrderCreationError(err) {
			c.JSON(http.StatusInternalServerError, orderDto.ErrorResponse{
				Error:   order.GetOrderErrorCode(err),
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, orderDto.ErrorResponse{
				Error:   "creation_error",
				Message: "Failed to create order: " + err.Error(),
			})
		}
		return
	}

	// Return created order
	response := mapOrderToResponse(createdOrder)
	c.JSON(http.StatusCreated, response)
}

// GetOrder retrieves an order by ID
// @Summary Get order by ID
// @Description Get order details by ID (user can only see their own orders)
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} orderDto.OrderResponse
// @Failure 400 {object} orderDto.ErrorResponse
// @Failure 401 {object} orderDto.ErrorResponse
// @Failure 403 {object} orderDto.ErrorResponse
// @Failure 404 {object} orderDto.ErrorResponse
// @Failure 500 {object} orderDto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, orderDto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid order ID format",
		})
		return
	}

	// Get user ID from context
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, orderDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	claims, ok := userClaims.(*auth.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, orderDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid authentication credentials",
		})
		return
	}

	foundOrder, err := h.orderService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		if order.IsOrderNotFoundError(err) {
			c.JSON(http.StatusNotFound, orderDto.ErrorResponse{
				Error:   order.GetOrderErrorCode(err),
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, orderDto.ErrorResponse{
				Error:   "retrieval_error",
				Message: "Failed to retrieve order: " + err.Error(),
			})
		}
		return
	}

	// Check if user can access this order (only own orders unless admin)
	if foundOrder.UserID != claims.UserID && !auth.HasRole(c, "ADMIN") {
		c.JSON(http.StatusForbidden, orderDto.ErrorResponse{
			Error:   "forbidden",
			Message: "You can only view your own orders",
		})
		return
	}

	response := mapOrderToResponse(foundOrder)
	c.JSON(http.StatusOK, response)
}

// GetMyOrders retrieves all orders for the current user
// @Summary Get my orders
// @Description Get all orders for the current user
// @Tags orders
// @Accept json
// @Produce json
// @Success 200 {object} orderDto.OrderListResponse
// @Failure 401 {object} orderDto.ErrorResponse
// @Failure 500 {object} orderDto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/orders/my-orders [get]
func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	// Get user ID from context
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, orderDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	claims, ok := userClaims.(*auth.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, orderDto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid authentication credentials",
		})
		return
	}

	orders, err := h.orderService.GetOrdersByUserID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, orderDto.ErrorResponse{
			Error:   "retrieval_error",
			Message: "Failed to retrieve orders: " + err.Error(),
		})
		return
	}

	response := orderDto.OrderListResponse{
		Orders: make([]orderDto.OrderResponse, len(orders)),
		Count:  len(orders),
	}

	for i, o := range orders {
		response.Orders[i] = mapOrderToResponse(o)
	}

	c.JSON(http.StatusOK, response)
}

// RegisterRoutes registers order routes with the gin router
func (h *OrderHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Create JWT middleware
	jwtMiddleware := auth.NewJWTMiddleware(h.jwtService)

	// Order routes group
	orderRoutes := router.Group("/orders")
	{
		// User routes (require USER role)
		orderRoutes.POST("",
			jwtMiddleware.AuthRequired(),
			auth.RequireUser(),
			h.CreateOrder)

		orderRoutes.GET("/:id",
			jwtMiddleware.AuthRequired(),
			auth.RequireUser(),
			h.GetOrder)

		orderRoutes.GET("/my-orders",
			jwtMiddleware.AuthRequired(),
			auth.RequireUser(),
			h.GetMyOrders)
	}
}

// mapOrderToResponse converts order entity to response DTO
func mapOrderToResponse(o *order.Order) orderDto.OrderResponse {
	return orderDto.OrderResponse{
		ID:          o.ID,
		UserID:      o.UserID,
		EventID:     o.EventID,
		Quantity:    o.Quantity,
		TotalAmount: o.TotalAmount,
		Status:      o.Status,
		CreatedAt:   o.CreatedAt,
	}
}