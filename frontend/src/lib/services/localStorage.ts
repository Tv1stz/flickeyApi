// src/lib/services/localStorage.ts

import type { ParsedAddress } from '$lib/services/geocoding';

const RECENT_KEY = 'rental_recent_addresses';
const GEOCODE_KEY = 'rental_geocode_cache';
const MAX_RECENT = 5;
const MAX_PERSISTED = 100;

// ─── Недавние адреса ──────────────────────────────────────────────────────────

export function getRecentAddresses(): ParsedAddress[] {
	try {
		const raw = localStorage.getItem(RECENT_KEY);
		return raw ? JSON.parse(raw) : [];
	} catch {
		return [];
	}
}

export function saveRecentAddress(addr: ParsedAddress) {
	try {
		const existing = getRecentAddresses();
		// Убираем дубль если уже есть
		const filtered = existing.filter((a) => a.displayName !== addr.displayName);
		// Новый адрес — в начало
		const updated = [addr, ...filtered].slice(0, MAX_RECENT);
		localStorage.setItem(RECENT_KEY, JSON.stringify(updated));
	} catch {
		// localStorage может быть недоступен (приватный режим Safari)
	}
}

export function clearRecentAddresses() {
	localStorage.removeItem(RECENT_KEY);
}

// ─── Персистентный кэш геокодинга ─────────────────────────────────────────────
// Переживает reload страницы — не нужно повторно запрашивать одни и те же точки

interface CacheEntry {
	data: ParsedAddress;
	timestamp: number;
}

const TTL_MS = 7 * 24 * 60 * 60 * 1000; // 7 дней

function loadPersistedCache(): Map<string, CacheEntry> {
	try {
		const raw = localStorage.getItem(GEOCODE_KEY);
		if (!raw) return new Map();
		const entries: [string, CacheEntry][] = JSON.parse(raw);
		// Фильтруем устаревшие
		const now = Date.now();
		const fresh = entries.filter(([, v]) => now - v.timestamp < TTL_MS);
		return new Map(fresh);
	} catch {
		return new Map();
	}
}

function savePersistedCache(cache: Map<string, CacheEntry>) {
	try {
		const entries = [...cache.entries()].slice(-MAX_PERSISTED);
		localStorage.setItem(GEOCODE_KEY, JSON.stringify(entries));
	} catch {
		// ignore storage write errors
	}
}

export const persistedGeoCache = loadPersistedCache();

export function persistGeoEntry(key: string, data: ParsedAddress) {
	persistedGeoCache.set(key, { data, timestamp: Date.now() });
	// Дебаунсим запись чтобы не писать на каждый символ
	clearTimeout(persistTimer);
	persistTimer = setTimeout(() => savePersistedCache(persistedGeoCache), 2000);
}

let persistTimer: ReturnType<typeof setTimeout>;
