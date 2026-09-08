const CYR_TO_LAT_MAP: Record<string, string> = {
	а: 'a',
	б: 'b',
	в: 'v',
	г: 'g',
	д: 'd',
	е: 'e',
	ё: 'yo',
	ж: 'zh',
	з: 'z',
	и: 'i',
	й: 'y',
	к: 'k',
	л: 'l',
	м: 'm',
	н: 'n',
	о: 'o',
	п: 'p',
	р: 'r',
	с: 's',
	т: 't',
	у: 'u',
	ф: 'f',
	х: 'kh',
	ц: 'ts',
	ч: 'ch',
	ш: 'sh',
	щ: 'shch',
	ы: 'y',
	э: 'e',
	ю: 'yu',
	я: 'ya',
	ъ: '',
	ь: ''
};

const LAT_TO_CYR_PAIRS: Array<[string, string]> = [
	['shch', 'щ'],
	['sch', 'щ'],
	['yo', 'ё'],
	['jo', 'ё'],
	['yu', 'ю'],
	['ju', 'ю'],
	['ya', 'я'],
	['ja', 'я'],
	['zh', 'ж'],
	['kh', 'х'],
	['ts', 'ц'],
	['ch', 'ч'],
	['sh', 'ш'],
	['ye', 'е'],
	['yi', 'ы']
];

const LAT_TO_CYR_MAP: Record<string, string> = {
	a: 'а',
	b: 'б',
	v: 'в',
	g: 'г',
	d: 'д',
	e: 'е',
	z: 'з',
	i: 'и',
	j: 'й',
	k: 'к',
	l: 'л',
	m: 'м',
	n: 'н',
	o: 'о',
	p: 'п',
	r: 'р',
	s: 'с',
	t: 'т',
	u: 'у',
	f: 'ф',
	h: 'х',
	c: 'с',
	y: 'й',
	q: 'к',
	w: 'в',
	x: 'кс'
};

export function normalizeForSearch(value: string): string {
	return value
		.trim()
		.toLowerCase()
		.replace(/ё/g, 'е')
		.replace(/[^a-z0-9а-яё\s-]/gi, ' ')
		.replace(/[-_]+/g, ' ')
		.replace(/\s+/g, ' ')
		.trim();
}

function transliterateToLatin(value: string): string {
	let result = '';
	for (const char of value) {
		result += CYR_TO_LAT_MAP[char] ?? char;
	}
	return result;
}

function transliterateToCyrillic(value: string): string {
	let result = value;
	for (const [latin, cyr] of LAT_TO_CYR_PAIRS) {
		result = result.replaceAll(latin, cyr);
	}
	let finalResult = '';
	for (const char of result) {
		finalResult += LAT_TO_CYR_MAP[char] ?? char;
	}
	return finalResult;
}

export function buildSearchVariants(value: string): string[] {
	const normalized = normalizeForSearch(value);
	if (!normalized) return [];
	const latin = transliterateToLatin(normalized);
	const cyrillic = transliterateToCyrillic(normalized);
	const variants = new Set([normalized, latin, cyrillic].filter(Boolean));
	return [...variants];
}

function levenshteinDistance(a: string, b: string, maxDistance: number): number {
	if (a === b) return 0;
	const lengthDiff = Math.abs(a.length - b.length);
	if (lengthDiff > maxDistance) return maxDistance + 1;

	const v0 = new Array(b.length + 1).fill(0);
	const v1 = new Array(b.length + 1).fill(0);

	for (let i = 0; i <= b.length; i += 1) v0[i] = i;

	for (let i = 0; i < a.length; i += 1) {
		v1[0] = i + 1;
		let rowMin = v1[0];
		for (let j = 0; j < b.length; j += 1) {
			const cost = a[i] === b[j] ? 0 : 1;
			v1[j + 1] = Math.min(v1[j] + 1, v0[j + 1] + 1, v0[j] + cost);
			rowMin = Math.min(rowMin, v1[j + 1]);
		}
		if (rowMin > maxDistance) return maxDistance + 1;
		for (let j = 0; j <= b.length; j += 1) v0[j] = v1[j];
	}

	return v1[b.length];
}

function fuzzyMatch(query: string, target: string): boolean {
	if (query.length < 3 || target.length < 3) return false;
	const maxDistance = query.length <= 4 ? 1 : 2;
	if (levenshteinDistance(query, target, maxDistance) <= maxDistance) return true;

	const parts = target.split(' ');
	return parts.some((part) => {
		if (part.length < 3) return false;
		return levenshteinDistance(query, part, maxDistance) <= maxDistance;
	});
}

export function matchesSearch(query: string, target: string): boolean {
	const queryVariants = buildSearchVariants(query);
	if (queryVariants.length === 0) return true;
	const targetVariants = buildSearchVariants(target);

	return queryVariants.some((q) => targetVariants.some((t) => t.includes(q) || fuzzyMatch(q, t)));
}

export function isExactNormalizedMatch(a: string, b: string): boolean {
	return normalizeForSearch(a) === normalizeForSearch(b);
}
