package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultCacheTTL = 7 * 24 * time.Hour
	httpTimeout     = 3 * time.Second
)

// GeoService handles geocoding requests, upstream Photon communication, and Redis caching.
type GeoService struct {
	Redis       *redis.Client
	HTTPClient  *http.Client
	GeocoderURL string
	Logger      *slog.Logger
}

// NewGeoService constructs a new GeoService.
func NewGeoService(rdb *redis.Client, geocoderURL string, logger *slog.Logger) *GeoService {
	if geocoderURL == "" {
		geocoderURL = "http://localhost:2322"
	}
	return &GeoService{
		Redis: rdb,
		HTTPClient: &http.Client{
			Timeout: httpTimeout,
		},
		GeocoderURL: geocoderURL,
		Logger:      logger,
	}
}

// Suggest fetches address suggestions matching the query string.
func (s *GeoService) Suggest(ctx context.Context, query string, lang string, limit int) (*GeoSuggestResponse, error) {
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return &GeoSuggestResponse{Results: []GeoSuggestItem{}}, nil
	}

	if lang == "" {
		lang = "ru"
	}
	if limit <= 0 || limit > 10 {
		limit = 5
	}

	cacheKey := fmt.Sprintf("geo:suggest:%s:%s:%d", strings.ToLower(query), lang, limit)

	// 1. Check Redis cache
	if s.Redis != nil {
		cachedData, err := s.Redis.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var cachedResp GeoSuggestResponse
			if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil {
				return &cachedResp, nil
			}
		}
	}

	// 2. Fetch from upstream Photon service
	upstreamURL := fmt.Sprintf(
		"%s/api?q=%s&lang=%s&limit=%d",
		strings.TrimRight(s.GeocoderURL, "/"),
		url.QueryEscape(query),
		url.QueryEscape(lang),
		limit,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL, nil)
	if err != nil {
		return &GeoSuggestResponse{Results: []GeoSuggestItem{}}, fmt.Errorf("creating geocoder request: %w", err)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		if s.Logger != nil {
			s.Logger.Warn("geocoder upstream request failed", "url", upstreamURL, "error", err)
		}
		// Graceful degradation: return empty results instead of crashing
		return &GeoSuggestResponse{Results: []GeoSuggestItem{}}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if s.Logger != nil {
			s.Logger.Warn("geocoder upstream returned non-200 status", "status", resp.StatusCode)
		}
		return &GeoSuggestResponse{Results: []GeoSuggestItem{}}, nil
	}

	var fc PhotonFeatureCollection
	if err := json.NewDecoder(resp.Body).Decode(&fc); err != nil {
		if s.Logger != nil {
			s.Logger.Warn("failed to decode geocoder response", "error", err)
		}
		return &GeoSuggestResponse{Results: []GeoSuggestItem{}}, nil
	}

	// 3. Transform features into standardized GeoSuggestItems
	results := make([]GeoSuggestItem, 0, len(fc.Features))
	for _, f := range fc.Features {
		item := transformPhotonFeature(f)
		if item.RawAddress != "" {
			results = append(results, item)
		}
	}

	response := &GeoSuggestResponse{Results: results}

	// 4. Cache results in Redis
	if s.Redis != nil && len(results) > 0 {
		if data, err := json.Marshal(response); err == nil {
			_ = s.Redis.Set(ctx, cacheKey, data, defaultCacheTTL).Err()
		}
	}

	return response, nil
}

// transformPhotonFeature normalizes Photon GeoJSON properties into a GeoSuggestItem.
func transformPhotonFeature(f PhotonFeature) GeoSuggestItem {
	p := f.Properties

	city := p.City
	if city == "" {
		city = p.Town
	}
	if city == "" {
		city = p.Village
	}
	if city == "" {
		city = p.District
	}

	street := p.Street
	houseNumber := p.Housenumber
	country := p.Country

	var lon, lat float64
	if len(f.Geometry.Coordinates) >= 2 {
		lon = f.Geometry.Coordinates[0]
		lat = f.Geometry.Coordinates[1]
	}

	rawAddress := buildRawAddress(country, city, street, houseNumber, p.Name)

	return GeoSuggestItem{
		Country:     country,
		City:        city,
		Street:      street,
		HouseNumber: houseNumber,
		RawAddress:  rawAddress,
		Latitude:    lat,
		Longitude:   lon,
	}
}

// buildRawAddress constructs a clean human-readable address line.
func buildRawAddress(country, city, street, houseNumber, name string) string {
	parts := make([]string, 0, 4)

	if city != "" {
		parts = append(parts, city)
	}

	if street != "" {
		if houseNumber != "" {
			parts = append(parts, fmt.Sprintf("%s, %s", street, houseNumber))
		} else {
			parts = append(parts, street)
		}
	} else if name != "" && name != city {
		parts = append(parts, name)
	} else if houseNumber != "" {
		parts = append(parts, houseNumber)
	}

	if country != "" && (city == "" || !strings.EqualFold(country, city)) {
		parts = append(parts, country)
	}

	return strings.Join(parts, ", ")
}
