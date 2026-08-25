import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

/**
 * Merge class names with Tailwind CSS conflict resolution.
 */
export function cn(...inputs: ClassValue[]): string {
	return twMerge(clsx(inputs));
}

/**
 * Format BYN currency amount.
 */
export function formatCurrency(amount: number, currency: string = 'BYN'): string {
	return `${amount.toLocaleString('ru-RU')} ${currency}`;
}

/**
 * Mask phone number for public/safe display (e.g. +375 •• ••• •• 67).
 */
export function maskPhone(phone: string): string {
	if (!phone) return '';
	if (phone.length <= 6) return phone;
	const prefix = phone.slice(0, 4);
	const suffix = phone.slice(-2);
	return `${prefix} •• ••• •• ${suffix}`;
}

/**
 * Format ISO date string into Russian locale string.
 */
export function formatDate(isoString: string): string {
	if (!isoString) return '';
	try {
		return new Date(isoString).toLocaleDateString('ru-RU', {
			day: 'numeric',
			month: 'long',
			year: 'numeric'
		});
	} catch {
		return isoString;
	}
}

/**
 * Human-readable housing type translation.
 */
export function translateHousingType(type: string): string {
	switch (type) {
		case 'apartment':
			return 'Квартира';
		case 'house':
			return 'Дом / Коттедж';
		case 'manor':
			return 'Усадьба';
		default:
			return type || 'Жилье';
	}
}

/**
 * Human-readable listing status translation.
 */
export function translateListingStatus(status: string): { label: string; color: string } {
	switch (status) {
		case 'published':
			return { label: 'Опубликовано', color: 'bg-emerald-100 text-emerald-800 border-emerald-200' };
		case 'pending_review':
			return { label: 'На модерации', color: 'bg-amber-100 text-amber-800 border-amber-200' };
		case 'awaiting_company_verification':
			return { label: 'Ожидает верификации', color: 'bg-blue-100 text-blue-800 border-blue-200' };
		case 'rejected':
			return { label: 'Отклонено', color: 'bg-rose-100 text-rose-800 border-rose-200' };
		case 'draft':
			return { label: 'Черновик', color: 'bg-slate-100 text-slate-800 border-slate-200' };
		default:
			return { label: status, color: 'bg-gray-100 text-gray-800 border-gray-200' };
	}
}

/**
 * Complete authoritative Russian translation dictionary for all amenity IDs.
 */
export function translateAmenity(id: string): string {
	const map: Record<string, string> = {
		// Basic
		wifi: 'Скоростной Wi-Fi',
		heating: 'Отопление',
		air_conditioning: 'Кондиционер',
		hot_water: 'Горячая вода',
		washing_machine: 'Стиральная машина',
		dryer: 'Сушильная машина',
		tv: 'Телевизор / Smart TV',
		iron: 'Утюг и гладильная доска',
		hair_dryer: 'Фен',
		towels: 'Полотенца',
		bed_linen: 'Постельное белье',

		// Kitchen
		full_kitchen: 'Полноценная кухня',
		refrigerator: 'Холодильник',
		stove: 'Плита',
		oven: 'Духовка',
		dishwasher: 'Посудомоечная машина',
		microwave: 'Микроволновка',
		coffee_machine: 'Кофемашина',
		kettle: 'Чайник',
		toaster: 'Тостер',
		dining_area: 'Обеденная зона',

		// Bedroom and Bathroom
		extra_pillows_blankets: 'Дополнительные подушки и одеяла',
		blackout_curtains: 'Плотные шторы',
		bathtub: 'Ванна',
		shower: 'Душ',
		bidet: 'Биде',

		// Work
		workspace: 'Рабочее место',
		external_monitor: 'Внешний монитор',

		// Comfort and Leisure
		balcony_terrace: 'Балкон / терраса',
		gym: 'Тренажерный зал',
		great_view: 'Красивый вид',
		sofa_lounge: 'Диван / зона отдыха',
		board_games: 'Настольные игры',
		books: 'Книги',

		// Family
		baby_crib: 'Детская кроватка',
		high_chair: 'Стульчик для кормления',
		pets_allowed: 'Можно с питомцами',
		toys: 'Игрушки',
		baby_bath: 'Детская ванночка',

		// Safety
		smoke_detector: 'Датчик дыма',
		carbon_monoxide_detector: 'Датчик угарного газа',
		fire_extinguisher: 'Огнетушитель',
		first_aid_kit: 'Аптечка',
		safe: 'Сейф',

		// Access and Parking
		self_check_in: 'Бесконтактное заселение',
		elevator: 'Лифт',
		parking: 'Парковка',
		ev_charger: 'Зарядка для электромобилей',

		// Accessibility
		wide_entrance: 'Широкий дверной проем',
		step_free_entrance: 'Вход без ступеней',
		accessible_bathroom: 'Оборудованная ванная комната'
	};
	return map[id] || id;
}

/**
 * Format ISO date string into human-readable relative time (e.g. "только что", "5 мин. назад", "2 ч. назад", "вчера").
 */
export function formatRelativeTime(isoString: string): string {
	if (!isoString) return '';
	try {
		const date = new Date(isoString);
		const now = new Date();
		const diffSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

		if (diffSeconds < 60) {
			return 'только что';
		}
		const diffMinutes = Math.floor(diffSeconds / 60);
		if (diffMinutes < 60) {
			return `${diffMinutes} мин. назад`;
		}
		const diffHours = Math.floor(diffMinutes / 60);
		if (diffHours < 24) {
			return `${diffHours} ч. назад`;
		}
		const diffDays = Math.floor(diffHours / 24);
		if (diffDays === 1) {
			return 'вчера';
		}
		if (diffDays < 7) {
			return `${diffDays} дн. назад`;
		}
		return date.toLocaleDateString('ru-RU', {
			day: 'numeric',
			month: 'short'
		});
	} catch {
		return isoString;
	}
}

