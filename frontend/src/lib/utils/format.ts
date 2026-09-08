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
