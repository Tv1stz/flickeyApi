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
	rg.GET("/reverse", h.Reverse)
	rg.GET("/tiles/:z/:x/:y", h.Tile)
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

// Reverse handles GET /geo/reverse.
//
//	@Summary		Reverse geocoding (coordinates to address)
//	@Description	Returns formatted address details in Russian for given latitude and longitude coordinates.
//	@Tags			Geo
//	@Produce		json
//	@Param			lat		query		number	true	"Latitude"
//	@Param			lon		query		number	true	"Longitude"
//	@Param			lang	query		string	false	"Language code (default: ru)"
//	@Success		200		{object}	GeoReverseResponse
//	@Failure		400		{object}	auth.ErrorResponse
//	@Failure		500		{object}	auth.ErrorResponse
//	@Router			/geo/reverse [get]
func (h *Handler) Reverse(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	if lonStr == "" {
		lonStr = c.Query("lng")
	}

	lat, err1 := strconv.ParseFloat(latStr, 64)
	lon, err2 := strconv.ParseFloat(lonStr, 64)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_COORDINATES",
			"message": "Valid lat and lon query parameters are required",
		})
		return
	}

	lang := c.DefaultQuery("lang", "ru")

	resp, err := h.Service.ReverseGeocode(c.Request.Context(), lat, lon, lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "GEOCODER_ERROR",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Tile handles GET /geo/tiles/:z/:x/:y.
//
//	@Summary		Map tile proxy
//	@Description	Proxies raster map tiles from TileServer GL with caching.
//	@Tags			Geo
//	@Produce		image/png
//	@Param			z		path		int	true	"Zoom level (0-19)"
//	@Param			x		path		int	true	"Tile X coordinate"
//	@Param			y		path		string	true	"Tile Y coordinate (optionally with .png suffix)"
//	@Success		200		{file}		binary
//	@Failure		400		{object}	auth.ErrorResponse
//	@Failure		404		{object}	auth.ErrorResponse
//	@Router			/geo/tiles/{z}/{x}/{y} [get]
func (h *Handler) Tile(c *gin.Context) {
	zStr := c.Param("z")
	xStr := c.Param("x")
	yStr := strings.TrimSuffix(c.Param("y"), ".png")

	z, errZ := strconv.Atoi(zStr)
	x, errX := strconv.Atoi(xStr)
	y, errY := strconv.Atoi(yStr)

	if errZ != nil || errX != nil || errY != nil || z < 0 || z > 19 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_TILE_COORDINATES",
			"message": "Zoom level z must be an integer between 0 and 19, and x, y must be valid integers",
		})
		return
	}

	maxCoord := 1 << z
	if x < 0 || x >= maxCoord || y < 0 || y >= maxCoord {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "TILE_OUT_OF_BOUNDS",
			"message": "Tile coordinates x or y are out of bounds for the specified zoom level",
		})
		return
	}

	data, contentType, err := h.Service.ProxyTile(c.Request.Context(), z, x, y)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.Header("Cache-Control", "public, max-age=604800, immutable")
	c.Data(http.StatusOK, contentType, data)
}
