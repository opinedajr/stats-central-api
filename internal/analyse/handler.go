package analyse

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/stats-central-api/internal/shared/api"
	"github.com/opinedajr/stats-central-api/internal/shared/logger"
)

var errInvalidID = errors.New("invalid id")

type AnalyseHandler struct {
	service Service
	logger  logger.Logger
}

func NewAnalyseHandler(service Service, logger logger.Logger) *AnalyseHandler {
	return &AnalyseHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AnalyseHandler) TeamTournamentAnalysis(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		h.handleError(c, errInvalidID)
		return
	}

	tournamentIDStr := c.Param("tournamentId")
	tournamentID, err := strconv.ParseUint(tournamentIDStr, 10, 32)
	if err != nil {
		h.handleError(c, errInvalidID)
		return
	}

	lastN, _ := strconv.Atoi(c.DefaultQuery("last_n", strconv.Itoa(defaultLastN)))

	output, err := h.service.TeamTournamentAnalysis(c.Request.Context(), uint(teamID), uint(tournamentID), lastN)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.Response[AnalyseOutput]{
		Data: output,
	})
}

func (h *AnalyseHandler) handleError(c *gin.Context, err error) {
	ctx := c.Request.Context()
	switch {
	case errors.Is(err, errInvalidID):
		c.JSON(http.StatusBadRequest, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "INVALID_ID",
				Message: "invalid id",
			},
		})
	case errors.Is(err, ErrTeamNotFound), errors.Is(err, ErrTournamentNotFound), errors.Is(err, ErrStatsNotFound):
		c.JSON(http.StatusNotFound, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "NOT_FOUND",
				Message: "resource not found",
			},
		})
	case errors.Is(err, ErrInvalidLastN):
		c.JSON(http.StatusBadRequest, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "INVALID_LAST_N",
				Message: "last_n must be between 1 and 50",
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

func (h *AnalyseHandler) MethodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, api.Response[struct{}]{
		Error: &api.APIError{
			Code:    "METHOD_NOT_ALLOWED",
			Message: "method not allowed",
		},
	})
}
