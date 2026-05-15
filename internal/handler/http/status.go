package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/service"
)

type StatusHandler struct {
	service *service.StatusService
}

func NewStatusHandler(service *service.StatusService) *StatusHandler {
	return &StatusHandler{service: service}
}

func (h *StatusHandler) Register(r *gin.Engine) {
	r.GET("/status", h.Status)
}

func (h *StatusHandler) Status(c *gin.Context) {
	status, err := h.service.Status(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}
