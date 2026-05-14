package teams

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/stats-central-api/internal/shared/api"
	"github.com/opinedajr/stats-central-api/internal/shared/logger"
)

var errInvalidID = errors.New("invalid id")

type TeamHandler struct {
	service Service
	logger  logger.Logger
}

func NewTeamHandler(service Service, logger logger.Logger) *TeamHandler {
	return &TeamHandler{
		service: service,
		logger:  logger,
	}
}

func (h *TeamHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var filter TeamFilter

	if country := c.Query("country"); country != "" {
		trimmed := strings.TrimSpace(country)
		if trimmed != "" {
			filter.Country = &trimmed
		}
	}

	if name := c.Query("name"); name != "" {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			filter.Name = &trimmed
		}
	}

	outputs, total, err := h.service.ListTeams(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 {
		totalPages = 1
	}

	c.JSON(http.StatusOK, api.Response[[]*TeamOutput]{
		Data: outputs,
		Meta: &api.PaginationMeta{
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

func (h *TeamHandler) Get(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		h.handleError(c, errInvalidID)
		return
	}

	output, err := h.service.GetTeamByID(c.Request.Context(), uint(teamID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.Response[*TeamOutput]{
		Data: output,
	})
}

func (h *TeamHandler) handleError(c *gin.Context, err error) {
	ctx := c.Request.Context()
	switch {
	case errors.Is(err, errInvalidID):
		c.JSON(http.StatusBadRequest, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "INVALID_ID",
				Message: "invalid id",
			},
		})
	case errors.Is(err, ErrTeamNotFound):
		c.JSON(http.StatusNotFound, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "TEAM_NOT_FOUND",
				Message: "team not found",
			},
		})
	case errors.Is(err, ErrValidationFailed):
		c.JSON(http.StatusBadRequest, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "VALIDATION_FAILED",
				Message: err.Error(),
			},
		})
	case errors.Is(err, ErrDatabaseError):
		h.logger.Error(ctx, "database error", "error", err)
		c.JSON(http.StatusInternalServerError, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "DATABASE_ERROR",
				Message: "database error",
			},
		})
	default:
		h.logger.Error(ctx, "unexpected error", "error", err)
		c.JSON(http.StatusInternalServerError, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "An unexpected error occurred",
			},
		})
	}
}

func (h *TeamHandler) MethodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, api.Response[struct{}]{
		Error: &api.APIError{
			Code:    "METHOD_NOT_ALLOWED",
			Message: "method not allowed",
		},
	})
}
