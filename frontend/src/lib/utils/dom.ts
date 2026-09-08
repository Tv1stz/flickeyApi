// src/lib/utils/dom.ts

/**
 * Throttle функция для scroll/resize events
 */
export function throttle<T extends (...args: unknown[]) => void>(fn: T, ms: number): T {
	let lastCall = 0;
	let timeoutId: ReturnType<typeof setTimeout> | null = null;

	return ((...args: unknown[]) => {
		const now = Date.now();
		const remaining = ms - (now - lastCall);

		if (remaining <= 0) {
			if (timeoutId) {
				clearTimeout(timeoutId);
				timeoutId = null;
			}
			lastCall = now;
			fn(...args);
		} else if (!timeoutId) {
			timeoutId = setTimeout(() => {
				lastCall = Date.now();
				timeoutId = null;
				fn(...args);
			}, remaining);
		}
	}) as T;
}

/**
 * Debounce функция
 */
export function debounce<Args extends unknown[]>(
	fn: (...args: Args) => void,
	ms: number
): (...args: Args) => void {
	let timeoutId: ReturnType<typeof setTimeout> | null = null;

	return ((...args: Args) => {
		if (timeoutId) clearTimeout(timeoutId);
		timeoutId = setTimeout(() => fn(...args), ms);
	}) as (...args: Args) => void;
}

/**
 * Проверка touch устройства
 */
export function isTouchDevice(): boolean {
	if (typeof window === 'undefined') return false;
	return 'ontouchstart' in window || navigator.maxTouchPoints > 0;
}

/**
 * Проверка Safari
 */
export function isSafari(): boolean {
	if (typeof window === 'undefined') return false;
	return /^((?!chrome|android).)*safari/i.test(navigator.userAgent);
}

/**
 * Проверка iOS
 */
export function isIOS(): boolean {
	if (typeof window === 'undefined') return false;
	return /iPad|iPhone|iPod/.test(navigator.userAgent);
}

// Храним позицию скролла
let scrollPosition = 0;

/**
 * Блокировка скролла body (улучшенная версия без мигания)
 */
export function lockBodyScroll(): void {
	if (typeof document === 'undefined') return;

	scrollPosition = window.scrollY;

	// Сохраняем ширину до блокировки
	const scrollbarWidth = window.innerWidth - document.documentElement.clientWidth;

	document.body.style.overflow = 'hidden';
	document.body.style.position = 'fixed';
	document.body.style.top = `-${scrollPosition}px`;
	document.body.style.left = '0';
	document.body.style.right = '0';
	document.body.style.width = '100%';

	// Компенсируем ширину скроллбара
	if (scrollbarWidth > 0) {
		document.body.style.paddingRight = `${scrollbarWidth}px`;
	}
}

/**
 * Разблокировка скролла body (улучшенная версия без мигания)
 */
export function unlockBodyScroll(): void {
	if (typeof document === 'undefined') return;

	// Убираем стили
	document.body.style.overflow = '';
	document.body.style.position = '';
	document.body.style.top = '';
	document.body.style.left = '';
	document.body.style.right = '';
	document.body.style.width = '';
	document.body.style.paddingRight = '';

	// Восстанавливаем позицию мгновенно без анимации
	window.scrollTo({
		top: scrollPosition,
		left: 0,
		behavior: 'instant'
	});
}

/**
 * Preload изображения
 */
export function preloadImage(src: string): Promise<void> {
	return new Promise((resolve, reject) => {
		const img = new Image();
		img.onload = () => resolve();
		img.onerror = reject;
		img.src = src;
	});
}

/**
 * Preload нескольких изображений
 */
export function preloadImages(srcs: string[]): Promise<void[]> {
	return Promise.all(srcs.map(preloadImage));
}
