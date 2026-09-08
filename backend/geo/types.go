// Package geo provides geocoding, address autocomplete, and reverse-geocoding services.
package geo

// GeoSuggestItem represents a normalized address suggestion with geographic coordinates.
type GeoSuggestItem struct {
	Country     string  `json:"country"`
	City        string  `json:"city"`
	Street      string  `json:"street"`
	HouseNumber string  `json:"house_number"`
	RawAddress  string  `json:"raw_address"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// GeoSuggestResponse is the standard response for address suggestion queries.
type GeoSuggestResponse struct {
	Results []GeoSuggestItem `json:"results"`
}

// GeoReverseResponse represents the reverse geocoding result for geographic coordinates.
type GeoReverseResponse struct {
	Item *GeoSuggestItem `json:"item"`
}

// PhotonFeatureCollection represents the raw GeoJSON returned by Photon.
type PhotonFeatureCollection struct {
	Features []PhotonFeature `json:"features"`
}

// PhotonFeature represents a single feature in Photon's GeoJSON output.
type PhotonFeature struct {
	Properties PhotonProperties `json:"properties"`
	Geometry   PhotonGeometry   `json:"geometry"`
}

// PhotonProperties contains address attributes extracted by Photon.
type PhotonProperties struct {
	Country     string `json:"country"`
	City        string `json:"city"`
	Town        string `json:"town"`
	Village     string `json:"village"`
	County      string `json:"county"`
	State       string `json:"state"`
	Street      string `json:"street"`
	Housenumber string `json:"housenumber"`
	Name        string `json:"name"`
	Postcode    string `json:"postcode"`
	District    string `json:"district"`
	Type        string `json:"type"`
}

// PhotonGeometry holds the coordinates in [longitude, latitude] format.
type PhotonGeometry struct {
	Coordinates []float64 `json:"coordinates"` // [lon, lat]
}

// NominatimFeatureCollection represents the GeoJSON returned by OpenStreetMap Nominatim.
type NominatimFeatureCollection struct {
	Features []NominatimFeature `json:"features"`
}

// NominatimFeature represents a single feature in Nominatim GeoJSON.
type NominatimFeature struct {
	Properties NominatimProperties `json:"properties"`
	Geometry   PhotonGeometry      `json:"geometry"`
}

// NominatimProperties contains address details from Nominatim.
type NominatimProperties struct {
	DisplayName string           `json:"display_name"`
	Name        string           `json:"name"`
	Address     NominatimAddress `json:"address"`
}

// NominatimAddress contains granular address components from Nominatim.
type NominatimAddress struct {
	Country     string `json:"country"`
	City        string `json:"city"`
	Town        string `json:"town"`
	Village     string `json:"village"`
	Road        string `json:"road"`
	HouseNumber string `json:"house_number"`
	State       string `json:"state"`
	Postcode    string `json:"postcode"`
}
