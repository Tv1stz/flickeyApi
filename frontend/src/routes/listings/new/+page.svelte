<!-- src/routes/listings/new/+page.svelte -->
<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { beforeNavigate, goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import { fade } from 'svelte/transition';
	import { pluralRu } from '$lib/utils/format';

	import ListingStepsHeader from '$lib/components/listing/ListingStepsHeader.svelte';
	import ListingWelcomeScreen from '$lib/components/listing/ListingWelcomeScreen.svelte';
	import Step1PropertyType from '$lib/components/listing/steps/Step1PropertyType.svelte';
	import Step2Location from '$lib/components/listing/steps/Step2Location.svelte';
	import Step3MainParams from '$lib/components/listing/steps/Step3MainParams.svelte';
	import Step4Photos from '$lib/components/listing/steps/Step4Photos.svelte';
	import Step5Amenities from '$lib/components/listing/steps/Step5Amenities.svelte';
	import Step6PriceRules from '$lib/components/listing/steps/Step6PriceRules.svelte';
	import Step7Description from '$lib/components/listing/steps/Step7Description.svelte';
	import Step8Contacts from '$lib/components/listing/steps/Step8Contacts.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	import {
		validateListingField,
		validateListingForm,
		validatedListingFields,
		type ListingFormValues,
		type ListingErrors,
		type ListingTouched,
		type ValidatedListingField,
		type ContactMethod
	} from '$lib/validation/listingValidation';

	import {
		getAmenityCategoriesForPropertyType,
		getAmenityIdsForPropertyType,
		getAmenityPresetOptions,
		normalizeAmenityIds,
		type AmenityPresetId
	} from '$lib/config/amenities';
	import { debounce } from '$lib/utils/dom';
	import { saveDraft, loadDraft, clearDraft } from '$lib/utils/formStorage';
	import { toast } from '$lib/stores/toastStore';
	import { X, ChevronLeft, ChevronRight } from 'lucide-svelte';
	import { formToCreateRequest, formToListingPreview } from '$lib/utils/listingConverters';
	import { resolveListingLocation } from '$lib/services/geocodingService';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { hostListingsStore } from '$lib/stores/listingsStore.svelte';
	import { locationStore } from '$lib/stores/location.svelte';
	import { isHostUser } from '$lib/auth/permissions';
	import type { PropertyType, CreateListingRequest } from '$lib/components/card/types';
	import { listingsApi } from '$lib/api/listings';
	import { mediaApi } from '$lib/api/media';
	import { formatApiError } from '$lib/utils/formatError';
	import type { HousingType } from '$lib/types/listings';

	// ═══════════════════════════════════════════════════════════════════
	// CONSTANTS
	// ═══════════════════════════════════════════════════════════════════

	const STEPS = [
		{ id: 1, label: 'Тип жилья' },
		{ id: 2, label: 'Локация' },
		{ id: 3, label: 'Основные параметры' },
		{ id: 4, label: 'Фото' },
		{ id: 5, label: 'Удобства' },
		{ id: 6, label: 'Цена и правила' },
		{ id: 7, label: 'Описание' },
		{ id: 8, label: 'Контакты' }
	] as const;

	const STEP_FIELDS: Record<number, ValidatedListingField[]> = {
		1: ['propertyType'],
		2: ['address'],
		3: ['title', 'size', 'floor', 'totalFloors', 'maxGuests', 'bedrooms', 'beds', 'bathrooms'],
		4: ['photos'],
		5: [],
		6: ['basePrice', 'minNights', 'checkInFrom', 'checkOutBefore'],
		7: ['description'],
		8: ['phone', 'contactMethods']
	};

	const ALWAYS_REQUIRED_FIELDS = new Set<ValidatedListingField>([
		'title',
		'address',
		'propertyType',
		'size',
		'totalFloors',
		'maxGuests',
		'bedrooms',
		'beds',
		'bathrooms',
		'basePrice',
		'minNights',
		'checkInFrom',
		'checkOutBefore',
		'description',
		'phone',
		'photos'
	]);

	const CONFIG = {
		MAX_PHOTOS: 25,
		AUTOSAVE_DEBOUNCE_MS: 2000
	} as const;
	const MIN_AMENITIES_REQUIRED = 1;

	// ═══════════════════════════════════════════════════════════════════
	// TYPES
	// ═══════════════════════════════════════════════════════════════════

	interface PhotoItem {
		id: string;
		mediaId?: string;
		file?: File;
		url: string;
		isCover: boolean;
	}

	// ═══════════════════════════════════════════════════════════════════
	// STATE
	// ═══════════════════════════════════════════════════════════════════

	let currentStep = $state(1);
	let previousStep = $state(1);
	let submitStatus = $state('');
	let isSubmitting = $state(false);
	let navigationConfirmed = $state(false);
	let ownerName = $state('Вы');
	let currentUserId = $state<string | null>(null);
	let guardChecked = $state(false);
	let guardRedirected = $state(false);
	let hostActivationRequested = $state(false);
	let canUseDrafts = $state(false);
	let shouldResumeDraft = $state(false);
	const hasInitialDraftParam =
		typeof window !== 'undefined'
			? new URLSearchParams(window.location.search).has('draft_id')
			: false;
	let showWelcomeScreen = $state(!hasInitialDraftParam);
	let restoredDraftForUser = $state<string | null>(null);
	let hasRestoredDraft = $state(false);
	let amenitiesError = $state('');
	let draftId = $state<string | null>(null);
	let isEditMode = $state<boolean>(false);

	let form = $state<ListingFormValues>(createInitialFormState());
	let touched = $state<ListingTouched>(createInitialTouchedState());
	let errors = $state<ListingErrors>(createInitialErrorsState());
	let photoItems = $state<PhotoItem[]>([]);
	let draftBaselineSnapshot = $state('');

	const amenityPropertyType = $derived((form.propertyType as PropertyType) || 'apartment');
	const amenityCategories = $derived(getAmenityCategoriesForPropertyType(amenityPropertyType));
	const amenityPresetOptions = $derived(getAmenityPresetOptions(amenityPropertyType));

	// ═══════════════════════════════════════════════════════════════════
	// INITIALIZATION
	// ═══════════════════════════════════════════════════════════════════

	function createInitialFormState(): ListingFormValues {
		return {
			title: '',
			address: '',
			propertyType: 'apartment',
			size: '',
			floor: '',
			totalFloors: '',
			maxGuests: '',
			bedrooms: '',
			beds: '',
			bathrooms: '',
			basePrice: '',
			minNights: '',
			checkInFrom: '14:00',
			checkOutBefore: '12:00',
			description: '',
			photos: [],
			amenities: [],
			allowChildren: false,
			allowPets: false,
			smokingAllowed: false,
			partiesAllowed: false,
			securityDepositRequired: false,
			reportingDocuments: false,
			phone: '',
			contactMethods: []
		};
	}

	function createInitialTouchedState(): ListingTouched {
		return validatedListingFields.reduce((acc, key) => {
			acc[key] = false;
			return acc;
		}, {} as ListingTouched);
	}

	function createInitialErrorsState(): ListingErrors {
		return validatedListingFields.reduce((acc, key) => {
			acc[key] = '';
			return acc;
		}, {} as ListingErrors);
	}

	function createDraftSnapshot(values: ListingFormValues, photos: PhotoItem[]): string {
		const normalizedContactMethods = values.contactMethods
			.map((method) => ({ type: method.type, value: method.value.trim() }))
			.filter((method) => method.value.length > 0)
			.sort((a, b) => a.type.localeCompare(b.type));

		const normalizedAmenities = normalizeAmenityIds(values.amenities).sort();
		const coverIndex = photos.findIndex((photo) => photo.isCover);

		return JSON.stringify({
			title: values.title.trim(),
			address: values.address.trim(),
			propertyType: values.propertyType,
			size: values.size.trim(),
			floor: values.floor.trim(),
			totalFloors: values.totalFloors.trim(),
			maxGuests: values.maxGuests.trim(),
			bedrooms: values.bedrooms.trim(),
			beds: values.beds.trim(),
			bathrooms: values.bathrooms.trim(),
			basePrice: values.basePrice.trim(),
			minNights: values.minNights.trim(),
			checkInFrom: values.checkInFrom,
			checkOutBefore: values.checkOutBefore,
			description: values.description.trim(),
			amenities: normalizedAmenities,
			allowChildren: values.allowChildren,
			allowPets: values.allowPets,
			smokingAllowed: values.smokingAllowed,
			partiesAllowed: values.partiesAllowed,
			securityDepositRequired: values.securityDepositRequired,
			reportingDocuments: values.reportingDocuments,
			phone: values.phone.trim(),
			contactMethods: normalizedContactMethods,
			photoCount: photos.length,
			coverIndex: coverIndex >= 0 ? coverIndex : null
		});
	}

	function resetDraftBaselineSnapshot(): void {
		draftBaselineSnapshot = createDraftSnapshot(form, photoItems);
	}

	resetDraftBaselineSnapshot();

	// ═══════════════════════════════════════════════════════════════════
	// DRAFT MANAGEMENT & SERVER SYNC
	// ═══════════════════════════════════════════════════════════════════

	$effect(() => {
		const unsubscribe = page.subscribe(($page) => {
			const navigationState = $page.state as { resumeDraft?: boolean } | null;
			shouldResumeDraft =
				$page.url.searchParams.get('resumeDraft') === '1' || Boolean(navigationState?.resumeDraft);
		});
		return unsubscribe;
	});

	$effect(() => {
		if (shouldResumeDraft) {
			showWelcomeScreen = false;
		}
	});

	$effect(() => {
		currentUserId = authStore.user?.id ?? null;
		const hasPublishedListings = (authStore.user?.hostProfile?.listingsCount || 0) > 0;
		canUseDrafts = hasPublishedListings;
		ownerName = authStore.user?.name || 'Вы';
		if (authStore.user?.phone && !form.phone.trim()) {
			form.phone = authStore.user.phone;
			resetDraftBaselineSnapshot();
		}
	});

	onMount(async () => {
		authStore.initialize();

		const ensureCreateListingAccess = () => {
			if (guardChecked || guardRedirected) return;
			if (!authStore.initialized) return;

			if (!authStore.user) {
				guardRedirected = true;
				authStore.setPendingAction('create-listing', '/listings/new');
				goto(resolve('/auth'));
				return;
			}

			if (!isHostUser(authStore.user)) {
				if (!hostActivationRequested) {
					hostActivationRequested = true;
					authStore.becomeHost();
					toast.success('Роль хозяина активирована', 'Заполните первое объявление');
				}
				return;
			}

			guardChecked = true;
		};

		ensureCreateListingAccess();

		// Check for server draft ID from query param
		const urlDraftId = $page.url.searchParams.get('draft_id');
		if (urlDraftId) {
			await loadServerDraft(urlDraftId);
		} else {
			const savedDraftId = localStorage.getItem('flickey_current_draft_id');
			if (savedDraftId) {
				hasRestoredDraft = true;
			}
		}
	});

	async function loadServerDraft(id: string) {
		try {
			const d = await listingsApi.getDraft(id);
			draftId = d.id;
			localStorage.setItem('flickey_current_draft_id', d.id);
			isEditMode = d.mode === 'edit';
			showWelcomeScreen = false;

			if (d.type) {
				form.propertyType = d.type === 'manor' ? 'estate' : (d.type as PropertyType);
			}
			if (d.name) form.title = d.name;
			if (d.address) form.address = d.address;
			if (d.square) form.size = String(d.square);
			if (d.floor) form.floor = String(d.floor);
			if (d.total_floors) form.totalFloors = String(d.total_floors);
			if (d.max_guests) form.maxGuests = String(d.max_guests);
			if (d.rooms_count) form.bedrooms = String(d.rooms_count);
			if (d.beds_count) form.beds = String(d.beds_count);
			if (d.bathrooms_count) form.bathrooms = String(d.bathrooms_count);
			if (d.amenities) form.amenities = d.amenities;
			if (d.price_per_night) form.basePrice = String(d.price_per_night);
			if (d.min_nights) form.minNights = String(d.min_nights);
			if (d.checkin_from) form.checkInFrom = d.checkin_from.slice(0, 5);
			if (d.checkout_until) form.checkOutBefore = d.checkout_until.slice(0, 5);
			if (d.allow_children !== null && d.allow_children !== undefined) form.allowChildren = d.allow_children;
			if (d.allow_pets !== null && d.allow_pets !== undefined) form.allowPets = d.allow_pets;
			if (d.allow_smoking !== null && d.allow_smoking !== undefined) form.smokingAllowed = d.allow_smoking;
			if (d.allow_parties !== null && d.allow_parties !== undefined) form.partiesAllowed = d.allow_parties;
			if (d.deposit_required !== null && d.deposit_required !== undefined) form.securityDepositRequired = d.deposit_required;
			if (d.with_invoicing !== null && d.with_invoicing !== undefined) form.reportingDocuments = d.with_invoicing;
			if (d.description) form.description = d.description;

			if (d.latitude && d.longitude) {
				locationStore.setCoords(d.latitude, d.longitude);
			}

			// Load photos from draft
			if (d.media && d.media.length > 0) {
				photoItems = d.media.map((m, index) => ({
					id: m.id,
					mediaId: m.id,
					url: m.url,
					isCover: index === 0
				}));
				syncPhotosToForm();
			} else if (d.media_ids && d.media_ids.length > 0) {
				photoItems = d.media_ids.map((mediaId, index) => ({
					id: mediaId,
					mediaId,
					url: `http://localhost:8000/api/v1/media/dev-upload/${mediaId}`,
					isCover: index === 0
				}));
				syncPhotosToForm();
			}

			// Step resolution:
			const requestedStep = Number($page.url.searchParams.get('step'));
			const stepMapping: Record<number, number> = {
				1: 1,
				2: 2,
				3: 4,
				4: 5,
				5: 6,
				6: 7,
				7: 8
			};
			if (requestedStep >= 1 && requestedStep <= STEPS.length) {
				currentStep = requestedStep;
				previousStep = currentStep;
			} else if (isEditMode) {
				currentStep = 1;
				previousStep = 1;
			} else if (d.current_step && stepMapping[d.current_step]) {
				currentStep = stepMapping[d.current_step];
				previousStep = currentStep;
			}

			toast.success('Черновик загружен', isEditMode ? 'Режим редактирования' : 'Продолжайте заполнение');
			resetDraftBaselineSnapshot();
		} catch (err) {
			localStorage.removeItem('flickey_current_draft_id');
		}
	}

	function persistDraftSnapshot(): void {
		if (!canUseDrafts || !isFormDirty) return;

		saveDraft(currentUserId, form, {
			step: currentStep,
			photoCount: photoItems.length,
			coverIndex: Math.max(
				0,
				photoItems.findIndex((p) => p.isCover)
			)
		});
	}

	const debouncedSaveDraft = debounce((formData: ListingFormValues) => {
		if (!canUseDrafts || !isFormDirty) return;

		saveDraft(currentUserId, formData, {
			step: currentStep,
			photoCount: photoItems.length,
			coverIndex: Math.max(
				0,
				photoItems.findIndex((p) => p.isCover)
			)
		});
	}, CONFIG.AUTOSAVE_DEBOUNCE_MS);

	const currentDraftSnapshot = $derived(createDraftSnapshot(form, photoItems));
	const isFormDirty = $derived(currentDraftSnapshot !== draftBaselineSnapshot);

	$effect(() => {
		debouncedSaveDraft(form);
		void currentDraftSnapshot;
	});

	$effect(() => {
		const next = sanitizeAmenitiesForPropertyType(amenityPropertyType);
		const changed =
			next.length !== form.amenities.length ||
			next.some((id, index) => id !== form.amenities[index]);

		if (changed) {
			const droppedCount = form.amenities.length - next.length;
			form.amenities = next;

			if (droppedCount > 0) {
				toast.info(
					'Список удобств обновлён',
					`${droppedCount} ${pluralRu(droppedCount, ['удобство недоступно', 'удобства недоступны', 'удобств недоступны'])} для этого типа жилья`
				);
			}

			if (amenitiesError) validateAmenitiesSelection();
		}
	});

	// ═══════════════════════════════════════════════════════════════════
	// UNSAVED CHANGES & NAVIGATION GUARD
	// ═══════════════════════════════════════════════════════════════════

	const hasUnsavedChanges = $derived(isFormDirty);

	beforeNavigate(({ cancel, to }) => {
		if (navigationConfirmed) {
			navigationConfirmed = false;
			return;
		}
		if (hasUnsavedChanges) {
			cancel();
			toast.confirm({
				title: 'Покинуть страницу?',
				message: canUseDrafts
					? 'У вас есть несохранённые изменения. Черновик будет сохранён автоматически.'
					: 'У вас есть несохранённые изменения.',
				confirmText: 'Покинуть',
				cancelText: 'Остаться',
				type: 'warning',
				onConfirm: () => {
					persistDraftSnapshot();
					navigationConfirmed = true;
					if (to?.url) goto(resolve((to.url.pathname + to.url.search) as '/'));
					else goto(resolve('/'));
				}
			});
		}
	});

	function handleBeforeUnload(e: BeforeUnloadEvent) {
		if (hasUnsavedChanges) {
			persistDraftSnapshot();
			e.preventDefault();
			return '';
		}
	}

	// ═══════════════════════════════════════════════════════════════════
	// HELPERS
	// ═══════════════════════════════════════════════════════════════════

	function generateId(): string {
		return crypto?.randomUUID?.() ?? Math.random().toString(36).slice(2);
	}

	function scrollToTop(): void {
		if (typeof window !== 'undefined') window.scrollTo({ top: 0, behavior: 'smooth' });
	}

	function getExitPath(): '/host/listings' | '/' {
		return authStore.viewMode === 'host' ? '/host/listings' : '/';
	}

	function handleExitCreation(): void {
		const exitPath = getExitPath();

		if (!hasUnsavedChanges) {
			navigationConfirmed = true;
			goto(resolve(exitPath));
			return;
		}

		toast.confirm({
			title: 'Выйти из создания объявления?',
			message: canUseDrafts
				? 'Черновик сохранится автоматически, вы сможете вернуться позже.'
				: 'У вас есть несохранённые изменения. Выйти без сохранения?',
			confirmText: 'Выйти',
			cancelText: 'Остаться',
			type: 'warning',
			onConfirm: () => {
				persistDraftSnapshot();
				navigationConfirmed = true;
				goto(resolve(exitPath));
			}
		});
	}

	async function startCreationFlow(): Promise<void> {
		showWelcomeScreen = false;
		scrollToTop();
		if (hasRestoredDraft) {
			const savedDraftId = localStorage.getItem('flickey_current_draft_id');
			if (savedDraftId && !draftId) {
				await loadServerDraft(savedDraftId);
			}
		}
	}

	function isFieldRequired(name: ValidatedListingField): boolean {
		if (name === 'floor') return form.propertyType === 'apartment';
		if (name === 'contactMethods') return form.contactMethods.some((method) => method.value.trim());
		return ALWAYS_REQUIRED_FIELDS.has(name);
	}

	function isFieldApplicable(name: ValidatedListingField): boolean {
		if (name === 'floor') return form.propertyType === 'apartment';
		return true;
	}

	function sanitizeAmenitiesForPropertyType(propertyType: PropertyType): string[] {
		const allowedIds = new Set(getAmenityIdsForPropertyType(propertyType));
		return normalizeAmenityIds(form.amenities).filter((id) => allowedIds.has(id));
	}

	function validateAmenitiesSelection(): boolean {
		const valid = form.amenities.length >= MIN_AMENITIES_REQUIRED;
		amenitiesError = valid ? '' : 'Выберите хотя бы одно удобство';
		return valid;
	}

	// ═══════════════════════════════════════════════════════════════════
	// VALIDATION
	// ═══════════════════════════════════════════════════════════════════

	function validateStep(step: number): boolean {
		if (step === 5) {
			return validateAmenitiesSelection();
		}

		const fields = STEP_FIELDS[step] ?? [];
		if (!fields.length) return true;
		let hasErrors = false;
		for (const name of fields) {
			if (!isFieldApplicable(name)) {
				errors[name] = '';
				continue;
			}
			if (!isFieldRequired(name)) {
				errors[name] = '';
				continue;
			}
			errors[name] = validateListingField(name, form);
			touched[name] = true;
			if (errors[name]) hasErrors = true;
		}
		return !hasErrors;
	}

	function validateAllFields(): boolean {
		const newErrors = validateListingForm(form);
		let hasErrors = false;
		for (const key of validatedListingFields) {
			if (!isFieldApplicable(key) || !isFieldRequired(key)) {
				errors[key] = '';
				continue;
			}
			errors[key] = newErrors[key] || '';
			if (newErrors[key]) {
				touched[key] = true;
				hasErrors = true;
			}
		}

		if (!validateAmenitiesSelection()) {
			hasErrors = true;
		}

		return !hasErrors;
	}

	function focusFirstInvalidField(fields: ValidatedListingField[]): void {
		if (typeof document === 'undefined') return;
		const firstInvalid = fields.find((f) => isFieldApplicable(f) && Boolean(errors[f]));
		if (!firstInvalid) return;
		const element = document.getElementById(firstInvalid);
		element?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		element?.focus();
	}

	// ═══════════════════════════════════════════════════════════════════
	// FORM HANDLERS
	// ═══════════════════════════════════════════════════════════════════

	function handleBlur(name: ValidatedListingField): void {
		touched[name] = true;
		if (!isFieldRequired(name)) {
			errors[name] = '';
			return;
		}
		errors[name] = validateListingField(name, form);
	}

	type StringInputField = Exclude<ValidatedListingField, 'photos' | 'contactMethods'>;

	function handleInput(name: StringInputField, value: string): void {
		(form as Record<StringInputField, string>)[name] = value;
		if (touched[name]) {
			errors[name] = isFieldRequired(name) ? validateListingField(name, form) : '';
		}
		if (name === 'floor' && touched.totalFloors) {
			errors.totalFloors = validateListingField('totalFloors', form);
		}
		if (name === 'totalFloors' && touched.floor && form.propertyType === 'apartment') {
			errors.floor = validateListingField('floor', form);
		}
		if (name === 'propertyType') {
			form.floor = '';
			form.totalFloors = '';
			touched.floor = false;
			touched.totalFloors = false;
			errors.floor = '';
			errors.totalFloors = '';
			form.amenities = sanitizeAmenitiesForPropertyType((value as PropertyType) || 'apartment');
			if (currentStep >= 5 || amenitiesError) {
				validateAmenitiesSelection();
			}
		}
	}

	function handleUpdateContactMethods(methods: ContactMethod[]): void {
		form.contactMethods = methods;
		if (touched.contactMethods && isFieldRequired('contactMethods')) {
			errors.contactMethods = validateListingField('contactMethods', form);
		} else {
			errors.contactMethods = '';
		}
	}

	function toggleBoolean(
		field:
			| 'allowChildren'
			| 'allowPets'
			| 'smokingAllowed'
			| 'partiesAllowed'
			| 'securityDepositRequired'
			| 'reportingDocuments',
		value: boolean
	): void {
		form[field] = value;
	}

	function toggleAmenity(id: string): void {
		const index = form.amenities.indexOf(id);
		form.amenities =
			index === -1 ? [...form.amenities, id] : form.amenities.filter((x) => x !== id);
		if (amenitiesError) validateAmenitiesSelection();
	}

	function resetAmenities(): void {
		form.amenities = [];
		validateAmenitiesSelection();
	}

	function applyAmenityPreset(presetId: AmenityPresetId): void {
		const preset = amenityPresetOptions.find((option) => option.id === presetId);
		if (!preset) return;

		const before = form.amenities.length;
		const nextSet = new Set([...form.amenities, ...preset.amenityIds]);
		form.amenities = Array.from(nextSet);

		const added = form.amenities.length - before;
		if (added > 0) {
			toast.success('Подборка добавлена', `Добавлено ${added} удобств`);
		} else {
			toast.info('Все удобства из этой подборки уже выбраны');
		}

		if (amenitiesError) validateAmenitiesSelection();
	}

	// ═══════════════════════════════════════════════════════════════════
	// STEP NAVIGATION & SERVER SYNC
	// ═══════════════════════════════════════════════════════════════════

	function handlePrev(): void {
		if (currentStep <= 1) return;
		previousStep = currentStep;
		currentStep -= 1;
		submitStatus = '';
		scrollToTop();
	}

	async function saveStepData(step: number): Promise<void> {
		if (step === 1) {
			const housingType: HousingType =
				form.propertyType === 'estate' ? 'manor' : (form.propertyType as HousingType);
			if (!draftId) {
				const res = await listingsApi.createDraft({ type: housingType });
				draftId = res.draft_id;
				localStorage.setItem('flickey_current_draft_id', res.draft_id);
			}
		} else if (step === 2) {
			// Step 2 is Location. Backend Step 2 bundles Location + Main Params (Step 3).
			// If already filled (e.g. in edit mode), save to server. Otherwise, it will be saved on Step 3.
			if (!draftId) return;
			if (form.title.trim().length >= 10 && Number(form.size) > 10) {
				const lat = locationStore.lat || 53.9006;
				const lng = locationStore.lng || 27.5590;
				await listingsApi.updateStep2(draftId, {
					name: form.title.trim(),
					address: form.address.trim(),
					city: locationStore.address?.city || '',
					street: locationStore.address?.street || '',
					house_number: locationStore.address?.house || '',
					latitude: lat,
					longitude: lng,
					square: Number(form.size) || 1,
					floor: Number(form.floor) || 1,
					total_floors: Number(form.totalFloors) || 1,
					max_guests: Number(form.maxGuests) || 1,
					rooms_count: Number(form.bedrooms) || 1,
					beds_count: Number(form.beds) || 1,
					bathrooms_count: Number(form.bathrooms) || 1
				});
			}
		} else if (step === 3) {
			if (!draftId) return;
			const lat = locationStore.lat || 53.9006;
			const lng = locationStore.lng || 27.5590;

			await listingsApi.updateStep2(draftId, {
				name: form.title.trim(),
				address: form.address.trim(),
				city: locationStore.address?.city || '',
				street: locationStore.address?.street || '',
				house_number: locationStore.address?.house || '',
				latitude: lat,
				longitude: lng,
				square: Number(form.size) || 1,
				floor: Number(form.floor) || 1,
				total_floors: Number(form.totalFloors) || 1,
				max_guests: Number(form.maxGuests) || 1,
				rooms_count: Number(form.bedrooms) || 1,
				beds_count: Number(form.beds) || 1,
				bathrooms_count: Number(form.bathrooms) || 1
			});
		} else if (step === 4) {
			if (!draftId) return;
			if (photoItems.length >= 5) {
				const mediaIds = photoItems.map((p) => p.mediaId || p.id).filter(Boolean);
				await listingsApi.updateStep3(draftId, { media_ids: mediaIds });
			}
		} else if (step === 5) {
			if (!draftId) return;
			if (form.amenities.length > 0) {
				await listingsApi.updateStep4(draftId, { amenities: form.amenities });
			}
		} else if (step === 6) {
			if (!draftId) return;
			await listingsApi.updateStep5(draftId, {
				price_per_night: Number(form.basePrice) || 0,
				currency: 'BYN',
				min_nights: Number(form.minNights) || 1,
				checkin_from: form.checkInFrom,
				checkout_until: form.checkOutBefore,
				rules: {
					allow_children: form.allowChildren,
					allow_pets: form.allowPets,
					allow_smoking: form.smokingAllowed,
					allow_parties: form.partiesAllowed,
					deposit_required: form.securityDepositRequired,
					with_invoicing: form.reportingDocuments
				}
			});
		} else if (step === 7) {
			if (!draftId) return;
			if (form.description.trim()) {
				await listingsApi.updateStep6(draftId, { description: form.description.trim() });
			}
		}
	}

	async function handleJumpToStep(targetStep: number): Promise<void> {
		if (targetStep === currentStep) return;
		if (targetStep < 1 || targetStep > STEPS.length) return;

		// If jumping ahead in creation mode, validate current step
		if (!isEditMode && targetStep > currentStep) {
			if (!validateStep(currentStep)) {
				toast.warning('Заполните обязательные поля перед переходом');
				return;
			}
		}

		try {
			await saveStepData(currentStep);
		} catch (err) {
			console.warn('Silent save current step on jump:', err);
		}

		previousStep = currentStep;
		currentStep = targetStep;
		submitStatus = '';
		scrollToTop();
	}

	async function handleNext(): Promise<void> {
		if (currentStep >= STEPS.length) return;
		if (!validateStep(currentStep)) {
			if (currentStep === 5) {
				toast.warning('Добавьте удобства', 'Выберите хотя бы одно удобство');
			} else {
				focusFirstInvalidField(STEP_FIELDS[currentStep] ?? []);
				toast.warning('Заполните обязательные поля', 'Проверьте выделенные поля формы');
			}
			return;
		}

		isSubmitting = true;
		submitStatus = '';
		try {
			await saveStepData(currentStep);
			previousStep = currentStep;
			currentStep += 1;
			submitStatus = '';
			scrollToTop();
		} catch (err: any) {
			console.error('Ошибка сохранения шага на сервере:', err);
			toast.error('Ошибка сохранения', formatApiError(err, 'Не удалось сохранить шаг'));
		} finally {
			isSubmitting = false;
		}
	}

	// ═══════════════════════════════════════════════════════════════════
	// PHOTO MANAGEMENT
	// ═══════════════════════════════════════════════════════════════════

	function syncPhotosToForm(): void {
		form.photos = photoItems.map(
			(p) => p.file || ({ id: p.id, url: p.url, mediaId: p.mediaId } as any)
		);
		if (touched.photos && isFieldRequired('photos')) {
			errors.photos = validateListingField('photos', form);
		} else {
			errors.photos = '';
		}
	}

	async function addPhotoFiles(input: FileList | File[]): Promise<void> {
		const files = Array.from(input).filter((f) => f.type.startsWith('image/'));
		if (!files.length) return;
		const space = Math.max(0, CONFIG.MAX_PHOTOS - photoItems.length);
		if (space === 0) {
			toast.warning('Лимит фото достигнут', `Максимум ${CONFIG.MAX_PHOTOS} фото`);
			return;
		}
		const toAdd = files.slice(0, space);
		if (toAdd.length < files.length) {
			toast.warning('Часть файлов не добавлена', `Осталось добавить только ${space} фото`);
		}

		for (const file of toAdd) {
			const id = generateId();
			const url = URL.createObjectURL(file);
			const isFirst = photoItems.length === 0;
			const item: PhotoItem = {
				id,
				file,
				url,
				isCover: isFirst
			};
			photoItems = [...photoItems, item];

			try {
				const uploaded = await mediaApi.uploadFile(file);
				item.mediaId = uploaded.mediaId;
			} catch (e) {
				console.error('Failed to upload file to storage:', e);
				toast.error('Не удалось загрузить фото на сервер');
			}
		}

		syncPhotosToForm();
	}

	function removePhoto(id: string): void {
		const item = photoItems.find((p) => p.id === id);
		if (item) URL.revokeObjectURL(item.url);
		photoItems = photoItems.filter((p) => p.id !== id);
		if (photoItems.length && !photoItems.some((p) => p.isCover)) photoItems[0].isCover = true;
		syncPhotosToForm();
	}

	function clearAllPhotos(): void {
		toast.confirm({
			title: 'Удалить все фото?',
			message: 'Это действие нельзя отменить.',
			confirmText: 'Удалить',
			cancelText: 'Отмена',
			type: 'danger',
			onConfirm: () => {
				photoItems.forEach((p) => URL.revokeObjectURL(p.url));
				photoItems = [];
				syncPhotosToForm();
				toast.success('Фотографии удалены');
			}
		});
	}

	function setCoverPhoto(id: string): void {
		const selected = photoItems.find((p) => p.id === id);
		if (!selected) return;
		photoItems = [
			{ ...selected, isCover: true },
			...photoItems.filter((p) => p.id !== id).map((p) => ({ ...p, isCover: false }))
		];
		syncPhotosToForm();
	}

	function handlePhotoReorder(items: PhotoItem[]): void {
		const newIds = new Set(items.map((i) => i.id));
		photoItems.forEach((p) => {
			if (!newIds.has(p.id)) URL.revokeObjectURL(p.url);
		});
		photoItems = items;
		syncPhotosToForm();
	}

	// ═══════════════════════════════════════════════════════════════════
	// FORM SUBMISSION
	// ═══════════════════════════════════════════════════════════════════

	async function handleSubmit(): Promise<void> {
		if (!authStore.user) {
			toast.error('Недостаточно прав', 'Войдите в аккаунт');
			goto(resolve('/auth'));
			return;
		}
		if (!isHostUser(authStore.user)) {
			authStore.becomeHost();
		}

		if (!draftId) {
			toast.error('Черновик не найден');
			return;
		}

		submitStatus = '';
		if (!validateAllFields()) {
			submitStatus = 'Проверьте выделенные поля формы.';
			if (amenitiesError && currentStep !== 5) {
				previousStep = currentStep;
				currentStep = 5;
			}
			focusFirstInvalidField(validatedListingFields.filter(isFieldApplicable));
			toast.error('Ошибка валидации', 'Проверьте выделенные поля');
			return;
		}

		isSubmitting = true;
		try {
			const idempotencyKey = crypto.randomUUID();
			const res = await listingsApi.submitDraft(draftId, idempotencyKey);

			clearDraft(currentUserId);
			localStorage.removeItem('flickey_current_draft_id');
			await hostListingsStore.refresh();
			authStore.becomeHost();

			navigationConfirmed = true;
			if (res.publication_action === 'NO_CHANGE') {
				toast.success('Изменения сохранены!', 'Объявление активно');
				goto(resolve('/host/listings'));
			} else if (isEditMode) {
				toast.success('Изменения отправлены на модерацию');
				goto(resolve('/host/listings'));
			} else {
				toast.success('Черновик сохранен!', 'Для публикации загрузите видео квартиры');
				goto(resolve('/host/listings'));
			}
		} catch (error: any) {
			console.error('Ошибка публикации:', error);
			submitStatus = formatApiError(error, 'Произошла ошибка при отправке.');
			toast.error('Ошибка публикации', submitStatus);
		} finally {
			isSubmitting = false;
		}
	}

	// ═══════════════════════════════════════════════════════════════════
	// RESET & CLEANUP
	// ═══════════════════════════════════════════════════════════════════

	onDestroy(() => {
		photoItems.forEach((p) => URL.revokeObjectURL(p.url));
	});

	// ═══════════════════════════════════════════════════════════════════
	// DERIVED
	// ═══════════════════════════════════════════════════════════════════

	const isFirstStep = $derived(currentStep === 1);
	const isLastStep = $derived(currentStep === STEPS.length);
	const showResetAmenities = $derived(currentStep === 5 && form.amenities.length > 0);
	const isErrorStatus = $derived(
		submitStatus.startsWith('Проверьте') || submitStatus.startsWith('Произошла')
	);
	const stepLabels = $derived(STEPS.map((s) => s.label));
</script>

<svelte:window onbeforeunload={handleBeforeUnload} />
<svelte:head>
	<title>{isEditMode ? 'Редактировать объявление' : 'Создать объявление'} — Flickey</title>
</svelte:head>

{#if guardChecked}
	{#if showWelcomeScreen}
		<ListingWelcomeScreen
			onStart={startCreationFlow}
			onExit={handleExitCreation}
			hasDraft={hasRestoredDraft}
			resumeStep={currentStep}
		/>
	{:else}
		<div class="flex min-h-screen flex-col bg-white">
			<!-- ── FIXED HEADER ── -->
			<header class="fixed inset-x-0 top-0 z-50 h-16 bg-white sm:h-20 border-b border-zinc-100">
				<div class="flex h-full items-center justify-between px-4 sm:px-8 md:px-12">
					<div class="flex items-center gap-3">
						<span
							class="inline-flex items-center rounded-lg px-2.5 py-1 text-xs font-semibold {isEditMode
								? 'bg-amber-100 text-amber-900'
								: 'bg-zinc-100 text-zinc-800'}"
						>
							{isEditMode ? 'Редактирование' : 'Создание'}
						</span>
						<p class="hidden text-sm font-semibold text-zinc-800 sm:block">
							Шаг {currentStep} из {STEPS.length}: {STEPS[currentStep - 1]?.label}
						</p>
					</div>

					<!-- Центр: название шага на мобильных -->
					<p class="text-sm font-semibold text-zinc-800 sm:hidden">
						{STEPS[currentStep - 1]?.label}
					</p>

					<!-- Правая часть: крестик -->
					<Button
						variant="ghost"
						tone="neutral"
						size="icon"
						radius="pill"
						onclick={handleExitCreation}
						aria-label="Выйти"
					>
						<X class="h-5 w-5" />
					</Button>
				</div>
			</header>

			<!-- ── MAIN CONTENT ── -->
			<main class="flex-1 pt-16 pb-28 sm:pt-20">
				<div class="mx-auto max-w-[800px] px-5 sm:px-6 lg:px-8">
					<form
						onsubmit={(e) => {
							e.preventDefault();
							handleSubmit();
						}}
						novalidate
					>
						<div class="relative py-4 sm:py-6">
							{#if currentStep === 1}
								<Step1PropertyType
									{form}
									{errors}
									{touched}
									onInput={handleInput}
									onBlur={handleBlur}
								/>
							{:else if currentStep === 2}
								<Step2Location
									{form}
									{errors}
									{touched}
									onInput={handleInput}
									onBlur={handleBlur}
								/>
							{:else if currentStep === 3}
								<Step3MainParams
									{form}
									{errors}
									{touched}
									onInput={handleInput}
									onBlur={handleBlur}
								/>
							{:else if currentStep === 4}
								<Step4Photos
									{errors}
									{touched}
									{photoItems}
									maxPhotos={CONFIG.MAX_PHOTOS}
									onAddFiles={addPhotoFiles}
									onRemovePhoto={removePhoto}
									onClearAll={clearAllPhotos}
									onSetCover={setCoverPhoto}
									onReorder={handlePhotoReorder}
								/>
							{:else if currentStep === 5}
								<Step5Amenities
									{form}
									propertyType={amenityPropertyType}
									categories={amenityCategories}
									presetOptions={amenityPresetOptions}
									error={amenitiesError}
									onToggleAmenity={toggleAmenity}
									onApplyPreset={applyAmenityPreset}
								/>
								{#if showResetAmenities}
									<div class="mt-4 flex justify-end">
										<Button
											variant="link"
											tone="neutral"
											size="sm"
											radius="xl"
											onclick={resetAmenities}
										>
											Сбросить выбор
										</Button>
									</div>
								{/if}
							{:else if currentStep === 6}
								<Step6PriceRules
									{form}
									{errors}
									{touched}
									onInput={handleInput}
									onBlur={handleBlur}
									onToggleBoolean={toggleBoolean}
								/>
							{:else if currentStep === 7}
								<Step7Description
									{form}
									{errors}
									{touched}
									onInput={handleInput}
									onBlur={handleBlur}
								/>
							{:else if currentStep === 8}
								<Step8Contacts
									{form}
									{errors}
									{touched}
									onBlur={handleBlur}
									onUpdateContactMethods={handleUpdateContactMethods}
								/>
							{/if}
						</div>
					</form>
				</div>
			</main>

			<!-- ── FIXED FOOTER ── -->
			<footer class="fixed inset-x-0 bottom-0 z-50 bg-white border-t border-zinc-200 shadow-lg">
				<ListingStepsHeader
					{currentStep}
					totalSteps={STEPS.length}
					{stepLabels}
					{isEditMode}
					onStepClick={handleJumpToStep}
				/>

				<div class="mx-auto max-w-[1400px] px-4 py-3 sm:px-6 sm:py-4">
					{#if isEditMode}
						<!-- РЕЖИМ РЕДАКТИРОВАНИЯ: кнопки с цифрами шагов 1..8 вместо назад/вперед -->
						<div class="flex items-center justify-between gap-3">
							<div class="flex flex-1 items-center overflow-x-auto py-1 scrollbar-none gap-1 sm:gap-2">
								{#each STEPS as step}
									<button
										type="button"
										onclick={() => handleJumpToStep(step.id)}
										class="flex items-center gap-1.5 rounded-xl px-2.5 py-1.5 text-xs font-medium transition-all duration-150 sm:px-3 sm:py-2 sm:text-sm shrink-0
											{currentStep === step.id
												? 'bg-zinc-900 text-white shadow-sm ring-2 ring-zinc-900/10'
												: 'bg-zinc-100 text-zinc-700 hover:bg-zinc-200 active:scale-95'}"
										title={step.label}
									>
										<span
											class="flex h-5 w-5 items-center justify-center rounded-full text-xs font-semibold
											{currentStep === step.id ? 'bg-white/20 text-white' : 'text-zinc-600'}"
										>
											{step.id}
										</span>
										<span class="hidden md:inline font-medium">
											{step.label}
										</span>
									</button>
								{/each}
							</div>

							<!-- Ошибка сабмита -->
							{#if submitStatus}
								<p
									class="hidden text-center text-xs font-medium lg:block
									{isErrorStatus ? 'text-red-600' : 'text-emerald-600'}"
								>
									{submitStatus}
								</p>
							{/if}

							<!-- Сохранить изменения -->
							<div class="shrink-0">
								<Button
									type="submit"
									variant="solid"
									tone="primary"
									size="md"
									radius="xl"
									loading={isSubmitting}
									loadingText="Сохранение..."
									disabled={isSubmitting}
									onclick={handleSubmit}
								>
									Сохранить
								</Button>
							</div>
						</div>
					{:else}
						<!-- РЕЖИМ СОЗДАНИЯ: Назад, цифры шагов 1..8, Далее -->
						<div class="flex items-center justify-between gap-3">
							<!-- Назад -->
							<div class="w-24 sm:w-28">
								{#if !isFirstStep}
									<Button
										variant="ghost"
										tone="neutral"
										size="md"
										radius="xl"
										iconLeft={ChevronLeft}
										disabled={isSubmitting}
										onclick={handlePrev}
									>
										Назад
									</Button>
								{/if}
							</div>

							<!-- Цифры шагов 1..8 -->
							<div class="flex items-center justify-center gap-1 sm:gap-1.5">
								{#each STEPS as step}
									<button
										type="button"
										onclick={() => handleJumpToStep(step.id)}
										class="flex h-8 w-8 items-center justify-center rounded-xl text-xs font-semibold transition-all duration-150 sm:h-9 sm:w-9 sm:text-sm
											{currentStep === step.id
												? 'bg-zinc-900 text-white shadow-sm'
												: 'bg-zinc-100 text-zinc-600 hover:bg-zinc-200'}"
										title="{step.id}. {step.label}"
									>
										{step.id}
									</button>
								{/each}
							</div>

							<!-- Далее / Опубликовать -->
							<div class="flex justify-end w-24 sm:w-28">
								{#if !isLastStep}
									<Button
										variant="solid"
										tone="primary"
										size="md"
										radius="xl"
										iconRight={ChevronRight}
										disabled={isSubmitting}
										onclick={handleNext}
									>
										Далее
									</Button>
								{:else}
									<Button
										type="submit"
										variant="solid"
										tone="primary"
										size="md"
										radius="xl"
										loading={isSubmitting}
										loadingText="Публикация..."
										disabled={isSubmitting}
										onclick={handleSubmit}
									>
										Опубликовать
									</Button>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			</footer>
		</div>
	{/if}
{:else}
	<div class="flex min-h-screen items-center justify-center bg-zinc-50">
		<div class="flex flex-col items-center gap-3">
			<div
				class="h-8 w-8 animate-spin rounded-full border-2 border-zinc-900 border-t-transparent"
			></div>
			<p class="text-sm text-zinc-500">Проверяем доступ...</p>
		</div>
	</div>
{/if}
