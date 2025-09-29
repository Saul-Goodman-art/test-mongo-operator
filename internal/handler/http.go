package handler

import (
	"net/http"
	"order-service/internal/models"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
)

type HTTPHandler struct {
	orderService service.OrderService
}

func NewHTTPHandler(orderService service.OrderService) *HTTPHandler {
	return &HTTPHandler{
		orderService: orderService,
	}
}

func (h *HTTPHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *HTTPHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *HTTPHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.POST("/orders", h.CreateOrder)
		api.GET("/orders/:id", h.GetOrder)
	}
}
