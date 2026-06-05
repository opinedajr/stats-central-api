package tournament

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/stats-central-api/internal/shared/api"
	"github.com/opinedajr/stats-central-api/internal/shared/logger"
)

var errInvalidID = errors.New("invalid id")

type TournamentHandler struct {
	service Service
	logger  logger.Logger
}

func NewTournamentHandler(service Service, logger logger.Logger) *TournamentHandler {
	return &TournamentHandler{
		service: service,
		logger:  logger,
	}
}

func (h *TournamentHandler) Create(c *gin.Context) {
	defer func() { _ = c.Request.Body.Close() }()
	ctx := c.Request.Context()

	var input CreateTournamentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error(ctx, "invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request body",
			},
		})
		return
	}

	output, err := h.service.CreateTournament(ctx, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, api.Response[*TournamentOutput]{
		Data: output,
	})
}

func (h *TournamentHandler) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil {
		pageSize = 20
	}

	var filter TournamentFilter

	if activeStr := c.Query("active"); activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err == nil {
			filter.Active = &active
		}
	}

	if country := c.Query("country"); country != "" {
		filter.Country = &country
	}

	if divisionStr := c.Query("division"); divisionStr != "" {
		division, err := strconv.Atoi(divisionStr)
		if err == nil {
			filter.Division = &division
		}
	}

	if season := c.Query("season"); season != "" {
		filter.Season = &season
	}

	outputs, total, err := h.service.ListTournaments(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 {
		totalPages = 1
	}

	c.JSON(http.StatusOK, api.Response[[]*TournamentOutput]{
		Data: outputs,
		Meta: &api.PaginationMeta{
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

func (h *TournamentHandler) Get(c *gin.Context) {
	tournamentIDStr := c.Param("tournamentId")
	tournamentID, err := strconv.ParseUint(tournamentIDStr, 10, 32)
	if err != nil {
		h.handleError(c, errInvalidID)
		return
	}

	output, err := h.service.GetTournamentByID(c.Request.Context(), uint(tournamentID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.Response[*TournamentOutput]{
		Data: output,
	})
}

func (h *TournamentHandler) Update(c *gin.Context) {
	defer func() { _ = c.Request.Body.Close() }()
	ctx := c.Request.Context()

	var input UpdateTournamentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error(ctx, "invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request body",
			},
		})
		return
	}

	tournamentIDStr := c.Param("tournamentId")
	tournamentID, err := strconv.ParseUint(tournamentIDStr, 10, 32)
	if err != nil {
		h.handleError(c, errInvalidID)
		return
	}

	output, err := h.service.UpdateTournament(ctx, uint(tournamentID), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.Response[*TournamentOutput]{
		Data: output,
	})
}

func (h *TournamentHandler) UpdateStatus(c *gin.Context) {
	defer func() { _ = c.Request.Body.Close() }()
	ctx := c.Request.Context()

	var input UpdateTournamentStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error(ctx, "invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request body",
			},
		})
		return
	}

	tournamentIDStr := c.Param("tournamentId")
	tournamentID, err := strconv.ParseUint(tournamentIDStr, 10, 32)
	if err != nil {
		h.handleError(c, errInvalidID)
		return
	}

	output, err := h.service.UpdateTournamentStatus(ctx, uint(tournamentID), input.Active)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.Response[*TournamentOutput]{
		Data: output,
	})
}

func (h *TournamentHandler) handleError(c *gin.Context, err error) {
	ctx := c.Request.Context()
	switch {
	case errors.Is(err, errInvalidID):
		c.JSON(http.StatusBadRequest, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "INVALID_ID",
				Message: "invalid id",
			},
		})
	case errors.Is(err, ErrTournamentNotFound):
		c.JSON(http.StatusNotFound, api.Response[struct{}]{
			Error: &api.APIError{
				Code:    "TOURNAMENT_NOT_FOUND",
				Message: "tournament not found",
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
