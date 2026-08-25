// Package amenities provides HTTP handlers for amenity discovery.
package amenities

import (
	"net/http"

	_ "flickey/go-backend/auth"

	"github.com/gin-gonic/gin"
)

// AmenitiesResponse is the response for GET /amenities.
type AmenitiesResponse struct {
	HousingType *string                   `json:"housing_type"`
	Categories  []AmenityCategoryResponse `json:"categories"`
}

// AmenityCategoryResponse is a single amenity category in the API response.
type AmenityCategoryResponse struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Amenities []AmenityItem `json:"amenities"`
}

// RegisterRoutes registers amenity routes under the provided RouterGroup.
func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", GetAmenities)
}

// GetAmenities handles GET /amenities?housing_type=...
//
//	@Summary		Get amenity catalogue
//	@Description	Returns categorized amenities, optionally filtered for a specific housing type (apartment, house, manor).
//	@Tags			Amenities
//	@Produce		json
//	@Param			housing_type	query		string				false	"Filter by housing type: apartment, house, manor"	Enums(apartment, house, manor)
//	@Success		200				{object}	AmenitiesResponse	"Categorized amenity catalogue"
//	@Failure		400				{object}	auth.ErrorResponse	"INVALID_HOUSING_TYPE"
//	@Router			/amenities [get]
func GetAmenities(c *gin.Context) {
	housingType := c.Query("housing_type")

	if housingType != "" && !IsValidHousingType(housingType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_HOUSING_TYPE",
			"message": "housing_type must be one of: apartment, house, manor",
		})
		return
	}

	grouped := GetGroupedAmenities(housingType)

	categories := make([]AmenityCategoryResponse, len(grouped))
	for i, g := range grouped {
		categories[i] = AmenityCategoryResponse{
			ID:        g.ID,
			Name:      g.Name,
			Amenities: g.Amenities,
		}
	}

	resp := AmenitiesResponse{
		Categories: categories,
	}
	if housingType != "" {
		resp.HousingType = &housingType
	}

	c.JSON(http.StatusOK, resp)
}
