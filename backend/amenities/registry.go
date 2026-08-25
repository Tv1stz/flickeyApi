// Package amenities provides the static in-memory amenity registry.
// This is a direct Go port of app/core/amenities.py.
// All amenity IDs, categories, and Russian display names are identical.
package amenities

// HousingType represents the supported property types.
type HousingType string

const (
	HousingTypeApartment HousingType = "apartment"
	HousingTypeHouse     HousingType = "house"
	HousingTypeManor     HousingType = "manor"
)

// AllHousingTypes is the set of all valid housing types.
var AllHousingTypes = map[HousingType]struct{}{
	HousingTypeApartment: {},
	HousingTypeHouse:     {},
	HousingTypeManor:     {},
}

// IsValidHousingType returns true if the given type is valid.
func IsValidHousingType(t string) bool {
	_, ok := AllHousingTypes[HousingType(t)]
	return ok
}

// AmenityDefinition describes a single amenity.
type AmenityDefinition struct {
	ID           string
	Category     string
	DisplayName  string
	AllowedTypes map[HousingType]struct{}
}

// IsAllowedFor returns true if this amenity is valid for the given housing type.
func (a AmenityDefinition) IsAllowedFor(t HousingType) bool {
	_, ok := a.AllowedTypes[t]
	return ok
}

// AmenityCategory groups amenities for display.
type AmenityCategory struct {
	ID        string
	Name      string
	Amenities []AmenityItem
}

// AmenityItem is an amenity for API responses.
type AmenityItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var all = AllHousingTypes

// Registry is the authoritative in-memory amenity store.
// Sorted in the same order as the Python implementation.
var Registry = map[string]AmenityDefinition{
	// ── Basic ─────────────────────────────────────────────────────────────────
	"wifi":             {ID: "wifi", Category: "basic", DisplayName: "Wi-Fi", AllowedTypes: all},
	"heating":          {ID: "heating", Category: "basic", DisplayName: "Отопление", AllowedTypes: all},
	"air_conditioning": {ID: "air_conditioning", Category: "basic", DisplayName: "Кондиционер", AllowedTypes: all},
	"hot_water":        {ID: "hot_water", Category: "basic", DisplayName: "Горячая вода", AllowedTypes: all},
	"washing_machine":  {ID: "washing_machine", Category: "basic", DisplayName: "Стиральная машина", AllowedTypes: all},
	"dryer":            {ID: "dryer", Category: "basic", DisplayName: "Сушильная машина", AllowedTypes: all},
	"tv":               {ID: "tv", Category: "basic", DisplayName: "Телевизор", AllowedTypes: all},
	"iron":             {ID: "iron", Category: "basic", DisplayName: "Утюг", AllowedTypes: all},
	"hair_dryer":       {ID: "hair_dryer", Category: "basic", DisplayName: "Фен", AllowedTypes: all},
	"towels":           {ID: "towels", Category: "basic", DisplayName: "Полотенца", AllowedTypes: all},
	"bed_linen":        {ID: "bed_linen", Category: "basic", DisplayName: "Постельное белье", AllowedTypes: all},
	// ── Kitchen ───────────────────────────────────────────────────────────────
	"full_kitchen":   {ID: "full_kitchen", Category: "kitchen", DisplayName: "Полноценная кухня", AllowedTypes: all},
	"refrigerator":   {ID: "refrigerator", Category: "kitchen", DisplayName: "Холодильник", AllowedTypes: all},
	"stove":          {ID: "stove", Category: "kitchen", DisplayName: "Плита", AllowedTypes: all},
	"oven":           {ID: "oven", Category: "kitchen", DisplayName: "Духовка", AllowedTypes: all},
	"dishwasher":     {ID: "dishwasher", Category: "kitchen", DisplayName: "Посудомоечная машина", AllowedTypes: all},
	"microwave":      {ID: "microwave", Category: "kitchen", DisplayName: "Микроволновка", AllowedTypes: all},
	"coffee_machine": {ID: "coffee_machine", Category: "kitchen", DisplayName: "Кофемашина", AllowedTypes: all},
	"kettle":         {ID: "kettle", Category: "kitchen", DisplayName: "Чайник", AllowedTypes: all},
	"toaster":        {ID: "toaster", Category: "kitchen", DisplayName: "Тостер", AllowedTypes: all},
	"dining_area":    {ID: "dining_area", Category: "kitchen", DisplayName: "Обеденная зона", AllowedTypes: all},
	// ── Bedroom and Bathroom ──────────────────────────────────────────────────
	"extra_pillows_blankets": {ID: "extra_pillows_blankets", Category: "bedroom_and_bathroom", DisplayName: "Дополнительные подушки и одеяла", AllowedTypes: all},
	"blackout_curtains":      {ID: "blackout_curtains", Category: "bedroom_and_bathroom", DisplayName: "Плотные шторы", AllowedTypes: all},
	"bathtub":                {ID: "bathtub", Category: "bedroom_and_bathroom", DisplayName: "Ванна", AllowedTypes: all},
	"shower":                 {ID: "shower", Category: "bedroom_and_bathroom", DisplayName: "Душ", AllowedTypes: all},
	"bidet":                  {ID: "bidet", Category: "bedroom_and_bathroom", DisplayName: "Биде", AllowedTypes: all},
	// ── Work ──────────────────────────────────────────────────────────────────
	"workspace":        {ID: "workspace", Category: "work", DisplayName: "Рабочее место", AllowedTypes: all},
	"external_monitor": {ID: "external_monitor", Category: "work", DisplayName: "Внешний монитор", AllowedTypes: all},
	// ── Comfort and Leisure ───────────────────────────────────────────────────
	"balcony_terrace": {ID: "balcony_terrace", Category: "comfort_and_leisure", DisplayName: "Балкон / терраса", AllowedTypes: all},
	"gym":             {ID: "gym", Category: "comfort_and_leisure", DisplayName: "Тренажерный зал", AllowedTypes: all},
	"great_view":      {ID: "great_view", Category: "comfort_and_leisure", DisplayName: "Красивый вид", AllowedTypes: all},
	"sofa_lounge":     {ID: "sofa_lounge", Category: "comfort_and_leisure", DisplayName: "Диван / зона отдыха", AllowedTypes: all},
	"board_games":     {ID: "board_games", Category: "comfort_and_leisure", DisplayName: "Настольные игры", AllowedTypes: all},
	"books":           {ID: "books", Category: "comfort_and_leisure", DisplayName: "Книги", AllowedTypes: all},
	// ── Family ────────────────────────────────────────────────────────────────
	"baby_crib":    {ID: "baby_crib", Category: "family", DisplayName: "Детская кроватка", AllowedTypes: all},
	"high_chair":   {ID: "high_chair", Category: "family", DisplayName: "Стульчик для кормления", AllowedTypes: all},
	"pets_allowed": {ID: "pets_allowed", Category: "family", DisplayName: "Можно с питомцами", AllowedTypes: all},
	"toys":         {ID: "toys", Category: "family", DisplayName: "Игрушки", AllowedTypes: all},
	"baby_bath":    {ID: "baby_bath", Category: "family", DisplayName: "Детская ванночка", AllowedTypes: all},
	// ── Safety ────────────────────────────────────────────────────────────────
	"smoke_detector":           {ID: "smoke_detector", Category: "safety", DisplayName: "Датчик дыма", AllowedTypes: all},
	"carbon_monoxide_detector": {ID: "carbon_monoxide_detector", Category: "safety", DisplayName: "Датчик угарного газа", AllowedTypes: all},
	"fire_extinguisher":        {ID: "fire_extinguisher", Category: "safety", DisplayName: "Огнетушитель", AllowedTypes: all},
	"first_aid_kit":            {ID: "first_aid_kit", Category: "safety", DisplayName: "Аптечка", AllowedTypes: all},
	"safe":                     {ID: "safe", Category: "safety", DisplayName: "Сейф", AllowedTypes: all},
	// ── Access and Parking ────────────────────────────────────────────────────
	"self_check_in": {ID: "self_check_in", Category: "access_and_parking", DisplayName: "Бесконтактное заселение", AllowedTypes: all},
	"elevator":      {ID: "elevator", Category: "access_and_parking", DisplayName: "Лифт", AllowedTypes: all},
	"parking":       {ID: "parking", Category: "access_and_parking", DisplayName: "Парковка", AllowedTypes: all},
	"ev_charger":    {ID: "ev_charger", Category: "access_and_parking", DisplayName: "Зарядка для электромобилей", AllowedTypes: all},
	// ── Accessibility ─────────────────────────────────────────────────────────
	"wide_entrance":       {ID: "wide_entrance", Category: "accessibility", DisplayName: "Широкий дверной проем", AllowedTypes: all},
	"step_free_entrance":  {ID: "step_free_entrance", Category: "accessibility", DisplayName: "Вход без ступеней", AllowedTypes: all},
	"accessible_bathroom": {ID: "accessible_bathroom", Category: "accessibility", DisplayName: "Оборудованная ванная комната", AllowedTypes: all},
}

// categoriesOrdered preserves the original Python category order.
var categoriesOrdered = []struct{ ID, Name string }{
	{"basic", "Основное"},
	{"kitchen", "Кухня"},
	{"bedroom_and_bathroom", "Спальня и ванная"},
	{"work", "Работа"},
	{"comfort_and_leisure", "Комфорт и отдых"},
	{"family", "Для семей"},
	{"safety", "Безопасность"},
	{"access_and_parking", "Доступ и парковка"},
	{"accessibility", "Доступность"},
}

// GetAmenity looks up an amenity by ID.
func GetAmenity(id string) (AmenityDefinition, bool) {
	a, ok := Registry[id]
	return a, ok
}

// GetGroupedAmenities returns amenities grouped by category, optionally filtered by housing type.
func GetGroupedAmenities(housingType string) []AmenityCategory {
	var filterType *HousingType
	if housingType != "" {
		ht := HousingType(housingType)
		filterType = &ht
	}

	var result []AmenityCategory
	for _, cat := range categoriesOrdered {
		var items []AmenityItem
		for _, a := range Registry {
			if a.Category != cat.ID {
				continue
			}
			if filterType != nil && !a.IsAllowedFor(*filterType) {
				continue
			}
			items = append(items, AmenityItem{ID: a.ID, Name: a.DisplayName})
		}
		if len(items) > 0 {
			result = append(result, AmenityCategory{
				ID:        cat.ID,
				Name:      cat.Name,
				Amenities: items,
			})
		}
	}
	return result
}
