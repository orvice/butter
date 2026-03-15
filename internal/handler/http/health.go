package http

import (
	"net/http"

	"go.orx.me/apps/butter/internal/service"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	service *service.HealthService
}

func NewHealthHandler(service *service.HealthService) *HealthHandler {
	return &HealthHandler{service: service}
}

func (h *HealthHandler) Register(r *gin.Engine) {
	r.GET("/ping", h.Ping)
}

func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.Ping())
}
