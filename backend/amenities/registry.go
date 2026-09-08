// Package amenities provides the static in-memory amenity registry.
// Authoritative single source of truth for amenity IDs, categories, and Russian display names.
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

var housesAndManors = map[HousingType]struct{}{
	HousingTypeHouse: {},
	HousingTypeManor: {},
}

var apartmentsAndHouses = map[HousingType]struct{}{
	HousingTypeApartment: {},
	HousingTypeHouse:     {},
}

var apartmentsOnly = map[HousingType]struct{}{
	HousingTypeApartment: {},
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
	Aliases      []string
}

// IsAllowedFor returns true if this amenity is valid for the given housing type.
func (a AmenityDefinition) IsAllowedFor(t HousingType) bool {
	if a.AllowedTypes == nil {
		return true
	}
	_, ok := a.AllowedTypes[t]
	return ok
}

// AmenityCategory groups amenities for display.
type AmenityCategory struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Amenities []AmenityItem `json:"amenities"`
}

// AmenityItem is an amenity for API responses.
type AmenityItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var all = AllHousingTypes

// categoriesOrdered defines the display order for categories.
var categoriesOrdered = []struct{ ID, Name string }{
	{"essentials", "Основное"},
	{"kitchen", "Кухня"},
	{"bedroom_bathroom", "Спальня и ванная"},
	{"work", "Работа"},
	{"comfort", "Комфорт и досуг"},
	{"family", "Для семьи"},
	{"safety", "Безопасность"},
	{"access", "Доступ и парковка"},
	{"outdoor", "На улице"},
	{"accessibility", "Доступность"},
	{"services", "Услуги"},
}

// canonicalRegistry contains all primary amenity definitions.
var canonicalRegistry = map[string]AmenityDefinition{
	// ── 1. Essentials ────────────────────────────────────────────────────────
	"wifi":       {ID: "wifi", Category: "essentials", DisplayName: "Wi‑Fi", AllowedTypes: all},
	"heating":    {ID: "heating", Category: "essentials", DisplayName: "Отопление", AllowedTypes: all},
	"ac":         {ID: "ac", Category: "essentials", DisplayName: "Кондиционер", AllowedTypes: all, Aliases: []string{"air_conditioning"}},
	"hot_water":  {ID: "hot_water", Category: "essentials", DisplayName: "Горячая вода", AllowedTypes: all},
	"washer":     {ID: "washer", Category: "essentials", DisplayName: "Стиральная машина", AllowedTypes: all, Aliases: []string{"washing_machine"}},
	"dryer":      {ID: "dryer", Category: "essentials", DisplayName: "Сушильная машина", AllowedTypes: all},
	"tv":         {ID: "tv", Category: "essentials", DisplayName: "Телевизор", AllowedTypes: all},
	"iron":       {ID: "iron", Category: "essentials", DisplayName: "Утюг", AllowedTypes: all},
	"hair_dryer": {ID: "hair_dryer", Category: "essentials", DisplayName: "Фен", AllowedTypes: all},
	"towels":     {ID: "towels", Category: "essentials", DisplayName: "Полотенца", AllowedTypes: all},
	"bed_linen":  {ID: "bed_linen", Category: "essentials", DisplayName: "Постельное бельё", AllowedTypes: all},

	// ── 2. Kitchen ───────────────────────────────────────────────────────────
	"kitchen":        {ID: "kitchen", Category: "kitchen", DisplayName: "Полноценная кухня", AllowedTypes: all, Aliases: []string{"full_kitchen"}},
	"fridge":         {ID: "fridge", Category: "kitchen", DisplayName: "Холодильник", AllowedTypes: all, Aliases: []string{"refrigerator"}},
	"stove":          {ID: "stove", Category: "kitchen", DisplayName: "Плита", AllowedTypes: all},
	"oven":           {ID: "oven", Category: "kitchen", DisplayName: "Духовка", AllowedTypes: all},
	"dishwasher":     {ID: "dishwasher", Category: "kitchen", DisplayName: "Посудомоечная машина", AllowedTypes: all},
	"microwave":      {ID: "microwave", Category: "kitchen", DisplayName: "Микроволновка", AllowedTypes: all},
	"coffee_machine": {ID: "coffee_machine", Category: "kitchen", DisplayName: "Кофемашина", AllowedTypes: all},
	"kettle":         {ID: "kettle", Category: "kitchen", DisplayName: "Чайник", AllowedTypes: all},
	"toaster":        {ID: "toaster", Category: "kitchen", DisplayName: "Тостер", AllowedTypes: all},
	"dining_area":    {ID: "dining_area", Category: "kitchen", DisplayName: "Обеденная зона", AllowedTypes: all},

	// ── 3. Bedroom & Bathroom ────────────────────────────────────────────────
	"extra_pillows":     {ID: "extra_pillows", Category: "bedroom_bathroom", DisplayName: "Дополнительные подушки и одеяла", AllowedTypes: all, Aliases: []string{"extra_pillows_blankets"}},
	"blackout_curtains": {ID: "blackout_curtains", Category: "bedroom_bathroom", DisplayName: "Блэкаут-шторы", AllowedTypes: all},
	"bathtub":           {ID: "bathtub", Category: "bedroom_bathroom", DisplayName: "Ванна", AllowedTypes: all},
	"shower":            {ID: "shower", Category: "bedroom_bathroom", DisplayName: "Душ", AllowedTypes: all},
	"bidet":             {ID: "bidet", Category: "bedroom_bathroom", DisplayName: "Биде", AllowedTypes: all},

	// ── 4. Work ──────────────────────────────────────────────────────────────
	"workspace": {ID: "workspace", Category: "work", DisplayName: "Рабочее место", AllowedTypes: all},
	"monitor":   {ID: "monitor", Category: "work", DisplayName: "Внешний монитор", AllowedTypes: all, Aliases: []string{"external_monitor"}},

	// ── 5. Comfort & Leisure ─────────────────────────────────────────────────
	"balcony":   {ID: "balcony", Category: "comfort", DisplayName: "Балкон / терраса", AllowedTypes: apartmentsAndHouses, Aliases: []string{"balcony_terrace"}},
	"garden":    {ID: "garden", Category: "comfort", DisplayName: "Сад / двор", AllowedTypes: housesAndManors},
	"pool":      {ID: "pool", Category: "comfort", DisplayName: "Бассейн", AllowedTypes: housesAndManors},
	"hot_tub":   {ID: "hot_tub", Category: "comfort", DisplayName: "Джакузи / горячая ванна", AllowedTypes: housesAndManors},
	"gym":       {ID: "gym", Category: "comfort", DisplayName: "Спортзал", AllowedTypes: all},
	"sauna":     {ID: "sauna", Category: "comfort", DisplayName: "Сауна", AllowedTypes: housesAndManors},
	"fireplace": {ID: "fireplace", Category: "comfort", DisplayName: "Камин", AllowedTypes: housesAndManors},
	"view":      {ID: "view", Category: "comfort", DisplayName: "Красивый вид", AllowedTypes: all, Aliases: []string{"great_view"}},
	"sofa":      {ID: "sofa", Category: "comfort", DisplayName: "Диван / зона отдыха", AllowedTypes: all, Aliases: []string{"sofa_lounge"}},
	"games":     {ID: "games", Category: "comfort", DisplayName: "Настольные игры", AllowedTypes: all, Aliases: []string{"board_games"}},
	"books":     {ID: "books", Category: "comfort", DisplayName: "Книги", AllowedTypes: all},

	// ── 6. Family ────────────────────────────────────────────────────────────
	"crib":          {ID: "crib", Category: "family", DisplayName: "Детская кроватка", AllowedTypes: all, Aliases: []string{"baby_crib"}},
	"high_chair":    {ID: "high_chair", Category: "family", DisplayName: "Детский стульчик", AllowedTypes: all},
	"pets_allowed":  {ID: "pets_allowed", Category: "family", DisplayName: "Можно с животными", AllowedTypes: all},
	"children_toys": {ID: "children_toys", Category: "family", DisplayName: "Игрушки", AllowedTypes: all, Aliases: []string{"toys"}},
	"baby_bath":     {ID: "baby_bath", Category: "family", DisplayName: "Детская ванночка", AllowedTypes: all},

	// ── 7. Safety ────────────────────────────────────────────────────────────
	"smoke_alarm":       {ID: "smoke_alarm", Category: "safety", DisplayName: "Датчик дыма", AllowedTypes: all, Aliases: []string{"smoke_detector"}},
	"co_alarm":          {ID: "co_alarm", Category: "safety", DisplayName: "Датчик угарного газа", AllowedTypes: all, Aliases: []string{"carbon_monoxide_detector"}},
	"fire_extinguisher": {ID: "fire_extinguisher", Category: "safety", DisplayName: "Огнетушитель", AllowedTypes: all},
	"first_aid":         {ID: "first_aid", Category: "safety", DisplayName: "Аптечка", AllowedTypes: all, Aliases: []string{"first_aid_kit"}},
	"lockbox":           {ID: "lockbox", Category: "safety", DisplayName: "Кодовый замок / сейф", AllowedTypes: all, Aliases: []string{"safe"}},

	// ── 8. Access & Parking ──────────────────────────────────────────────────
	"self_checkin": {ID: "self_checkin", Category: "access", DisplayName: "Самостоятельный заезд", AllowedTypes: all, Aliases: []string{"self_check_in"}},
	"elevator":     {ID: "elevator", Category: "access", DisplayName: "Лифт", AllowedTypes: apartmentsOnly},
	"parking":      {ID: "parking", Category: "access", DisplayName: "Парковка", AllowedTypes: all},
	"ev_charger":   {ID: "ev_charger", Category: "access", DisplayName: "Зарядка для электромобиля", AllowedTypes: all},

	// ── 9. Outdoor ───────────────────────────────────────────────────────────
	"bbq":             {ID: "bbq", Category: "outdoor", DisplayName: "Мангал / барбекю", AllowedTypes: housesAndManors},
	"outdoor_seating": {ID: "outdoor_seating", Category: "outdoor", DisplayName: "Зона отдыха на улице", AllowedTypes: housesAndManors},
	"outdoor_shower":  {ID: "outdoor_shower", Category: "outdoor", DisplayName: "Уличный душ", AllowedTypes: housesAndManors},
	"sun_loungers":    {ID: "sun_loungers", Category: "outdoor", DisplayName: "Шезлонги", AllowedTypes: housesAndManors},

	// ── 10. Accessibility ────────────────────────────────────────────────────
	"wide_entrance":       {ID: "wide_entrance", Category: "accessibility", DisplayName: "Широкий вход (для коляски / кресла)", AllowedTypes: all},
	"step_free":           {ID: "step_free", Category: "accessibility", DisplayName: "Вход без ступенек", AllowedTypes: all, Aliases: []string{"step_free_entrance"}},
	"accessible_bathroom": {ID: "accessible_bathroom", Category: "accessibility", DisplayName: "Ванная для людей с ОВЗ", AllowedTypes: all},
	"pool_hoist":          {ID: "pool_hoist", Category: "accessibility", DisplayName: "Подъёмник для бассейна", AllowedTypes: housesAndManors},

	// ── 11. Services ─────────────────────────────────────────────────────────
	"breakfast":       {ID: "breakfast", Category: "services", DisplayName: "Завтрак включён", AllowedTypes: all},
	"long_term_stays": {ID: "long_term_stays", Category: "services", DisplayName: "Длительное проживание (от месяца)", AllowedTypes: all},
}

// Registry is the full lookup map including canonical IDs and alias mappings.
var Registry = func() map[string]AmenityDefinition {
	m := make(map[string]AmenityDefinition, len(canonicalRegistry)*2)
	for id, def := range canonicalRegistry {
		m[id] = def
		for _, alias := range def.Aliases {
			m[alias] = def
		}
	}
	return m
}()

// GetAmenity looks up an amenity by canonical ID or alias.
func GetAmenity(id string) (AmenityDefinition, bool) {
	a, ok := Registry[id]
	return a, ok
}

// GetGroupedAmenities returns canonical amenities grouped by category, optionally filtered by housing type.
func GetGroupedAmenities(housingType string) []AmenityCategory {
	var filterType *HousingType
	if housingType != "" {
		ht := HousingType(housingType)
		filterType = &ht
	}

	var result []AmenityCategory
	for _, cat := range categoriesOrdered {
		var items []AmenityItem
		for _, a := range canonicalRegistry {
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
