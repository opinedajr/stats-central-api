package healthcheck

import (
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service}
}

func (h *Handler) Handle(c *gin.Context) {
	health := h.service.Check()
	c.JSON(200, health)
}
