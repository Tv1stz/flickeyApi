export function formatBYN(value: number): string {
	return `${new Intl.NumberFormat('ru-RU').format(value)} BYN`;
}

// forms: ['гость', 'гостя', 'гостей']
export function pluralRu(n: number, forms: [string, string, string]) {
	const abs = Math.abs(n) % 100;
	const last = abs % 10;

	if (abs > 10 && abs < 20) return forms[2];
	if (last > 1 && last < 5) return forms[1];
	if (last === 1) return forms[0];
	return forms[2];
}

export function formatRelativeTime(dateStr: string): string {
	const date = new Date(dateStr);
	const now = new Date();
	const diffMs = now.getTime() - date.getTime();
	if (isNaN(date.getTime()) || diffMs < 0) return 'Только что';

	const diffSec = Math.floor(diffMs / 1000);
	const diffMin = Math.floor(diffSec / 60);
	const diffHours = Math.floor(diffMin / 60);
	const diffDays = Math.floor(diffHours / 24);

	if (diffSec < 60) return 'Только что';
	if (diffMin < 60) {
		return `${diffMin} ${pluralRu(diffMin, ['минуту', 'минуты', 'минут'])} назад`;
	}
	if (diffHours < 24) {
		return `${diffHours} ${pluralRu(diffHours, ['час', 'часа', 'часов'])} назад`;
	}
	if (diffDays === 1) {
		const timeStr = date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
		return `Вчера в ${timeStr}`;
	}
	if (diffDays < 7) {
		return `${diffDays} ${pluralRu(diffDays, ['день', 'дня', 'дней'])} назад`;
	}
	return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
}
