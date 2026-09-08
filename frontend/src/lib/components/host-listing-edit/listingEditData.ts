// src/lib/components/host-listing-edit/listingEditData.ts
import type { ComponentType } from 'svelte';
import {
	Home,
	Images,
	AlignLeft,
	MapPin,
	Sparkles,
	Banknote,
	ClipboardList,
	Phone
} from 'lucide-svelte';
import type { ListingEditNavItem, ListingEditSectionId } from './types';
import type { Listing } from '$lib/components/card/types';

export const LISTING_EDIT_NAV_ITEMS: ListingEditNavItem[] = [
	{
		id: 'basics',
		title: 'Основное',
		subtitle: 'Название, тип, параметры',
		icon: Home,
		group: 'main'
	},
	{
		id: 'photos',
		title: 'Фотографии',
		subtitle: 'Загрузка и порядок',
		icon: Images,
		group: 'main'
	},
	{
		id: 'description',
		title: 'Описание',
		subtitle: 'Заголовок и детали',
		icon: AlignLeft,
		group: 'main'
	},
	{
		id: 'location',
		title: 'Местоположение',
		subtitle: 'Адрес и координаты',
		icon: MapPin,
		group: 'details'
	},
	{
		id: 'amenities',
		title: 'Удобства',
		subtitle: 'Что есть в жилье',
		icon: Sparkles,
		group: 'details'
	},
	{
		id: 'pricing',
		title: 'Цены',
		subtitle: 'Стоимость и валюта',
		icon: Banknote,
		group: 'details'
	},
	{
		id: 'rules',
		title: 'Правила',
		subtitle: 'Заезд, выезд, ограничения',
		icon: ClipboardList,
		group: 'settings'
	},
	{
		id: 'contacts',
		title: 'Контакты',
		subtitle: 'Как с вами связаться',
		icon: Phone,
		group: 'settings'
	}
];

const LISTING_EDIT_SECTION_IDS = new Set<ListingEditSectionId>(
	LISTING_EDIT_NAV_ITEMS.map((item) => item.id)
);

export function isListingEditSectionId(value: string): value is ListingEditSectionId {
	return LISTING_EDIT_SECTION_IDS.has(value as ListingEditSectionId);
}

export function getNavItemById(id: ListingEditSectionId): ListingEditNavItem {
	return LISTING_EDIT_NAV_ITEMS.find((item) => item.id === id) || LISTING_EDIT_NAV_ITEMS[0];
}

export function getSectionGroupTitle(group: 'main' | 'details' | 'settings'): string {
	const titles: Record<typeof group, string> = {
		main: 'Основное',
		details: 'Детали',
		settings: 'Настройки'
	};
	return titles[group];
}

export function formatListingEditBreadcrumbs(listing: Listing | null): {
	listingTitle: string;
	listingLocation: string;
} {
	if (!listing) {
		return { listingTitle: 'Новое объявление', listingLocation: '' };
	}
	return {
		listingTitle: listing.title,
		listingLocation: listing.location?.city || listing.address
	};
}

export function getProgressForSection(sectionId: ListingEditSectionId, listing: Listing | null): number {
	if (!listing) return 0;
	
	switch (sectionId) {
		case 'basics':
			return listing.title && listing.propertyType && listing.maxGuests ? 100 : 50;
		case 'photos':
			const photoCount = listing.images?.length || 0;
			return Math.min((photoCount / 5) * 100, 100);
		case 'description':
			return (listing.description?.length || 0) > 30 ? 100 : 50;
		case 'location':
			return listing.address ? 100 : 0;
		case 'amenities':
			const amenityCount = listing.amenities?.length || 0;
			return Math.min((amenityCount / 3) * 100, 100);
		case 'pricing':
			return listing.pricePerNight > 0 ? 100 : 0;
		case 'rules':
			return listing.rules?.checkIn && listing.rules?.checkOut ? 100 : 50;
		case 'contacts':
			return 100;
		default:
			return 0;
	}
}
