<script lang="ts">
	import { auth } from '$lib/auth/auth.svelte';
	import { listingsApi } from '$lib/api/listings';
	import { mediaApi } from '$lib/api/media';
	import { amenitiesApi } from '$lib/api/amenities';
	import type { HousingType, ListingRules } from '$lib/types/listings';
	import type { AmenitiesResponse } from '$lib/types/amenities';
	import type { UploadedPhoto } from '$lib/types/media';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import Switch from '$lib/components/ui/Switch.svelte';
	import Alert from '$lib/components/ui/Alert.svelte';
	import { toast } from '$lib/components/ui/toast.svelte';
	import { formatCurrency, translateHousingType, cn } from '$lib/utils';
	import {
		Building,
		Home,
		Trees,
		UploadCloud,
		CheckCircle2,
		X,
		Image as ImageIcon,
		Sparkles,
		Clock,
		Banknote,
		ShieldAlert,
		ArrowRight,
		ArrowLeft,
		Plus,
		Eye,
		EyeOff
	} from 'lucide-svelte';

	import { ApiError } from '$lib/types/api';
	import type { ListingPublic } from '$lib/types/listings';
	import type { GeoSuggestItem } from '$lib/types/geo';
	import ListingView from '$lib/components/listings/ListingView.svelte';
	import AddressAutocomplete from '$lib/components/listings/AddressAutocomplete.svelte';
	import AddressMapPicker from '$lib/components/listings/AddressMapPicker.svelte';

	// Wizard navigation state
	let currentStep = $state<number>(1);
	let draftId = $state<string | null>(null);
	let loading = $state<boolean>(false);
	let errorMessage = $state<string | null>(null);
	let submittedListingId = $state<string | null>(null);
	let isLivePreviewMode = $state<boolean>(false);

	// Step 1: Housing type
	let housingType = $state<HousingType>('apartment');

	// Step 2: Parameters & Location
	let name = $state('');
	let address = $state('');
	let latitude = $state<number>(53.9006);
	let longitude = $state<number>(27.5590);
	let geoSelected = $state<GeoSuggestItem | null>(null);
	let square = $state<number>(55);
	let floor = $state<number>(3);
	let totalFloors = $state<number>(9);
	let maxGuests = $state<number>(4);
	let roomsCount = $state<number>(2);
	let bedsCount = $state<number>(2);
	let bathroomsCount = $state<number>(1);

	function handleAddressSelect(item: GeoSuggestItem) {
		const lat = typeof item.latitude === 'number' ? item.latitude : parseFloat(String(item.latitude));
		const lng = typeof item.longitude === 'number' ? item.longitude : parseFloat(String(item.longitude));

		address = item.raw_address;
		if (!isNaN(lat) && !isNaN(lng)) {
			latitude = lat;
			longitude = lng;
		}
		if (!name.trim()) {
			const streetPart = item.street ? `${item.street}${item.house_number ? ', ' + item.house_number : ''}` : item.city || item.raw_address;
			name = `Уютное жилье — ${streetPart}`;
		}
	}

	// Step 3: Photos
	let photos = $state<UploadedPhoto[]>([]);
	let isDragging = $state(false);

	// Step 4: Amenities
	let amenitiesCatalog = $state<AmenitiesResponse | null>(null);
	let selectedAmenities = $state<Set<string>>(new Set());
	let loadingAmenities = $state(false);

	// Step 5: Pricing & Rules
	let pricePerNight = $state<number>(120);
	let minNights = $state<number>(1);
	let checkinFrom = $state('14:00');
	let checkoutUntil = $state('12:00');
	let rules = $state<ListingRules>({
		allow_children: true,
		allow_pets: false,
		allow_smoking: false,
		allow_parties: false,
		deposit_required: false,
		with_invoicing: false
	});

	// Step 6: Description
	let description = $state('');

	// Steps configuration
	const steps = [
		{ num: 1, label: 'Тип' },
		{ num: 2, label: 'Инфо' },
		{ num: 3, label: 'Фото' },
		{ num: 4, label: 'Удобства' },
		{ num: 5, label: 'Цены' },
		{ num: 6, label: 'Описание' },
		{ num: 7, label: 'Публикация' }
	];

	// Derived counts
	let uploadedPhotosCount = $derived(photos.filter((p) => p.status === 'uploaded' && p.mediaId).length);
	let uploadedMediaIds = $derived(
		photos.filter((p) => p.status === 'uploaded' && p.mediaId).map((p) => p.mediaId as string)
	);

	// Reactive preview listing representation
	let livePreviewListing = $derived<ListingPublic>({
		id: draftId || 'draft-preview',
		type: housingType,
		name: name.trim() || 'Объект недвижимости',
		square: Number(square) || 50,
		floor: Number(floor) || 1,
		total_floors: Number(totalFloors) || 1,
		max_guests: Number(maxGuests) || 1,
		rooms_count: Number(roomsCount) || 1,
		beds_count: Number(bedsCount) || 1,
		bathrooms_count: Number(bathroomsCount) || 1,
		price_per_night: Number(pricePerNight) || 100,
		currency: 'BYN',
		min_nights: Number(minNights) || 1,
		checkin_from: checkinFrom || '14:00',
		checkout_until: checkoutUntil || '12:00',
		description: description.trim() || 'Описание объекта еще не заполнено.',
		amenities: Array.from(selectedAmenities),
		media: photos.map((p) => p.previewUrl).filter(Boolean),
		created_at: new Date().toISOString()
	});

	// Step 1: Submit Housing Type
	async function submitStep1() {
		if (auth.isProfilePending) {
			auth.openAuthModal('complete_profile');
			toast.info('Пожалуйста, заполните имя и фамилию для завершения регистрации');
			return;
		}

		errorMessage = null;
		loading = true;
		try {
			const res = await listingsApi.createDraft({ type: housingType });
			draftId = res.draft_id;
			currentStep = 2;
		} catch (err) {
			if (err instanceof ApiError && err.code === 'PROFILE_INCOMPLETE') {
				auth.openAuthModal('complete_profile');
				toast.info('Пожалуйста, заполните имя и фамилию для завершения регистрации');
				return;
			}
			errorMessage = err instanceof Error ? err.message : 'Не удалось создать черновик';
		} finally {
			loading = false;
		}
	}

	// Step 2: Submit Physical Info
	async function submitStep2() {
		if (!draftId) return;
		if (name.trim().length < 10) {
			errorMessage = 'Название объекта должно содержать не менее 10 символов';
			return;
		}
		if (floor > totalFloors) {
			errorMessage = 'Этаж не может быть больше общего количества этажей';
			return;
		}

		errorMessage = null;
		loading = true;
		try {
			await listingsApi.updateStep2(draftId, {
				name: name.trim(),
				square: Number(square),
				floor: Number(floor),
				total_floors: Number(totalFloors),
				max_guests: Number(maxGuests),
				rooms_count: Number(roomsCount),
				beds_count: Number(bedsCount),
				bathrooms_count: Number(bathroomsCount)
			});
			currentStep = 3;
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Не удалось сохранить данные шага 2';
		} finally {
			loading = false;
		}
	}

	// Step 3: Handle Photo Files
	async function handlePhotoFiles(fileList: FileList | File[]) {
		const files = Array.from(fileList);
		if (photos.length + files.length > 15) {
			toast.warning('Максимум можно загрузить 15 фотографий');
		}

		const allowedFiles = files.slice(0, 15 - photos.length);

		for (const file of allowedFiles) {
			const id = crypto.randomUUID();
			const previewUrl = URL.createObjectURL(file);

			const photoItem: UploadedPhoto = {
				id,
				file,
				previewUrl,
				status: 'uploading'
			};

			photos = [...photos, photoItem];

			// Start upload process by ID
			uploadPhotoById(id, file);
		}
	}

	async function uploadPhotoById(id: string, file: File) {
		try {
			// 1. Presign
			const contentType = (file.type as 'image/jpeg' | 'image/png' | 'image/webp') || 'image/jpeg';
			const presignRes = await mediaApi.presign({
				content_type: contentType,
				file_size_bytes: file.size || 1024
			});

			// 2. Direct upload to S3 / dev storage
			await mediaApi.uploadDirect(presignRes.upload_url, file, contentType);

			// 3. Complete verification
			await mediaApi.complete(presignRes.media_id);

			// Update reactive state by ID
			const idx = photos.findIndex((p) => p.id === id);
			if (idx !== -1) {
				photos[idx] = {
					...photos[idx],
					mediaId: presignRes.media_id,
					status: 'uploaded'
				};
			}
		} catch (err) {
			const idx = photos.findIndex((p) => p.id === id);
			if (idx !== -1) {
				photos[idx] = {
					...photos[idx],
					status: 'failed',
					errorMessage: err instanceof Error ? err.message : 'Ошибка загрузки'
				};
			}
			toast.error('Не удалось загрузить фото');
		}
	}

	function removePhoto(id: string) {
		const idx = photos.findIndex((p) => p.id === id);
		if (idx !== -1) {
			if (photos[idx].previewUrl) URL.revokeObjectURL(photos[idx].previewUrl);
			photos = photos.filter((p) => p.id !== id);
		}
	}

	async function submitStep3() {
		if (!draftId) return;
		if (uploadedPhotosCount < 5) {
			errorMessage = 'Необходимо загрузить и подтвердить минимум 5 фотографий (загружено: ' + uploadedPhotosCount + ')';
			return;
		}

		errorMessage = null;
		loading = true;
		try {
			await listingsApi.updateStep3(draftId, {
				media_ids: uploadedMediaIds
			});
			currentStep = 4;
			loadAmenitiesCatalog();
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Не удалось сохранить фотографии';
		} finally {
			loading = false;
		}
	}

	// Step 4: Amenities
	async function loadAmenitiesCatalog() {
		if (amenitiesCatalog) return;
		loadingAmenities = true;
		try {
			amenitiesCatalog = await amenitiesApi.getAmenities(housingType);
		} catch (err) {
			toast.error('Не удалось загрузить каталог удобств');
		} finally {
			loadingAmenities = false;
		}
	}

	function toggleAmenity(id: string) {
		if (selectedAmenities.has(id)) {
			selectedAmenities.delete(id);
		} else {
			selectedAmenities.add(id);
		}
		selectedAmenities = new Set(selectedAmenities);
	}

	async function submitStep4() {
		if (!draftId) return;
		if (selectedAmenities.size === 0) {
			errorMessage = 'Выберите хотя бы одно удобство';
			return;
		}

		errorMessage = null;
		loading = true;
		try {
			await listingsApi.updateStep4(draftId, {
				amenities: Array.from(selectedAmenities)
			});
			currentStep = 5;
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Не удалось сохранить удобства';
		} finally {
			loading = false;
		}
	}

	// Step 5: Pricing
	async function submitStep5() {
		if (!draftId) return;
		if (pricePerNight <= 0) {
			errorMessage = 'Цена за сутки должна быть больше 0';
			return;
		}

		errorMessage = null;
		loading = true;
		try {
			await listingsApi.updateStep5(draftId, {
				price_per_night: Number(pricePerNight),
				currency: 'BYN',
				min_nights: Number(minNights),
				checkin_from: checkinFrom,
				checkout_until: checkoutUntil,
				rules
			});
			currentStep = 6;
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Не удалось сохранить цены и правила';
		} finally {
			loading = false;
		}
	}

	// Step 6: Description
	async function submitStep6() {
		if (!draftId) return;
		if (description.trim().length < 30) {
			errorMessage = 'Описание должно содержать не менее 30 символов (сейчас: ' + description.trim().length + ')';
			return;
		}

		errorMessage = null;
		loading = true;
		try {
			await listingsApi.updateStep6(draftId, {
				description: description.trim()
			});
			currentStep = 7;
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Не удалось сохранить описание';
		} finally {
			loading = false;
		}
	}

	// Step 7: Submit Draft (Publish)
	async function handleFinalPublish() {
		if (!draftId) return;

		errorMessage = null;
		loading = true;
		try {
			// Generate UUIDv4 idempotency key
			const idempotencyKey = crypto.randomUUID();
			const res = await listingsApi.submitDraft(draftId, idempotencyKey);
			submittedListingId = res.listing_id;

			// Refresh user profile to detect automatic Host role promotion
			await auth.refreshUser();
			auth.switchContext('host');

			toast.success('Объявление успешно создано! Вы стали Хостом.');
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Не удалось опубликовать объявление';
		} finally {
			loading = false;
		}
	}

	function resetWizard() {
		currentStep = 1;
		draftId = null;
		submittedListingId = null;
		photos = [];
		selectedAmenities = new Set();
		name = '';
		description = '';
	}
</script>

<div class={cn('mx-auto py-6 px-4 transition-all duration-300', isLivePreviewMode ? 'max-w-6xl' : 'max-w-3xl')}>
	<!-- Top Bar with Live Preview Toggle -->
	<div class="flex items-center justify-between gap-4 mb-6">
		<div>
			<h1 class="text-xl sm:text-2xl font-extrabold text-foreground tracking-tight">
				{isLivePreviewMode ? 'Предпросмотр объявления' : 'Создание нового объявления'}
			</h1>
			<p class="text-xs text-muted-foreground mt-0.5">
				{isLivePreviewMode
					? 'Так ваше объявление увидят потенциальные гости на сайте'
					: `Шаг ${currentStep} из ${steps.length}: ${steps[currentStep - 1]?.label}`}
			</p>
		</div>

		{#if !submittedListingId}
			<Button
				type="button"
				variant={isLivePreviewMode ? 'default' : 'outline'}
				size="sm"
				onclick={() => (isLivePreviewMode = !isLivePreviewMode)}
				class="gap-1.5 text-xs font-semibold shrink-0"
			>
				{#if isLivePreviewMode}
					<EyeOff class="h-3.5 w-3.5" /> Режим заполнения
				{:else}
					<Eye class="h-3.5 w-3.5" /> Предпросмотр
				{/if}
			</Button>
		{/if}
	</div>

	<!-- Error Alert Box -->
	{#if errorMessage}
		<div class="mb-6 animate-in fade-in">
			<Alert variant="destructive">{errorMessage}</Alert>
		</div>
	{/if}

	<!-- LIVE PREVIEW VIEW -->
	{#if isLivePreviewMode}
		<div class="space-y-6 animate-in fade-in">
			<ListingView
				listing={livePreviewListing}
				mode="preview"
				onedit={() => (isLivePreviewMode = false)}
				onpublish={handleFinalPublish}
			/>
		</div>
	{:else if submittedListingId}
		<!-- SUCCESS VIEW: Submitted -->
		<div class="rounded-3xl border border-emerald-500/30 bg-card p-8 sm:p-12 text-center shadow-xl animate-in zoom-in-95">
			<div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-emerald-100 text-emerald-600 mb-6">
				<Sparkles class="h-8 w-8" />
			</div>
			<h2 class="text-2xl sm:text-3xl font-extrabold text-foreground mb-3">
				Объявление успешно создано!
			</h2>
			<p class="text-muted-foreground text-sm max-w-md mx-auto mb-6 leading-relaxed">
				Ваше объявление отправлено на верификацию. Ваша роль в системе автоматически повышена до <strong>Хоста</strong>.
			</p>

			<div class="rounded-2xl bg-muted/60 p-4 max-w-md mx-auto text-left text-xs space-y-2 mb-8 font-mono border border-border">
				<div class="flex justify-between">
					<span class="text-muted-foreground">ID объявления:</span>
					<span class="font-bold text-foreground truncate max-w-[200px]">{submittedListingId}</span>
				</div>
				<div class="flex justify-between">
					<span class="text-muted-foreground">Статус:</span>
					<span class="font-bold text-emerald-600">awaiting_company_verification</span>
				</div>
			</div>

			<div class="flex flex-wrap gap-3 justify-center">
				<a href="/host">
					<Button class="font-semibold shadow-md gap-2">
						<Home class="h-4 w-4" /> В кабинет хоста
					</Button>
				</a>
				<Button variant="outline" onclick={resetWizard}>
					<Plus class="h-4 w-4" /> Создать еще одно
				</Button>
			</div>
		</div>
	{:else}
		<!-- Stepper Navigation Bar -->
		<div class="mb-8">
			<div class="flex items-center justify-between relative">
				<!-- Background track line -->
				<div class="absolute left-0 top-1/2 -translate-y-1/2 h-1 w-full bg-muted z-0"></div>
				<div
					class="absolute left-0 top-1/2 -translate-y-1/2 h-1 bg-primary transition-all duration-300 z-0"
					style={`width: ${((currentStep - 1) / (steps.length - 1)) * 100}%`}
				></div>

				{#each steps as s}
					<div class="flex flex-col items-center relative z-10">
						<div
							class={cn(
								'flex h-8 w-8 sm:h-9 sm:w-9 items-center justify-center rounded-full text-xs font-bold transition-all',
								currentStep === s.num
									? 'bg-primary text-primary-foreground ring-4 ring-primary/20 shadow-md'
									: currentStep > s.num
										? 'bg-primary text-primary-foreground'
										: 'bg-card border-2 border-muted text-muted-foreground'
							)}
						>
							{#if currentStep > s.num}
								<CheckCircle2 class="h-4 w-4" />
							{:else}
								{s.num}
							{/if}
						</div>
						<span class="mt-1 text-[11px] font-semibold text-muted-foreground hidden sm:block">
							{s.label}
						</span>
					</div>
				{/each}
			</div>
		</div>

		<!-- STEP 1: Type -->
		{#if currentStep === 1}
			<div class="rounded-2xl border border-border bg-card p-6 sm:p-8 shadow-sm space-y-6">
				<div>
					<h2 class="text-xl font-bold text-foreground">Шаг 1: Выберите тип недвижимости</h2>
					<p class="text-sm text-muted-foreground mt-1">
						Укажите категорию жилья, которое вы планируете сдать в аренду.
					</p>
				</div>

				<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
					<!-- Apartment -->
					<button
						type="button"
						onclick={() => (housingType = 'apartment')}
						class={cn(
							'flex flex-col items-start p-5 rounded-2xl border-2 text-left transition-all cursor-pointer',
							housingType === 'apartment'
								? 'border-primary bg-primary/5 shadow-md'
								: 'border-border hover:border-border/80 hover:bg-accent'
						)}
					>
						<div class="p-3 rounded-xl bg-primary/10 text-primary mb-3">
							<Building class="h-6 w-6" />
						</div>
						<h4 class="font-bold text-base text-foreground">Квартира</h4>
						<p class="text-xs text-muted-foreground mt-1 leading-relaxed">
							Апартаменты, студии и городские квартиры.
						</p>
					</button>

					<!-- House -->
					<button
						type="button"
						onclick={() => (housingType = 'house')}
						class={cn(
							'flex flex-col items-start p-5 rounded-2xl border-2 text-left transition-all cursor-pointer',
							housingType === 'house'
								? 'border-primary bg-primary/5 shadow-md'
								: 'border-border hover:border-border/80 hover:bg-accent'
						)}
					>
						<div class="p-3 rounded-xl bg-primary/10 text-primary mb-3">
							<Home class="h-6 w-6" />
						</div>
						<h4 class="font-bold text-base text-foreground">Дом / Коттедж</h4>
						<p class="text-xs text-muted-foreground mt-1 leading-relaxed">
							Частные загородные дома, коттеджи и таунхаусы.
						</p>
					</button>

					<!-- Manor -->
					<button
						type="button"
						onclick={() => (housingType = 'manor')}
						class={cn(
							'flex flex-col items-start p-5 rounded-2xl border-2 text-left transition-all cursor-pointer',
							housingType === 'manor'
								? 'border-primary bg-primary/5 shadow-md'
								: 'border-border hover:border-border/80 hover:bg-accent'
						)}
					>
						<div class="p-3 rounded-xl bg-primary/10 text-primary mb-3">
							<Trees class="h-6 w-6" />
						</div>
						<h4 class="font-bold text-base text-foreground">Усадьба</h4>
						<p class="text-xs text-muted-foreground mt-1 leading-relaxed">
							Агроусадьбы и крупные загородные комплексы.
						</p>
					</button>
				</div>

				<div class="flex justify-end pt-4 border-t border-border">
					<Button onclick={submitStep1} {loading} class="gap-2 font-semibold">
						<span>Продолжить</span> <ArrowRight class="h-4 w-4" />
					</Button>
				</div>
			</div>
		{/if}

		<!-- STEP 2: Parameters -->
		{#if currentStep === 2}
			<div class="rounded-2xl border border-border bg-card p-6 sm:p-8 shadow-sm space-y-6">
				<div>
					<h2 class="text-xl font-bold text-foreground">Шаг 2: Основная информация</h2>
					<p class="text-sm text-muted-foreground mt-1">
						Укажите название и физические параметры объекта.
					</p>
				</div>

				<div class="space-y-4">
					<!-- Address & Geocoding -->
					<div class="space-y-2.5 p-4 rounded-2xl bg-muted/20 border border-border/70">
						<div class="flex items-center justify-between">
							<span class="text-xs font-semibold text-foreground">Адрес объекта</span>
							<span class="text-[10px] text-muted-foreground font-normal">Поиск по базе OpenStreetMap</span>
						</div>
						<AddressAutocomplete
							bind:value={address}
							bind:latitude={latitude}
							bind:longitude={longitude}
							bind:selected={geoSelected}
							onselect={handleAddressSelect}
							placeholder="Введите город, улицу, номер дома..."
						/>

						{#if address || (latitude && longitude)}
							<div class="pt-2">
								<AddressMapPicker
									bind:latitude={latitude}
									bind:longitude={longitude}
									{address}
								/>
							</div>
						{/if}
					</div>

					<div class="space-y-1.5">
						<label for="w-name" class="text-xs font-semibold text-foreground">
							Название объявления <span class="text-destructive">*</span>
						</label>
						<Input
							id="w-name"
							placeholder="Например: Просторная 2-комнатная квартира в центре Минска"
							bind:value={name}
						/>
						<p class="text-[11px] text-muted-foreground">Минимум 10 символов</p>
					</div>

					<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
						<div class="space-y-1.5">
							<label for="w-square" class="text-xs font-semibold text-foreground">Площадь (м²)</label>
							<Input id="w-square" type="number" min="10" max="1000" bind:value={square} />
						</div>
						<div class="space-y-1.5">
							<label for="w-floor" class="text-xs font-semibold text-foreground">Этаж</label>
							<Input id="w-floor" type="number" min="1" max="100" bind:value={floor} />
						</div>
						<div class="space-y-1.5">
							<label for="w-tot-floor" class="text-xs font-semibold text-foreground">Всего этажей</label>
							<Input id="w-tot-floor" type="number" min="1" max="100" bind:value={totalFloors} />
						</div>
					</div>

					<div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
						<div class="space-y-1.5">
							<label for="w-guests" class="text-xs font-semibold text-foreground">Макс. гостей</label>
							<Input id="w-guests" type="number" min="1" max="50" bind:value={maxGuests} />
						</div>
						<div class="space-y-1.5">
							<label for="w-rooms" class="text-xs font-semibold text-foreground">Комнат</label>
							<Input id="w-rooms" type="number" min="1" max="30" bind:value={roomsCount} />
						</div>
						<div class="space-y-1.5">
							<label for="w-beds" class="text-xs font-semibold text-foreground">Спальных мест</label>
							<Input id="w-beds" type="number" min="1" max="30" bind:value={bedsCount} />
						</div>
						<div class="space-y-1.5">
							<label for="w-baths" class="text-xs font-semibold text-foreground">Санузлов</label>
							<Input id="w-baths" type="number" min="1" max="10" bind:value={bathroomsCount} />
						</div>
					</div>
				</div>

				<div class="flex justify-between pt-4 border-t border-border">
					<Button variant="outline" onclick={() => (currentStep = 1)}>
						<ArrowLeft class="h-4 w-4 mr-1" /> Назад
					</Button>
					<Button onclick={submitStep2} {loading} class="gap-2 font-semibold">
						<span>Далее (Фотографии)</span> <ArrowRight class="h-4 w-4" />
					</Button>
				</div>
			</div>
		{/if}

		<!-- STEP 3: Photos (Direct S3 Drag-and-Drop) -->
		{#if currentStep === 3}
			<div class="rounded-2xl border border-border bg-card p-6 sm:p-8 shadow-sm space-y-6">
				<div class="flex items-start justify-between gap-4">
					<div>
						<h2 class="text-xl font-bold text-foreground">Шаг 3: Фотографии объекта</h2>
						<p class="text-sm text-muted-foreground mt-1">
							Загрузите от 5 до 15 качественных фото (JPEG, PNG или WebP).
						</p>
					</div>
					<div class="text-right">
						<span
							class={cn(
								'inline-flex items-center rounded-full px-3 py-1 text-xs font-bold',
								uploadedPhotosCount >= 5
									? 'bg-emerald-100 text-emerald-800'
									: 'bg-amber-100 text-amber-800'
							)}
						>
							{uploadedPhotosCount} / 15 фото (мин. 5)
						</span>
					</div>
				</div>

				<!-- Drag & Drop Zone -->
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div
					class={cn(
						'border-2 border-dashed rounded-3xl p-8 sm:p-12 text-center transition-all cursor-pointer flex flex-col items-center justify-center',
						isDragging ? 'border-primary bg-primary/10' : 'border-border hover:border-primary/50 bg-muted/20'
					)}
					ondragover={(e) => {
						e.preventDefault();
						isDragging = true;
					}}
					ondragleave={() => (isDragging = false)}
					ondrop={(e) => {
						e.preventDefault();
						isDragging = false;
						if (e.dataTransfer?.files) handlePhotoFiles(e.dataTransfer.files);
					}}
					onclick={() => document.getElementById('photo-file-picker')?.click()}
					onkeydown={(e) => {
						if (e.key === 'Enter' || e.key === ' ') document.getElementById('photo-file-picker')?.click();
					}}
					tabindex="0"
					role="button"
				>
					<input
						type="file"
						id="photo-file-picker"
						multiple
						accept="image/jpeg,image/png,image/webp"
						class="hidden"
						onchange={(e) => {
							const target = e.currentTarget;
							if (target.files) handlePhotoFiles(target.files);
						}}
					/>

					<div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 text-primary mb-4">
						<UploadCloud class="h-7 w-7" />
					</div>
					<h4 class="font-bold text-base text-foreground mb-1">
						Перетащите фотографии сюда или нажмите для выбора
					</h4>
					<p class="text-xs text-muted-foreground">
						Поддерживаются форматы JPG, PNG, WEBP до 15 МБ
					</p>
				</div>

				<!-- Photo Previews Grid -->
				{#if photos.length > 0}
					<div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
						{#each photos as photo (photo.id)}
							<div class="group relative aspect-square rounded-2xl overflow-hidden border border-border bg-muted">
								<img src={photo.previewUrl} alt="Фото" class="h-full w-full object-cover" />

								<!-- Status Badge -->
								<div class="absolute top-2 left-2">
									{#if photo.status === 'uploaded'}
										<span class="rounded-full bg-emerald-500/90 text-white p-1 flex items-center justify-center shadow">
											<CheckCircle2 class="h-3.5 w-3.5" />
										</span>
									{:else if photo.status === 'uploading'}
										<span class="rounded-full bg-primary text-white text-[10px] px-2 py-0.5 font-bold shadow animate-pulse">
											Загрузка...
										</span>
									{:else}
										<span class="rounded-full bg-destructive text-white text-[10px] px-2 py-0.5 font-bold shadow">
											Ошибка
										</span>
									{/if}
								</div>

								<!-- Remove Button -->
								<button
									type="button"
									onclick={(e) => {
										e.stopPropagation();
										removePhoto(photo.id || '');
									}}
									class="absolute top-2 right-2 rounded-full bg-black/60 p-1.5 text-white hover:bg-destructive transition-colors cursor-pointer opacity-80 group-hover:opacity-100"
									aria-label="Удалить фото"
								>
									<X class="h-3.5 w-3.5" />
								</button>
							</div>
						{/each}
					</div>
				{/if}

				<div class="flex justify-between pt-4 border-t border-border">
					<Button variant="outline" onclick={() => (currentStep = 2)}>
						<ArrowLeft class="h-4 w-4 mr-1" /> Назад
					</Button>
					<Button
						onclick={submitStep3}
						disabled={uploadedPhotosCount < 5}
						{loading}
						class="gap-2 font-semibold"
					>
						<span>Далее (Удобства)</span> <ArrowRight class="h-4 w-4" />
					</Button>
				</div>
			</div>
		{/if}

		<!-- STEP 4: Amenities -->
		{#if currentStep === 4}
			<div class="rounded-2xl border border-border bg-card p-6 sm:p-8 shadow-sm space-y-6">
				<div>
					<h2 class="text-xl font-bold text-foreground">Шаг 4: Удобства</h2>
					<p class="text-sm text-muted-foreground mt-1">
						Отметьте удобства, доступные гостям в вашем жилье.
					</p>
				</div>

				{#if loadingAmenities}
					<div class="py-12 text-center text-sm text-muted-foreground animate-pulse">
						Загрузка каталога удобств...
					</div>
				{:else if amenitiesCatalog}
					<div class="space-y-6">
						{#each amenitiesCatalog.categories as cat}
							<div class="rounded-2xl border border-border/80 p-4 bg-muted/10">
								<h4 class="font-bold text-sm text-foreground mb-3">{cat.name}</h4>
								<div class="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
									{#each cat.amenities as item}
										<button
											type="button"
											onclick={() => toggleAmenity(item.id)}
											class={cn(
												'flex items-center gap-2.5 p-2.5 rounded-xl border text-left text-xs font-medium transition-all cursor-pointer',
												selectedAmenities.has(item.id)
													? 'border-primary bg-primary/10 text-primary font-semibold'
													: 'border-border bg-card hover:bg-accent text-foreground'
											)}
										>
											<div
												class={cn(
													'h-4 w-4 rounded flex items-center justify-center border',
													selectedAmenities.has(item.id)
														? 'bg-primary border-primary text-white'
														: 'border-muted-foreground/40'
												)}
											>
												{#if selectedAmenities.has(item.id)}
													<CheckCircle2 class="h-3 w-3" />
												{/if}
											</div>
											<span>{item.name}</span>
										</button>
									{/each}
								</div>
							</div>
						{/each}
					</div>
				{/if}

				<div class="flex justify-between pt-4 border-t border-border">
					<Button variant="outline" onclick={() => (currentStep = 3)}>
						<ArrowLeft class="h-4 w-4 mr-1" /> Назад
					</Button>
					<Button onclick={submitStep4} {loading} class="gap-2 font-semibold">
						<span>Далее (Цены и правила)</span> <ArrowRight class="h-4 w-4" />
					</Button>
				</div>
			</div>
		{/if}

		<!-- STEP 5: Pricing & Rules -->
		{#if currentStep === 5}
			<div class="rounded-2xl border border-border bg-card p-6 sm:p-8 shadow-sm space-y-6">
				<div>
					<h2 class="text-xl font-bold text-foreground">Шаг 5: Цены и правила проживания</h2>
					<p class="text-sm text-muted-foreground mt-1">
						Установите стоимость аренды за сутки и настройте правила заезда.
					</p>
				</div>

				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<div class="space-y-1.5">
						<label for="w-price" class="text-xs font-semibold text-foreground">
							Цена за 1 сутки (BYN) <span class="text-destructive">*</span>
						</label>
						<div class="relative">
							<Banknote class="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
							<Input id="w-price" type="number" min="1" bind:value={pricePerNight} class="pl-9 font-bold" />
						</div>
					</div>

					<div class="space-y-1.5">
						<label for="w-min-stay" class="text-xs font-semibold text-foreground">
							Минимальный срок (суток)
						</label>
						<Input id="w-min-stay" type="number" min="1" max="365" bind:value={minNights} />
					</div>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div class="space-y-1.5">
						<label for="w-checkin" class="text-xs font-semibold text-foreground">Заезд с</label>
						<Input id="w-checkin" type="time" bind:value={checkinFrom} />
					</div>
					<div class="space-y-1.5">
						<label for="w-checkout" class="text-xs font-semibold text-foreground">Выезд до</label>
						<Input id="w-checkout" type="time" bind:value={checkoutUntil} />
					</div>
				</div>

				<!-- House Rules Switches -->
				<div class="space-y-3 pt-2">
					<h4 class="text-xs font-bold text-foreground uppercase tracking-wider">Правила дома</h4>
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
						<div class="flex items-center justify-between p-3 rounded-xl border border-border bg-muted/10">
							<span class="text-xs font-medium">Можно с детьми</span>
							<Switch bind:checked={rules.allow_children} />
						</div>
						<div class="flex items-center justify-between p-3 rounded-xl border border-border bg-muted/10">
							<span class="text-xs font-medium">Можно с питомцами</span>
							<Switch bind:checked={rules.allow_pets} />
						</div>
						<div class="flex items-center justify-between p-3 rounded-xl border border-border bg-muted/10">
							<span class="text-xs font-medium">Разрешено курение</span>
							<Switch bind:checked={rules.allow_smoking} />
						</div>
						<div class="flex items-center justify-between p-3 rounded-xl border border-border bg-muted/10">
							<span class="text-xs font-medium">Вечеринки / мероприятия</span>
							<Switch bind:checked={rules.allow_parties} />
						</div>
					</div>
				</div>

				<div class="flex justify-between pt-4 border-t border-border">
					<Button variant="outline" onclick={() => (currentStep = 4)}>
						<ArrowLeft class="h-4 w-4 mr-1" /> Назад
					</Button>
					<Button onclick={submitStep5} {loading} class="gap-2 font-semibold">
						<span>Далее (Описание)</span> <ArrowRight class="h-4 w-4" />
					</Button>
				</div>
			</div>
		{/if}

		<!-- STEP 6: Description -->
		{#if currentStep === 6}
			<div class="rounded-2xl border border-border bg-card p-6 sm:p-8 shadow-sm space-y-6">
				<div>
					<h2 class="text-xl font-bold text-foreground">Шаг 6: Описание объекта</h2>
					<p class="text-sm text-muted-foreground mt-1">
						Расскажите подробнее об атмосфере, спальных местах и расположении.
					</p>
				</div>

				<div class="space-y-2">
					<label for="w-desc" class="text-xs font-semibold text-foreground">
						Подробное описание <span class="text-destructive">*</span>
					</label>
					<Textarea
						id="w-desc"
						rows={7}
						placeholder="Опишите ваше жилье: светлые комнаты, современная техника, удобная транспортная развязка..."
						bind:value={description}
					/>
					<div class="flex justify-between text-[11px] text-muted-foreground">
						<span>Минимум 30 символов, максимум 5000</span>
						<span class={cn(description.trim().length < 30 && 'text-amber-600 font-bold')}>
							{description.trim().length} / 5000
						</span>
					</div>
				</div>

				<div class="flex justify-between pt-4 border-t border-border">
					<Button variant="outline" onclick={() => (currentStep = 5)}>
						<ArrowLeft class="h-4 w-4 mr-1" /> Назад
					</Button>
					<Button
						onclick={submitStep6}
						disabled={description.trim().length < 30}
						{loading}
						class="gap-2 font-semibold"
					>
						<span>Перейти к проверке</span> <ArrowRight class="h-4 w-4" />
					</Button>
				</div>
			</div>
		{/if}

		<!-- STEP 7: Review & Publish -->
		{#if currentStep === 7}
			<div class="rounded-2xl border border-border bg-card p-6 sm:p-8 shadow-sm space-y-6">
				<div>
					<h2 class="text-xl font-bold text-foreground">Шаг 7: Проверка и публикация</h2>
					<p class="text-sm text-muted-foreground mt-1">
						Проверьте указанные данные перед отправкой объявления на модерацию.
					</p>
				</div>

				<div class="rounded-2xl border border-border bg-muted/20 p-5 space-y-4 text-xs">
					<div class="flex justify-between py-1 border-b border-border/60">
						<span class="text-muted-foreground font-medium">Тип объекта:</span>
						<span class="font-bold text-foreground">{translateHousingType(housingType)}</span>
					</div>
					<div class="flex justify-between py-1 border-b border-border/60">
						<span class="text-muted-foreground font-medium">Название:</span>
						<span class="font-bold text-foreground text-right max-w-[280px]">{name}</span>
					</div>
					<div class="flex justify-between py-1 border-b border-border/60">
						<span class="text-muted-foreground font-medium">Параметры:</span>
						<span class="font-bold text-foreground">
							{square} м² • {roomsCount} комн. • {bedsCount} спальн. • {floor}/{totalFloors} эт.
						</span>
					</div>
					<div class="flex justify-between py-1 border-b border-border/60">
						<span class="text-muted-foreground font-medium">Фотографий загружено:</span>
						<span class="font-bold text-emerald-600">{uploadedPhotosCount} фото</span>
					</div>
					<div class="flex justify-between py-1 border-b border-border/60">
						<span class="text-muted-foreground font-medium">Удобств выбрано:</span>
						<span class="font-bold text-foreground">{selectedAmenities.size} шт.</span>
					</div>
					<div class="flex justify-between py-1 border-b border-border/60">
						<span class="text-muted-foreground font-medium">Цена за сутки:</span>
						<span class="font-extrabold text-foreground text-sm text-primary">
							{formatCurrency(pricePerNight, 'BYN')}
						</span>
					</div>
				</div>

				<div class="flex flex-wrap items-center justify-between gap-3 pt-4 border-t border-border">
					<div class="flex items-center gap-2">
						<Button variant="outline" onclick={() => (currentStep = 6)}>
							<ArrowLeft class="h-4 w-4 mr-1" /> Назад
						</Button>
						{#if draftId}
							<a href={`/host/new/preview?draft_id=${draftId}`} target="_blank">
								<Button variant="secondary" class="gap-1.5 text-xs">
									<span>Предпросмотр</span>
								</Button>
							</a>
						{/if}
					</div>
					<Button
						onclick={handleFinalPublish}
						{loading}
						class="bg-emerald-600 hover:bg-emerald-700 text-white font-bold gap-2 shadow-lg"
					>
						<Sparkles class="h-4 w-4" /> Опубликовать объявление
					</Button>
				</div>
			</div>
		{/if}
	{/if}
</div>
