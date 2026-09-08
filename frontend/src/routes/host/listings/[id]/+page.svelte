<!-- src/routes/host/listings/[id]/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { X, Save, Loader2 } from 'lucide-svelte';

	import ListingEditHeader from '$lib/components/host-listing-edit/ListingEditHeader.svelte';
	import ListingEditNav from '$lib/components/host-listing-edit/ListingEditNav.svelte';
	import ListingEditPreview from '$lib/components/host-listing-edit/ListingEditPreview.svelte';
	import ListingSaveChoiceModal from '$lib/components/host-listing-edit/ListingSaveChoiceModal.svelte';

	import EditBasics from '$lib/components/host-listing-edit/sections/EditBasics.svelte';
	import EditDescription from '$lib/components/host-listing-edit/sections/EditDescription.svelte';
	import EditLocation from '$lib/components/host-listing-edit/sections/EditLocation.svelte';
	import EditPricing from '$lib/components/host-listing-edit/sections/EditPricing.svelte';
	import EditPhotos from '$lib/components/host-listing-edit/sections/EditPhotos.svelte';
	import EditAmenities from '$lib/components/host-listing-edit/sections/EditAmenities.svelte';
	import EditRules from '$lib/components/host-listing-edit/sections/EditRules.svelte';
	import EditContacts from '$lib/components/host-listing-edit/sections/EditContacts.svelte';

	import {
		LISTING_EDIT_NAV_ITEMS,
		getNavItemById,
		getProgressForSection
	} from '$lib/components/host-listing-edit/listingEditData';
	import type {
		ListingEditSectionId,
		ListingEditFormData
	} from '$lib/components/host-listing-edit/types';

	import { hostListingsStore } from '$lib/stores/listingsStore.svelte';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { canAccessHostArea } from '$lib/auth/permissions';
	import { toast } from '$lib/stores/toastStore';
	import { listingsApi } from '$lib/api/listings';
	import { apiListingToCardListing } from '$lib/utils/listingConverters';
	import type { Listing } from '$lib/components/card/types';

	const listingId = $derived($page.params.id ?? '');

	let listing = $state<Listing | null>(null);
	let isLoading = $state(true);
	let activeSection = $state<ListingEditSectionId>('basics');
	let isDirty = $state(false);
	let isSaving = $state(false);
	let saveChoiceModalOpen = $state(false);

	let formData = $state<Partial<ListingEditFormData>>({});

	let completedSections = $derived.by<ListingEditSectionId[]>(() => {
		if (!listing) return [];
		return LISTING_EDIT_NAV_ITEMS.filter(
			(item) => getProgressForSection(item.id, listing) >= 100
		).map((item) => item.id);
	});

	function initFormData(l: Listing) {
		formData = {
			title: l.title,
			description: l.description,
			propertyType: l.propertyType,
			address: l.address,
			city: l.location?.city || '',
			latitude: l.location?.lat,
			longitude: l.location?.lng,
			area: l.area,
			floor: l.floor,
			totalFloors: l.totalFloors,
			maxGuests: l.maxGuests,
			bedrooms: l.bedrooms,
			beds: l.beds,
			bathrooms: l.bathrooms,
			pricePerNight: l.pricePerNight,
			currency: l.currency || 'BYN',
			images: [...(l.images || [])],
			amenities: [...(l.amenities || [])],
			rules: {
				checkIn: l.rules?.checkIn || '14:00',
				checkOut: l.rules?.checkOut || '12:00',
				minNights: l.rules?.minNights || 1,
				depositRequired: l.rules?.depositRequired || false,
				documentsProvided: l.rules?.documentsProvided || false,
				childrenAllowed: l.rules?.childrenAllowed ?? true,
				petsAllowed: l.rules?.petsAllowed || false,
				smokingAllowed: l.rules?.smokingAllowed || false,
				partiesAllowed: l.rules?.partiesAllowed || false
			},
			contacts: {
				phone: authStore.user?.phone || '',
				email: '',
				telegram: '',
				whatsapp: ''
			},
			isActive: l.isActive
		};
	}

	async function loadListing() {
		isLoading = true;
		try {
			if (!hostListingsStore.isInitialized) {
				await hostListingsStore.initialize();
			}

			let found = hostListingsStore.getById(listingId);
			if (!found) {
				const raw = await listingsApi.getPublicListing(listingId);
				if (raw) {
					found = apiListingToCardListing(raw);
				}
			}

			if (found) {
				listing = found;
				initFormData(found);
			} else {
				toast.error('Ошибка', 'Объявление не найдено');
				goto(resolve('/host/listings'));
			}
		} catch (err: any) {
			console.error('Failed to load listing for edit:', err);
			toast.error('Ошибка загрузки', err?.message || 'Не удалось загрузить объявление');
			goto(resolve('/host/listings'));
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		if (authStore.initialized) {
			if (!authStore.user) {
				authStore.setPendingAction(null, `/host/listings/${listingId}`);
				goto(resolve('/auth'));
				return;
			}
			if (!canAccessHostArea(authStore.user)) {
				goto(resolve('/profile'));
				return;
			}
		}
		loadListing();
	});

	function selectSection(id: ListingEditSectionId) {
		activeSection = id;
	}

	function handleFormUpdate(updates: Partial<ListingEditFormData>) {
		formData = { ...formData, ...updates };
		isDirty = true;
	}

	function handleCancelClick() {
		if (isDirty) {
			toast.confirm({
				title: 'Отменить изменения?',
				message: 'У вас есть несохранённые изменения. Если вы выйдете сейчас, все внесённые правки будут потеряны.',
				confirmText: 'Выйти без сохранения',
				cancelText: 'Продолжить редактирование',
				type: 'warning',
				onConfirm: () => {
					isDirty = false;
					goto(resolve('/host/listings'));
				}
			});
		} else {
			goto(resolve('/host/listings'));
		}
	}

	function handleSaveClick() {
		saveChoiceModalOpen = true;
	}

	function previewListing() {
		goto(resolve('/listings/[id]', { id: listingId }));
	}

	async function executeSave(mode: 'draft' | 'moderation') {
		if (!listing) return;

		isSaving = true;
		try {
			const payload = {
				title: formData.title,
				name: formData.title,
				description: formData.description,
				property_type: formData.propertyType,
				type: formData.propertyType,
				address: formData.address,
				city: formData.city,
				street: formData.street,
				house_number: formData.houseNumber,
				latitude: formData.latitude,
				longitude: formData.longitude,
				area: formData.area,
				square: formData.area,
				floor: formData.floor,
				total_floors: formData.totalFloors,
				max_guests: formData.maxGuests,
				bedrooms: formData.bedrooms,
				rooms_count: formData.bedrooms,
				beds: formData.beds,
				beds_count: formData.beds,
				bathrooms: formData.bathrooms,
				bathrooms_count: formData.bathrooms,
				price_per_night: formData.pricePerNight,
				currency: formData.currency || 'BYN',
				min_nights: formData.rules?.minNights,
				checkin_from: formData.rules?.checkIn,
				checkout_until: formData.rules?.checkOut,
				allow_children: formData.rules?.childrenAllowed,
				allow_pets: formData.rules?.petsAllowed,
				allow_smoking: formData.rules?.smokingAllowed,
				allow_parties: formData.rules?.partiesAllowed,
				deposit_required: formData.rules?.depositRequired,
				amenities: formData.amenities,
				images: formData.images,
				submit_for_moderation: mode === 'moderation',
				save_as_draft: mode === 'draft'
			};

			const updated = await listingsApi.updateListing(listingId, payload);
			const updatedCard = apiListingToCardListing(updated);

			listing = updatedCard;
			initFormData(updatedCard);
			hostListingsStore.update(listingId, updatedCard);

			isDirty = false;
			saveChoiceModalOpen = false;

			if (mode === 'moderation') {
				toast.success('Объявление отправлено на модерацию');
			} else {
				toast.success('Объявление сохранено в черновиках');
			}

			goto(resolve('/host/listings'));
		} catch (error: any) {
			console.error('Failed to save listing changes:', error);
			toast.error('Ошибка сохранения', error?.message || 'Не удалось сохранить изменения');
		} finally {
			isSaving = false;
		}
	}
</script>

<svelte:head>
	<title>{listing ? `Редактирование: ${listing.title}` : 'Редактирование объявления'} — Flickey</title>
</svelte:head>

{#if isLoading}
	<div class="flex h-screen items-center justify-center bg-white">
		<div class="flex flex-col items-center gap-3">
			<div class="h-8 w-8 animate-spin rounded-full border-2 border-zinc-900 border-t-transparent"></div>
			<p class="text-sm text-zinc-500">Загрузка объявления...</p>
		</div>
	</div>
{:else if listing}
	<div class="flex h-screen flex-col bg-white">
		<!-- Sticky Top Header with back button and submit for moderation -->
		<ListingEditHeader
			listingTitle={listing.title}
			listingStatus={listing.status}
			{isDirty}
			{isSaving}
			onBack={handleCancelClick}
			onSubmitForModeration={() => executeSave('moderation')}
		/>

		<!-- Main Split Screen Content -->
		<div class="flex flex-1 overflow-hidden">
			<!-- Sidebar (Listing Preview + Navigation) -->
			<aside class="hidden w-80 shrink-0 overflow-y-auto border-r border-zinc-200 bg-white lg:block xl:w-96">
				<div class="space-y-6 p-5">
					<!-- Listing Preview Card -->
					<ListingEditPreview {listing} onPreview={previewListing} />

					<!-- Progress Summary -->
					<div class="rounded-2xl border border-zinc-200 bg-zinc-50 px-4 py-3.5">
						<div class="flex items-center justify-between">
							<span class="text-xs font-medium text-zinc-700">Заполнено разделов</span>
							<span class="text-xs font-bold text-zinc-900">
								{completedSections.length} / {LISTING_EDIT_NAV_ITEMS.length}
							</span>
						</div>
						<div class="mt-2 h-1.5 overflow-hidden rounded-full bg-zinc-200">
							<div
								class="h-full rounded-full bg-emerald-500 transition-all duration-500"
								style={`width: ${(completedSections.length / LISTING_EDIT_NAV_ITEMS.length) * 100}%`}
							></div>
						</div>
					</div>

					<!-- Navigation -->
					<div class="border-t border-zinc-200 pt-5">
						<p class="mb-3 px-1 text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
							Разделы
						</p>
						<ListingEditNav
							items={LISTING_EDIT_NAV_ITEMS}
							{activeSection}
							{completedSections}
							onSelect={selectSection}
						/>
					</div>

					<!-- Tips -->
					<div class="rounded-2xl border border-blue-100 bg-blue-50/70 p-4">
						<p class="text-xs leading-relaxed text-blue-900">
							<strong class="font-semibold text-blue-950">Совет:</strong> Полностью заполненные
							объявления с качественными фото привлекают до 3 раз больше бронирований.
						</p>
					</div>
				</div>
			</aside>

			<!-- Main Content Area -->
			<main class="flex-1 overflow-y-auto bg-white pb-24">
				<div class="mx-auto max-w-2xl px-4 py-8 sm:px-6 lg:px-8">
					<!-- Section Title -->
					<div class="mb-8">
						<h2 class="text-xl font-bold text-zinc-900 sm:text-2xl">
							{getNavItemById(activeSection)?.title ?? ''}
						</h2>
						<p class="mt-1 text-sm text-zinc-500">
							{getNavItemById(activeSection)?.subtitle ?? ''}
						</p>
					</div>

					<!-- Form Section Content -->
					<div class="space-y-6">
						{#if activeSection === 'basics'}
							<EditBasics form={formData} onUpdate={handleFormUpdate} />
						{:else if activeSection === 'photos'}
							<EditPhotos form={formData} onUpdate={handleFormUpdate} />
						{:else if activeSection === 'description'}
							<EditDescription form={formData} onUpdate={handleFormUpdate} />
						{:else if activeSection === 'location'}
							<EditLocation form={formData} onUpdate={handleFormUpdate} />
						{:else if activeSection === 'pricing'}
							<EditPricing form={formData} onUpdate={handleFormUpdate} />
						{:else if activeSection === 'amenities'}
							<EditAmenities form={formData} onUpdate={handleFormUpdate} />
						{:else if activeSection === 'rules'}
							<EditRules form={formData} onUpdate={handleFormUpdate} />
						{:else if activeSection === 'contacts'}
							<EditContacts form={formData} onUpdate={handleFormUpdate} />
						{/if}
					</div>

					<!-- Mobile Navigation -->
					<div class="mt-10 border-t border-zinc-200 pt-6 lg:hidden">
						<p class="mb-3 text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
							Другие разделы
						</p>
						<ListingEditNav
							items={LISTING_EDIT_NAV_ITEMS}
							{activeSection}
							{completedSections}
							onSelect={selectSection}
						/>
					</div>
				</div>
			</main>
		</div>

		<!-- Floating Bottom Action Bar (slides in with animation on changes) -->
		{#if isDirty}
			<div
				class="fixed bottom-6 left-1/2 -translate-x-1/2 z-40 flex items-center gap-3 rounded-2xl border border-zinc-200/90 bg-white/95 px-4 py-3 shadow-2xl backdrop-blur-md ring-1 ring-black/5"
				transition:fly={{ y: 50, duration: 250, easing: cubicOut }}
			>
				<div class="hidden sm:flex items-center gap-2 pr-1 text-xs font-medium text-amber-700">
					<span class="h-2 w-2 rounded-full bg-amber-500 animate-pulse"></span>
					<span>Есть несохранённые изменения</span>
				</div>

				<div class="hidden sm:block h-5 w-px bg-zinc-200"></div>

				<button
					type="button"
					class="flex items-center gap-1.5 rounded-xl border border-zinc-200 bg-zinc-100/80 px-4 py-2 text-xs sm:text-sm font-semibold text-zinc-700 transition hover:border-zinc-300 hover:bg-zinc-200/80 hover:text-zinc-900 active:scale-95 cursor-pointer"
					onclick={handleCancelClick}
				>
					<X class="h-4 w-4 text-zinc-500" />
					<span>Отмена</span>
				</button>

				<button
					type="button"
					class="flex items-center gap-2 rounded-xl bg-zinc-900 px-5 py-2 text-xs sm:text-sm font-semibold text-white shadow-md transition hover:bg-zinc-800 active:scale-95 disabled:opacity-50 cursor-pointer"
					onclick={handleSaveClick}
					disabled={isSaving}
				>
					{#if isSaving}
						<Loader2 class="h-4 w-4 animate-spin" />
						<span>Сохранение...</span>
					{:else}
						<Save class="h-4 w-4" />
						<span>Сохранить</span>
					{/if}
				</button>
			</div>
		{/if}

		<!-- Save Choice Modal -->
		<ListingSaveChoiceModal
			open={saveChoiceModalOpen}
			{isSaving}
			onClose={() => {
				if (!isSaving) saveChoiceModalOpen = false;
			}}
			onSaveAsDraft={() => executeSave('draft')}
			onSubmitForModeration={() => executeSave('moderation')}
		/>
	</div>
{/if}
