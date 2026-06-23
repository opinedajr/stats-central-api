package match

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/stats-central-api/internal/shared/api"
	"github.com/opinedajr/stats-central-api/internal/shared/logger"
)

type MatchHandler struct {
	service Service
	logger  logger.Logger
}

func NewMatchHandler(service Service, logger logger.Logger) *MatchHandler {
	return &MatchHandler{
		service: service,
		logger:  logger,
	}
}

func (h *MatchHandler) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil {
		pageSize = 20
	}

	var filter MatchFilter

	if v := c.Query("tournament_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 32)
		if err == nil {
			uid := uint(id)
			filter.TournamentID = &uid
		}
	}

	if v := c.Query("season"); v != "" {
		filter.Season = &v
	}

	if v := c.Query("round"); v != "" {
		round, err := strconv.Atoi(v)
		if err == nil {
			filter.Round = &round
		}
	}

	if v := c.Query("status"); v != "" {
		filter.Status = &v
	}

	if v := c.Query("home_team_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 32)
		if err == nil {
			uid := uint(id)
			filter.HomeTeamID = &uid
		}
	}

	if v := c.Query("away_team_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 32)
		if err == nil {
			uid := uint(id)
			filter.AwayTeamID = &uid
		}
	}

	outputs, total, err := h.service.ListMatches(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 {
		totalPages = 1
	}

	c.JSON(http.StatusOK, api.Response[[]*MatchOutput]{
		Data: outputs,
		Meta: &api.PaginationMeta{
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

func (h *MatchHandler) handleError(c *gin.Context, err error) {
	ctx := c.Request.Context()
	switch {
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
				Message: "an unexpected error occurred",
			},
		})
	}
}

func (h *MatchHandler) MethodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, api.Response[struct{}]{
		Error: &api.APIError{
			Code:    "METHOD_NOT_ALLOWED",
			Message: "method not allowed",
		},
	})
}
