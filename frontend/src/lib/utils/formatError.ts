/**
 * Error formatting utility for human-readable error messages in UI toasts.
 */

const FIELD_LABELS: Record<string, string> = {
	TotalFloors: 'Количество этажей',
	Floor: 'Этаж',
	Square: 'Площадь',
	Name: 'Название объявления',
	Address: 'Адрес',
	Latitude: 'Координаты (широта)',
	Longitude: 'Координаты (долгота)',
	MaxGuests: 'Максимум гостей',
	RoomsCount: 'Количество комнат',
	BedsCount: 'Количество спальных мест',
	BathroomsCount: 'Количество ванных комнат',
	MediaIDs: 'Фотографии',
	Amenities: 'Удобства',
	PricePerNight: 'Цена за ночь',
	Currency: 'Валюта',
	MinNights: 'Минимум ночей',
	CheckinFrom: 'Время заезда',
	CheckoutUntil: 'Время выезда',
	Description: 'Описание',
	Phone: 'Телефон'
};

const ERROR_CODE_MESSAGES: Record<string, string> = {
	VALIDATION_ERROR: 'Проверьте правильность заполненных полей',
	MEDIA_NOT_OWNED: 'Одна или несколько фотографий не загружены или не принадлежат вам',
	EMPTY_AMENITIES: 'Выберите хотя бы одно удобство',
	INVALID_AMENITIES: 'Выбраны недопустимые удобства',
	STEP_NOT_ALLOWED: 'Переход к этому шагу пока недоступен',
	DRAFT_NOT_FOUND: 'Черновик не найден',
	DRAFT_ALREADY_SUBMITTED: 'Это объявление уже отправлено на публикацию',
	DRAFT_INCOMPLETE: 'Заполните все обязательные шаги перед отправкой',
	INVALID_STEP_TRANSITION: 'Невозможно перейти к выбранному шагу',
	NETWORK_ERROR: 'Ошибка сети. Проверьте подключение к интернету',
	MISSING_USER: 'Пользователь не авторизован',
	MISSING_IDEMPOTENCY_KEY: 'Ошибка запроса: отсутствует ключ идемпотентности',
	HTTP_401: 'Сессия истекла. Пожалуйста, войдите снова',
	HTTP_403: 'Недостаточно прав для выполнения действия',
	HTTP_404: 'Запрашиваемый ресурс не найден',
	HTTP_500: 'Внутренняя ошибка сервера. Попробуйте позже'
};

/**
 * Extracts and translates Go validator error strings like:
 * "Key: 'DraftStep2Request.TotalFloors' Error:Field validation for 'TotalFloors' failed on the 'max' tag"
 */
function parseValidatorErrorString(message: string): string | null {
	const match = /Field validation for '(\w+)' failed on the '(\w+)' tag/i.exec(message);
	if (!match) return null;

	const [, field, tag] = match;
	const fieldLower = field.toLowerCase();
	const tagLower = tag.toLowerCase();

	if (fieldLower.includes('totalfloor')) {
		if (tagLower === 'max') return 'Количество этажей не может превышать 150';
		if (tagLower === 'min' || tagLower === 'required') return 'Укажите этажность дома (от 1)';
	}

	if (fieldLower.includes('floor')) {
		if (tagLower === 'max') return 'Этаж не может превышать 150';
		if (tagLower === 'min' || tagLower === 'required') return 'Укажите этаж (от 1)';
	}

	if (fieldLower.includes('square')) {
		if (tagLower === 'gt') return 'Площадь должна быть больше 10 м²';
		if (tagLower === 'lt') return 'Площадь должна быть меньше 1000 м²';
		return 'Площадь должна быть от 10 до 1000 м²';
	}

	if (fieldLower.includes('name') || fieldLower.includes('title')) {
		if (tagLower === 'min') return 'Название объявления должно содержать не менее 10 символов';
		if (tagLower === 'max') return 'Название объявления не должно превышать 100 символов';
		if (tagLower === 'required') return 'Укажите название объявления';
	}

	if (fieldLower.includes('maxguest')) {
		if (tagLower === 'max') return 'Максимальное число гостей: не более 50';
		return 'Укажите количество гостей (от 1 до 50)';
	}

	if (fieldLower.includes('room')) {
		if (tagLower === 'max') return 'Количество комнат не может превышать 30';
		return 'Укажите количество комнат (от 1 до 30)';
	}

	if (fieldLower.includes('bed')) {
		if (tagLower === 'max') return 'Количество спальных мест не может превышать 30';
		return 'Укажите количество спальных мест (от 1 до 30)';
	}

	if (fieldLower.includes('bathroom')) {
		if (tagLower === 'max') return 'Количество ванных комнат не может превышать 20';
		return 'Укажите количество ванных комнат (от 1 до 20)';
	}

	if (fieldLower.includes('media')) {
		if (tagLower === 'min') return 'Загрузите минимум 5 фотографий';
		if (tagLower === 'max') return 'Можно загрузить максимум 25 фотографий';
		return 'Добавьте фотографии жилья';
	}

	if (fieldLower.includes('price')) {
		return 'Укажите стоимость за ночь (больше 0)';
	}

	if (fieldLower.includes('night')) {
		if (tagLower === 'max') return 'Минимальный срок проживания: до 30 ночей';
		return 'Минимальный срок проживания: от 1 ночи';
	}

	if (fieldLower.includes('description')) {
		if (tagLower === 'min') return 'Описание должно содержать не менее 30 символов';
		if (tagLower === 'max') return 'Описание не должно превышать 5000 символов';
		return 'Заполните описание жилья';
	}

	const label = FIELD_LABELS[field] || field;
	if (tagLower === 'required') return `Поле «${label}» обязательно для заполнения`;
	if (tagLower === 'max') return `Значение поля «${label}» превышает допустимый максимум`;
	if (tagLower === 'min') return `Значение поля «${label}» меньше допустимого минимума`;

	return `Некорректное значение поля «${label}»`;
}

/**
 * Transforms any API error or exception into a clear, user-friendly Russian message.
 */
export function formatApiError(err: unknown, fallback = 'Произошла ошибка при сохранении'): string {
	if (!err) return fallback;

	const errObj = typeof err === 'object' && err !== null ? (err as Record<string, any>) : {};
	const message = typeof errObj.message === 'string' ? errObj.message : '';
	const code = typeof errObj.code === 'string' ? errObj.code : '';

	// 1. Check for Go validator error patterns in message
	if (message) {
		const parsedValidator = parseValidatorErrorString(message);
		if (parsedValidator) return parsedValidator;

		// If message already contains Cyrillic characters, it is likely already localized
		if (/[а-яА-ЯёЁ]/.test(message)) {
			return message;
		}
	}

	// 2. Check for known application error codes
	if (code && ERROR_CODE_MESSAGES[code]) {
		return ERROR_CODE_MESSAGES[code];
	}

	// 3. Fall back to clean message or default fallback
	return message || fallback;
}
