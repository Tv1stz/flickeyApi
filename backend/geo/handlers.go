// Package geo provides HTTP handlers for address geocoding and autocomplete endpoints.
package geo

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for address geocoding and search suggestions.
type Handler struct {
	Service *GeoService
}

// NewHandler constructs a new geo Handler.
func NewHandler(svc *GeoService) *Handler {
	return &Handler{Service: svc}
}

// RegisterRoutes registers geo endpoints on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/suggest", h.Suggest)
}

// Suggest handles GET /geo/suggest.
//
//	@Summary		Address autocomplete suggestions
//	@Description	Returns geocoded address suggestions and coordinates for a given search query (proxied to Photon OSM with Redis caching).
//	@Tags			Geo
//	@Produce		json
//	@Param			q		query		string	true	"Search query string (min 2 characters)"
//	@Param			lang	query		string	false	"Language code (default: ru)"
//	@Param			limit	query		int		false	"Max suggestions to return (default: 5, max: 10)"
//	@Success		200		{object}	GeoSuggestResponse
//	@Failure		500		{object}	auth.ErrorResponse
//	@Router			/geo/suggest [get]
func (h *Handler) Suggest(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if len(query) < 2 {
		c.JSON(http.StatusOK, GeoSuggestResponse{
			Results: []GeoSuggestItem{},
		})
		return
	}

	lang := c.DefaultQuery("lang", "ru")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit <= 0 || limit > 10 {
		limit = 5
	}

	resp, err := h.Service.Suggest(c.Request.Context(), query, lang, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "GEOCODER_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
