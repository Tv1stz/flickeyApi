package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultCacheTTL  = 7 * 24 * time.Hour
	tileCacheTTL     = 14 * 24 * time.Hour
	httpTimeout      = 4 * time.Second
	maxTileSizeBytes = 512 * 1024 // 512 KB maximum tile payload
)

// GeoService handles geocoding requests, upstream Photon / OSM communication, tile proxying, and Redis caching.
type GeoService struct {
	Redis         *redis.Client
	HTTPClient    *http.Client
	GeocoderURL   string
	TileServerURL string
	Logger        *slog.Logger
}

// NewGeoService constructs a new GeoService.
func NewGeoService(rdb *redis.Client, geocoderURL, tileServerURL string, logger *slog.Logger) *GeoService {
	if geocoderURL == "" {
		geocoderURL = "http://localhost:2322"
	}
	if tileServerURL == "" {
		tileServerURL = "http://localhost:8081"
	}
	return &GeoService{
		Redis: rdb,
		HTTPClient: &http.Client{
			Timeout: httpTimeout,
		},
		GeocoderURL:   geocoderURL,
		TileServerURL: tileServerURL,
		Logger:        logger,
	}
}

// Suggest fetches address suggestions matching the query string with intelligent fallback and strict Russian normalization.
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

	cacheKey := fmt.Sprintf("geo:v3:suggest:%s:%s:%d", strings.ToLower(query), lang, limit)

	// 1. Check Redis cache
	if s.Redis != nil {
		cachedData, err := s.Redis.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var cachedResp GeoSuggestResponse
			if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil && len(cachedResp.Results) > 0 {
				return &cachedResp, nil
			}
		}
	}

	// 2. Try primary configured Geocoder instance (self-hosted Photon / mock server)
	results := s.fetchFromPhoton(ctx, s.GeocoderURL, query, lang, limit)

	// 3. Fallback: try OSM Nominatim in Russian if primary returned empty / offline
	if len(results) == 0 {
		results = s.fetchFromNominatim(ctx, query, limit)
	}

	// 4. Fallback: try public Photon if still empty
	if len(results) == 0 {
		results = s.fetchFromPhoton(ctx, "https://photon.komoot.io", query, "", limit)
	}

	response := &GeoSuggestResponse{Results: results}

	// 5. Cache valid results in Redis
	if s.Redis != nil && len(results) > 0 {
		if data, err := json.Marshal(response); err == nil {
			_ = s.Redis.Set(ctx, cacheKey, data, defaultCacheTTL).Err()
		}
	}

	return response, nil
}

// ReverseGeocode retrieves the address details in Russian for a given latitude and longitude.
func (s *GeoService) ReverseGeocode(ctx context.Context, lat, lon float64, lang string) (*GeoReverseResponse, error) {
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return &GeoReverseResponse{Item: nil}, nil
	}

	if lang == "" {
		lang = "ru"
	}

	cacheKey := fmt.Sprintf("geo:v3:reverse:%.5f:%.5f:%s", lat, lon, lang)

	// 1. Check Redis cache
	if s.Redis != nil {
		cachedData, err := s.Redis.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var cachedResp GeoReverseResponse
			if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil && cachedResp.Item != nil {
				return &cachedResp, nil
			}
		}
	}

	// 2. Fetch from OSM Nominatim reverse endpoint
	upstreamURL := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/reverse?lat=%.6f&lon=%.6f&format=geojson&addressdetails=1&accept-language=ru,ru-RU;q=0.9&zoom=18",
		lat,
		lon,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL, nil)
	if err != nil {
		return &GeoReverseResponse{Item: nil}, fmt.Errorf("creating reverse request: %w", err)
	}
	req.Header.Set("User-Agent", "Flickey/1.0 (support@flickey.by)")
	req.Header.Set("Accept-Language", "ru,ru-RU;q=0.9,en;q=0.1")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		if s.Logger != nil {
			s.Logger.Debug("nominatim reverse request failed", "error", err)
		}
		return &GeoReverseResponse{Item: nil}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &GeoReverseResponse{Item: nil}, nil
	}

	var nc NominatimFeatureCollection
	if err := json.NewDecoder(resp.Body).Decode(&nc); err != nil || len(nc.Features) == 0 {
		return &GeoReverseResponse{Item: nil}, nil
	}

	item := transformNominatimFeature(nc.Features[0])
	item.Latitude = lat
	item.Longitude = lon

	response := &GeoReverseResponse{Item: &item}

	// 3. Cache valid response
	if s.Redis != nil && item.RawAddress != "" {
		if data, err := json.Marshal(response); err == nil {
			_ = s.Redis.Set(ctx, cacheKey, data, defaultCacheTTL).Err()
		}
	}

	return response, nil
}

// fetchFromNominatim queries OpenStreetMap Nominatim with explicit Russian language preference.
func (s *GeoService) fetchFromNominatim(ctx context.Context, query string, limit int) []GeoSuggestItem {
	upstreamURL := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/search?q=%s&format=geojson&addressdetails=1&accept-language=ru,ru-RU;q=0.9&limit=%d",
		url.QueryEscape(query),
		limit,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Flickey/1.0 (support@flickey.by)")
	req.Header.Set("Accept-Language", "ru,ru-RU;q=0.9,en;q=0.1")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		if s.Logger != nil {
			s.Logger.Debug("nominatim request failed", "error", err)
		}
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var nc NominatimFeatureCollection
	if err := json.NewDecoder(resp.Body).Decode(&nc); err != nil {
		return nil
	}

	items := make([]GeoSuggestItem, 0, len(nc.Features))
	for _, f := range nc.Features {
		item := transformNominatimFeature(f)
		if item.RawAddress != "" {
			items = append(items, item)
		}
	}

	return items
}

// fetchFromPhoton queries a Photon API endpoint and normalizes results to Russian.
func (s *GeoService) fetchFromPhoton(ctx context.Context, baseURL, query, lang string, limit int) []GeoSuggestItem {
	var upstreamURL string
	if lang != "" {
		upstreamURL = fmt.Sprintf(
			"%s/api?q=%s&lang=%s&limit=%d",
			strings.TrimRight(baseURL, "/"),
			url.QueryEscape(query),
			url.QueryEscape(lang),
			limit,
		)
	} else {
		upstreamURL = fmt.Sprintf(
			"%s/api?q=%s&limit=%d",
			strings.TrimRight(baseURL, "/"),
			url.QueryEscape(query),
			limit,
		)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Flickey/1.0")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		if s.Logger != nil {
			s.Logger.Debug("photon request failed", "url", upstreamURL, "error", err)
		}
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var fc PhotonFeatureCollection
	if err := json.NewDecoder(resp.Body).Decode(&fc); err != nil {
		return nil
	}

	items := make([]GeoSuggestItem, 0, len(fc.Features))
	for _, f := range fc.Features {
		item := transformPhotonFeature(f)
		if item.RawAddress != "" {
			items = append(items, item)
		}
	}

	return items
}

// transformNominatimFeature normalizes Nominatim GeoJSON properties into a Russian GeoSuggestItem.
func transformNominatimFeature(f NominatimFeature) GeoSuggestItem {
	p := f.Properties
	a := p.Address

	city := normalizeToRussian(a.City)
	if city == "" {
		city = normalizeToRussian(a.Town)
	}
	if city == "" {
		city = normalizeToRussian(a.Village)
	}

	street := normalizeToRussian(a.Road)
	houseNumber := normalizeToRussian(a.HouseNumber)
	country := normalizeToRussian(a.Country)

	var lon, lat float64
	if len(f.Geometry.Coordinates) >= 2 {
		lon = f.Geometry.Coordinates[0]
		lat = f.Geometry.Coordinates[1]
	}

	rawAddress := buildRawAddress(country, city, street, houseNumber, normalizeToRussian(p.Name))
	if rawAddress == "" && p.DisplayName != "" {
		rawAddress = normalizeToRussian(p.DisplayName)
	}

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

// transformPhotonFeature normalizes Photon GeoJSON properties into a Russian GeoSuggestItem.
func transformPhotonFeature(f PhotonFeature) GeoSuggestItem {
	p := f.Properties

	city := normalizeToRussian(p.City)
	if city == "" {
		city = normalizeToRussian(p.Town)
	}
	if city == "" {
		city = normalizeToRussian(p.Village)
	}
	if city == "" {
		city = normalizeToRussian(p.District)
	}

	street := normalizeToRussian(p.Street)
	houseNumber := normalizeToRussian(p.Housenumber)
	country := normalizeToRussian(p.Country)

	var lon, lat float64
	if len(f.Geometry.Coordinates) >= 2 {
		lon = f.Geometry.Coordinates[0]
		lat = f.Geometry.Coordinates[1]
	}

	rawAddress := buildRawAddress(country, city, street, houseNumber, normalizeToRussian(p.Name))

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

// buildRawAddress constructs a clean human-readable address line in Russian.
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

// normalizeToRussian translates Belarusian and CIS regional address terms/characters to standard Russian.
func normalizeToRussian(text string) string {
	if text == "" {
		return ""
	}

	replacements := []string{
		// Common address types
		"вуліца", "улица",
		"Вуліца", "Улица",
		"вул.", "ул.",
		"Вул.", "Ул.",
		"праспект", "проспект",
		"Праспект", "Проспект",
		"просп.", "просп.",
		"завулак", "переулок",
		"Завулак", "Переулок",
		"зав.", "пер.",
		"плошча", "площадь",
		"Плошча", "Площадь",
		"шоса", "шоссе",
		"Шоса", "Шоссе",
		"набярэжная", "набережная",
		"Набярэжная", "Набережная",
		"сельскі Савет", "сельский Совет",
		"сельскі савет", "сельский совет",
		"сельсавет", "сельсовет",
		"Сельсавет", "Сельсовет",
		"раён", "район",
		"Раён", "Район",
		"вобласць", "область",
		"Вобласць", "Область",
		"аграгарадок", "агрогородок",
		"Аграгарадок", "Агрогородок",
		"аг.", "аг.",
		"вёска", "деревня",
		"Вёска", "Деревня",
		"в.", "д.",
		"пасёлак", "посёлок",
		"Пасёлак", "Посёлок",
		"пас.", "пос.",
		"хутар", "хутор",
		"Хутар", "Хутор",
		"будынак", "здание",
		"Будынак", "Здание",

		// Cities & Towns
		"Мінск", "Минск",
		"Мінскі", "Минский",
		"Менск", "Минск",
		"Гродна", "Гродно",
		"Гродзенскі", "Гродненский",
		"Брэст", "Брест",
		"Брэсцкі", "Брестский",
		"Берасьце", "Брест",
		"Віцебск", "Витебск",
		"Віцебскі", "Витебский",
		"Магілёў", "Могилёв",
		"Магілёўскі", "Могилёвский",
		"Гомель", "Гомель",
		"Гомельскі", "Гомельский",
		"Баранавічы", "Барановичи",
		"Барысаў", "Борисов",
		"Пінск", "Пинск",
		"Орша", "Орша",
		"Ворша", "Орша",
		"Мазыр", "Мозырь",
		"Салігорск", "Солигорск",
		"Наваполацк", "Новополоцк",
		"Ліда", "Лида",
		"Маладзечна", "Молодечно",
		"Полацк", "Полоцк",
		"Жлобін", "Жлобин",
		"Светлагорск", "Светлогорск",
		"Рэчыца", "Речица",
		"Кобрын", "Кобрин",
		"Ваўкавыск", "Волковыск",
		"Калінкавічы", "Калинковичи",
		"Смаргонь", "Сморгонь",
		"Рагачоў", "Рогачев",
		"Асіповічы", "Осиповичи",
		"Горкі", "Горки",
		"Навагрудак", "Новогрудок",
		"Вілейка", "Вилейка",
		"Бяроза", "Берёза",
		"Крычаў", "Кричев",
		"Дзяржынск", "Дзержинск",
		"Івацэвічы", "Ивацевичи",
		"Лунінец", "Лунинец",
		"Мар'іна Горка", "Марьина Горка",
		"Паставы", "Поставы",
		"Пружаны", "Пружаны",
		"Глыбокае", "Глубокое",
		"Добруш", "Добруш",
		"Лепель", "Лепель",
		"Быхаў", "Быхов",
		"Іванава", "Иваново",
		"Клімавічы", "Климовичи",
		"Шклоў", "Шклов",
		"Касцюковічы", "Костюковичи",
		"Столін", "Столин",
		"Жыткавічы", "Житковичи",
		"Мосты", "Мосты",
		"Ашмяны", "Ошмяны",
		"Шчучын", "Щучин",
		"Драгічын", "Дрогичин",
		"Нясвіж", "Несвиж",
		"Ганцавічы", "Ганцевичи",
		"Хойнікі", "Хойники",
		"Фаніпаль", "Фаниполь",
		"Мікашэвічы", "Микашевичи",
		"Новалукомль", "Новолукомль",
		"Гарадок", "Городок",
		"Стоўбцы", "Столбцы",
		"Смалявічы", "Смолевичи",
		"Белаазёрск", "Белоозёрск",
		"Бярэзань", "Березино",
		"Бярозаўка", "Берёзовка",
		"Любань", "Любань",
		"Старыя Дарогі", "Старые Дороги",
		"Заслаўе", "Заславль",
		"Заслаўль", "Заславль",
		"Лагойск", "Логойск",
		"Клецк", "Клецк",
		"Узда", "Узда",
		"Чэрвень", "Червень",
		"Капыль", "Копыль",
		"Валожын", "Воложин",
		"Крупкі", "Крупки",
		"Мядзел", "Мядель",
		"Міханавічы", "Михановичи",
		"Міханавіцкі", "Михановичский",
		"Беларусь", "Беларусь",
		"Расія", "Россия",
	}

	replacer := strings.NewReplacer(replacements...)
	res := replacer.Replace(text)

	// Character-level cleanup for any remaining Belarusian characters
	charReplacer := strings.NewReplacer(
		"і", "и",
		"І", "И",
		"ў", "в",
		"Ў", "В",
		"'", "",
		"’", "",
	)
	res = charReplacer.Replace(res)

	return strings.TrimSpace(res)
}

// ProxyTile serves raster map tiles with Redis caching and upstream fallbacks.
func (s *GeoService) ProxyTile(ctx context.Context, z, x, y int) ([]byte, string, error) {
	cacheKey := fmt.Sprintf("geo:tile:%d:%d:%d", z, x, y)

	// 1. Check Redis cache
	if s.Redis != nil {
		if cachedData, err := s.Redis.Get(ctx, cacheKey).Bytes(); err == nil && len(cachedData) > 0 {
			return cachedData, "image/png", nil
		}
	}

	// 2. Fetch tile (try configured TileServer if custom, or OpenStreetMap standard tiles)
	var data []byte
	var err error
	if s.TileServerURL != "" && !strings.Contains(s.TileServerURL, "localhost:8081") {
		tileServerURL := fmt.Sprintf("%s/styles/osm-bright/%d/%d/%d.png", strings.TrimRight(s.TileServerURL, "/"), z, x, y)
		localCtx, localCancel := context.WithTimeout(ctx, 400*time.Millisecond)
		data, err = s.fetchTileHTTP(localCtx, tileServerURL, "Flickey/1.0")
		localCancel()
	}

	if err != nil || len(data) == 0 {
		fallbackURL := fmt.Sprintf("https://tile.openstreetmap.org/%d/%d/%d.png", z, x, y)
		data, err = s.fetchTileHTTP(ctx, fallbackURL, "Flickey/1.0 (support@flickey.by)")
	}

	if err != nil || len(data) == 0 {
		return nil, "", fmt.Errorf("tile not found")
	}

	// 4. Cache valid tile in Redis
	if s.Redis != nil && len(data) > 0 {
		_ = s.Redis.Set(ctx, cacheKey, data, tileCacheTTL).Err()
	}

	return data, "image/png", nil
}

// fetchTileHTTP safely performs a GET request for a map tile with bounded memory reads.
func (s *GeoService) fetchTileHTTP(ctx context.Context, targetURL, userAgent string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected upstream status: %d", resp.StatusCode)
	}

	limitedReader := io.LimitReader(resp.Body, maxTileSizeBytes)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty tile payload")
	}

	return data, nil
}
