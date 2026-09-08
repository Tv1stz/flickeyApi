package geo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGeoSuggestItem_JSON_SnakeCase(t *testing.T) {
	item := GeoSuggestItem{
		Country:     "Беларусь",
		City:        "Минск",
		Street:      "проспект Победителей",
		HouseNumber: "7",
		RawAddress:  "Минск, проспект Победителей, 7, Беларусь",
		Latitude:    53.9056,
		Longitude:   27.5542,
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal GeoSuggestItem: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	expectedKeys := []string{
		"country",
		"city",
		"street",
		"house_number",
		"raw_address",
		"latitude",
		"longitude",
	}

	for _, key := range expectedKeys {
		if _, exists := raw[key]; !exists {
			t.Errorf("expected JSON key %q in serialized output", key)
		}
	}

	// Verify absence of PascalCase / camelCase keys
	prohibitedKeys := []string{"Country", "City", "Street", "HouseNumber", "houseNumber", "RawAddress", "rawAddress", "Latitude", "Longitude"}
	for _, key := range prohibitedKeys {
		if _, exists := raw[key]; exists {
			t.Errorf("found prohibited key %q in serialized JSON", key)
		}
	}
}

func TestTransformPhotonFeature(t *testing.T) {
	feat := PhotonFeature{
		Properties: PhotonProperties{
			Country:     "Беларусь",
			City:        "Минск",
			Street:      "проспект Независимости",
			Housenumber: "10",
		},
		Geometry: PhotonGeometry{
			Coordinates: []float64{27.5615, 53.9006},
		},
	}

	item := transformPhotonFeature(feat)

	if item.Country != "Беларусь" {
		t.Errorf("expected country 'Беларусь', got %q", item.Country)
	}
	if item.City != "Минск" {
		t.Errorf("expected city 'Минск', got %q", item.City)
	}
	if item.Street != "проспект Независимости" {
		t.Errorf("expected street 'проспект Независимости', got %q", item.Street)
	}
	if item.HouseNumber != "10" {
		t.Errorf("expected house_number '10', got %q", item.HouseNumber)
	}
	if item.Longitude != 27.5615 {
		t.Errorf("expected longitude 27.5615, got %v", item.Longitude)
	}
	if item.Latitude != 53.9006 {
		t.Errorf("expected latitude 53.9006, got %v", item.Latitude)
	}
	if item.RawAddress != "Минск, проспект Независимости, 10, Беларусь" {
		t.Errorf("expected raw_address 'Минск, проспект Независимости, 10, Беларусь', got %q", item.RawAddress)
	}
}

func TestGeoService_ShortQueryReturnsEmpty(t *testing.T) {
	svc := NewGeoService(nil, "http://localhost:2322", "", nil)
	resp, err := svc.Suggest(context.Background(), "a", "ru", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected empty results for short query, got %d items", len(resp.Results))
	}
}

func TestGeoService_PhotonMockResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fc := PhotonFeatureCollection{
			Features: []PhotonFeature{
				{
					Properties: PhotonProperties{
						Country:     "Беларусь",
						City:        "Минск",
						Street:      "улица Ленина",
						Housenumber: "1",
					},
					Geometry: PhotonGeometry{
						Coordinates: []float64{27.5580, 53.9020},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fc)
	}))
	defer server.Close()

	svc := NewGeoService(nil, server.URL, "", nil)
	resp, err := svc.Suggest(context.Background(), "Ленина", "ru", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].City != "Минск" || resp.Results[0].Street != "улица Ленина" {
		t.Errorf("unexpected result: %+v", resp.Results[0])
	}
}

func TestHandler_Tile_Validation(t *testing.T) {
	svc := NewGeoService(nil, "", "", nil)
	handler := NewHandler(svc)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "invalid zoom negative",
			path:       "/tiles/-1/0/0",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_TILE_COORDINATES",
		},
		{
			name:       "invalid zoom excessive",
			path:       "/tiles/20/0/0",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_TILE_COORDINATES",
		},
		{
			name:       "x coordinate out of bounds for z=2",
			path:       "/tiles/2/4/0",
			wantStatus: http.StatusBadRequest,
			wantCode:   "TILE_OUT_OF_BOUNDS",
		},
		{
			name:       "y coordinate out of bounds for z=2",
			path:       "/tiles/2/0/4.png",
			wantStatus: http.StatusBadRequest,
			wantCode:   "TILE_OUT_OF_BOUNDS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Parse params from path
			parts := strings.Split(strings.TrimPrefix(tt.path, "/tiles/"), "/")
			c.Params = gin.Params{
				{Key: "z", Value: parts[0]},
				{Key: "x", Value: parts[1]},
				{Key: "y", Value: parts[2]},
			}
			c.Request = httptest.NewRequest(http.MethodGet, tt.path, nil)

			handler.Tile(c)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.wantStatus, w.Code, w.Body.String())
			}

			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err == nil {
				if code, ok := body["code"].(string); ok && code != tt.wantCode {
					t.Errorf("expected error code %q, got %q", tt.wantCode, code)
				}
			}
		})
	}
}

func TestGeoService_ProxyTile_MockServer(t *testing.T) {
	expectedPng := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write(expectedPng)
	}))
	defer server.Close()

	svc := NewGeoService(nil, "", server.URL, nil)
	data, contentType, err := svc.ProxyTile(context.Background(), 2, 1, 1)
	if err != nil {
		t.Fatalf("unexpected ProxyTile error: %v", err)
	}
	if contentType != "image/png" {
		t.Errorf("expected contentType image/png, got %q", contentType)
	}
	if len(data) != len(expectedPng) {
		t.Errorf("expected %d bytes, got %d", len(expectedPng), len(data))
	}
}
