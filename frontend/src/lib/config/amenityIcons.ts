// src/lib/config/amenityIcons.ts
import {
	Baby,
	Bath,
	Bed,
	BookOpen,
	Building2,
	Coffee,
	CookingPot,
	Dog,
	DoorOpen,
	Droplets,
	Dumbbell,
	Flame,
	Gamepad2,
	Hammer,
	HandPlatter,
	KeyRound,
	Microwave,
	MonitorDot,
	ParkingCircle,
	Refrigerator,
	ShieldAlert,
	ShieldCheck,
	Shirt,
	Snowflake,
	Sofa,
	Sparkles,
	Sun,
	Thermometer,
	TreePine,
	Tv,
	Utensils,
	Waves,
	Wifi,
	Wind,
	WashingMachine,
	Eye
} from 'lucide-svelte';

export type AmenityIconComponent = typeof Wifi;

const iconByKey: Record<string, AmenityIconComponent> = {
	// Essentials
	wifi: Wifi,
	heating: Thermometer,
	ac: Snowflake,
	'hot-water': Droplets,
	washer: WashingMachine,
	dryer: Wind,
	tv: Tv,
	iron: Shirt,
	'hair-dryer': Wind,
	towels: HandPlatter,
	'bed-linen': Bed,

	// Kitchen
	kitchen: Utensils,
	fridge: Refrigerator,
	stove: CookingPot,
	oven: CookingPot,
	dishwasher: WashingMachine,
	microwave: Microwave,
	coffee: Coffee,
	kettle: Coffee,
	toaster: Coffee,
	'dining-area': HandPlatter,

	// Work
	desk: MonitorDot,
	monitor: MonitorDot,

	// Bedroom & bathroom
	'extra-pillows': Bed,
	'blackout-curtains': Bed,
	bathtub: Bath,
	shower: Droplets,
	bidet: Bath,

	// Comfort & leisure
	balcony: DoorOpen,
	garden: TreePine,
	pool: Waves,
	'hot-tub': Bath,
	gym: Dumbbell,
	sauna: Flame,
	view: Eye,
	sofa: Sofa,
	games: Gamepad2,
	books: BookOpen,
	fireplace: Flame,

	// Family
	crib: Baby,
	'high-chair': Baby,
	pet: Dog,
	'children-toys': Baby,
	'children-books': BookOpen,
	'baby-bath': Baby,

	// Safety
	smoke: ShieldAlert,
	co: ShieldAlert,
	fire: Flame,
	'first-aid': ShieldCheck,
	lock: KeyRound,

	// Access & parking
	'self-checkin': KeyRound,
	elevator: Building2,
	parking: ParkingCircle,
	'ev-charger': ParkingCircle,

	// Outdoor
	bbq: Flame,
	outdoor: Sofa,
	'outdoor-shower': Droplets,
	'sun-loungers': Sun,

	// Accessibility
	'wide-entrance': Sun,
	'accessible-bathroom': Sun,
	'step-free': Sun,
	'pool-hoist': Sun,

	// Services
	breakfast: HandPlatter,
	'long-term': Hammer
};

export function getAmenityIcon(iconKey: string): AmenityIconComponent | null {
	return iconByKey[iconKey] ?? null;
}

export const fallbackAmenityIcon: AmenityIconComponent = Sparkles;
