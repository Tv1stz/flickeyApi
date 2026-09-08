<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import Map from '$lib/components/location/Map.svelte';
	import { cardStyles, formStyles, mapOverlayStyles } from '$lib/config/styles';
	import { locationStore } from '$lib/stores/location.svelte';
	import {
		reverseGeocode,
		searchAddress,
		haversineKm,
		formatDistance,
		type ParsedAddress
	} from '$lib/services/geocoding';
	import {
		getRecentAddresses,
		saveRecentAddress,
		clearRecentAddresses
	} from '$lib/services/localStorage';
	import {
		AlertCircle,
		LocateFixed,
		Search,
		X,
		Clock3,
		MapPin,
		Navigation,
		Loader2,
		House,
		Plus,
		Minus
	} from 'lucide-svelte';

	import type {
		ListingFormValues,
		ListingErrors,
		ListingTouched,
		ValidatedListingField
	} from '$lib/validation/listingValidation';

	interface Props {
		form: ListingFormValues;
		errors: ListingErrors;
		touched: ListingTouched;
		onInput: (
			field: Exclude<ValidatedListingField, 'photos' | 'contactMethods'>,
			value: string
		) => void;
		onBlur: (field: ValidatedListingField) => void;
	}

	let { form, errors, touched, onInput, onBlur }: Props = $props();

	let mapRef: {
		flyTo: (lat: number, lng: number, zoom?: number) => void;
		zoomIn: (step?: number) => void;
		zoomOut: (step?: number) => void;
	} | null = $state(null);

	let query = $state('');
	let results = $state<ParsedAddress[]>([]);
	let recents = $state<ParsedAddress[]>([]);
	let isSearching = $state(false);
	let isLocating = $state(false);
	let isOffline = $state(false);
	let locateError = $state('');
	let isFocused = $state(false);
	let showDropdown = $state(false);
	let showHouseInput = $state(false);
	let manualHouse = $state('');
	let hydratedFromFormAddress = $state(false);

	let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null;
	let blurTimer: ReturnType<typeof setTimeout> | null = null;
	let searchAbortController: AbortController | null = null;

	const SEARCH_DEBOUNCE_MS = 400;

	const selectedAddress = $derived(locationStore.address);
	const isMapBusy = $derived(locationStore.isLoading || locationStore.isMapMoving);
	const showRecents = $derived(isFocused && query.trim().length < 2 && recents.length > 0);
	const showResults = $derived(isFocused && query.trim().length >= 2 && results.length > 0);
	const showEmptyState = $derived(
		isFocused && query.trim().length >= 2 && !isSearching && !results.length
	);
	const dropdownOpen = $derived(showDropdown && (showRecents || showResults || showEmptyState));
	const primaryAddressLine = $derived(
		selectedAddress
			? [selectedAddress.street, selectedAddress.house].filter(Boolean).join(', ')
			: ''
	);
	const secondaryAddressLine = $derived(
		selectedAddress
			? [selectedAddress.city, selectedAddress.district].filter(Boolean).join(', ')
			: ''
	);

	onMount(() => {
		recents = getRecentAddresses();
		query = form.address;

		void hydrateLocationFromFormAddress();
	});

	onDestroy(() => {
		if (searchDebounceTimer) clearTimeout(searchDebounceTimer);
		if (blurTimer) clearTimeout(blurTimer);
		searchAbortController?.abort();
	});

	$effect(() => {
		const address = selectedAddress;
		if (!address) return;

		if (!isFocused) {
			query = address.displayName;
		}

		if (form.address !== address.displayName) {
			onInput('address', address.displayName);
		}

		if (address.house) {
			manualHouse = '';
			showHouseInput = false;
		}
	});

	async function hydrateLocationFromFormAddress(): Promise<void> {
		if (hydratedFromFormAddress) return;
		hydratedFromFormAddress = true;

		const existingAddress = form.address.trim();
		if (!existingAddress || selectedAddress?.displayName === existingAddress) return;

		const [firstMatch] = await searchAddress(existingAddress, {
			nearLat: locationStore.lat,
			nearLng: locationStore.lng
		});
		if (!firstMatch) return;

		setLocation(firstMatch, { saveToRecent: false, fly: true });
	}

	function setLocation(
		address: ParsedAddress,
		options: { saveToRecent?: boolean; fly?: boolean } = {}
	): void {
		locationStore.setError(null);
		locationStore.setAddress(address);
		locationStore.setCoords(address.lat, address.lng);

		onInput('address', address.displayName);
		if (touched.address) {
			onBlur('address');
		}

		if (options.saveToRecent !== false) {
			saveRecentAddress(address);
			recents = getRecentAddresses();
		}

		if (options.fly !== false) {
			mapRef?.flyTo(address.lat, address.lng, 16);
		}
	}

	async function runSearch(searchQuery: string): Promise<void> {
		const normalizedQuery = searchQuery.trim();
		if (normalizedQuery.length < 2) {
			results = [];
			showDropdown = true;
			isSearching = false;
			return;
		}

		if (typeof navigator !== 'undefined' && !navigator.onLine) {
			isOffline = true;
			results = [];
			isSearching = false;
			return;
		}

		searchAbortController?.abort();
		searchAbortController = new AbortController();
		isSearching = true;
		isOffline = false;

		results = await searchAddress(normalizedQuery, {
			nearLat: locationStore.lat,
			nearLng: locationStore.lng,
			signal: searchAbortController.signal
		});

		isSearching = false;
		showDropdown = true;
	}

	function handleSearchInput(event: Event): void {
		const input = event.currentTarget as HTMLInputElement;
		query = input.value;
		if (searchDebounceTimer) clearTimeout(searchDebounceTimer);
		searchDebounceTimer = setTimeout(() => {
			void runSearch(query);
		}, SEARCH_DEBOUNCE_MS);
	}

	function handleSearchFocus(): void {
		if (blurTimer) clearTimeout(blurTimer);
		isFocused = true;
		showDropdown = true;
	}

	function handleSearchBlur(): void {
		blurTimer = setTimeout(() => {
			isFocused = false;
			showDropdown = false;
			onBlur('address');
		}, 150);
	}

	function handleSearchKeydown(event: KeyboardEvent): void {
		if (event.key === 'Escape') {
			showDropdown = false;
			(event.currentTarget as HTMLInputElement).blur();
		}
	}

	function clearSearchQuery(): void {
		query = '';
		results = [];
		showDropdown = true;
		searchAbortController?.abort();
	}

	function selectAddress(address: ParsedAddress): void {
		query = address.displayName;
		results = [];
		showDropdown = false;
		setLocation(address, { saveToRecent: true, fly: true });
	}

	function distanceTo(address: ParsedAddress): string | null {
		if (!address.lat || !address.lng) return null;
		const km = haversineKm(locationStore.lat, locationStore.lng, address.lat, address.lng);
		return formatDistance(km);
	}

	function applyManualHouse(): void {
		const baseAddress = selectedAddress;
		const house = manualHouse.trim();
		if (!baseAddress || !house) return;

		const displayName = [baseAddress.city, baseAddress.street, house].filter(Boolean).join(', ');
		const patched: ParsedAddress = {
			...baseAddress,
			house,
			displayName: displayName || baseAddress.displayName,
			confidence: 'full'
		};

		setLocation(patched, { saveToRecent: true, fly: false });
		showHouseInput = false;
	}

	function clearRecentHistory(): void {
		clearRecentAddresses();
		recents = [];
	}

	function handleMapAddressResolved(): void {
		locateError = '';
	}

	function locateUser(): void {
		if (!navigator.geolocation) {
			locateError = 'Геолокация не поддерживается браузером';
			return;
		}

		isLocating = true;
		locateError = '';

		navigator.geolocation.getCurrentPosition(
			async ({ coords: { latitude, longitude } }) => {
				locationStore.setCoords(latitude, longitude);
				locationStore.setMapMoving(false);
				locationStore.setLoading(true);
				mapRef?.flyTo(latitude, longitude, 16);

				const resolved = await reverseGeocode(latitude, longitude);
				if (resolved) {
					setLocation(resolved, { saveToRecent: true, fly: false });
				} else {
					locationStore.setError('Не удалось определить адрес');
				}

				locationStore.setLoading(false);
				isLocating = false;
			},
			(error) => {
				locateError =
					error.code === GeolocationPositionError.PERMISSION_DENIED
						? 'Разрешите доступ к геолокации в настройках браузера'
						: error.code === GeolocationPositionError.TIMEOUT
							? 'Превышено время ожидания геолокации'
							: 'Не удалось получить местоположение';
				isLocating = false;
			},
			{ timeout: 8000, maximumAge: 30_000 }
		);
	}

	function zoomInMap(): void {
		mapRef?.zoomIn(1);
	}

	function zoomOutMap(): void {
		mapRef?.zoomOut(1);
	}
</script>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Укажите локацию</h2>
	<p class={cardStyles.subtitle}>Найдите адрес через поиск или поставьте точку на карте</p>

	<div class={cardStyles.contentGap}>
		<div class="relative">
			<div
				class="flex items-center gap-3 rounded-2xl border bg-white px-4 py-3 shadow-[0_8px_28px_rgba(24,24,27,0.07)] transition-all duration-200
					{isFocused
					? 'border-zinc-900'
					: touched.address && errors.address
						? 'border-red-300'
						: 'border-zinc-200'}"
			>
				{#if isSearching}
					<Loader2 class="h-5 w-5 shrink-0 animate-spin text-zinc-500" />
				{:else}
					<Search class="h-5 w-5 shrink-0 text-zinc-400" />
				{/if}

				<input
					id="address"
					type="search"
					value={query}
					placeholder="Например: ул. Ленина, 10"
					autocomplete="off"
					spellcheck={false}
					class="h-7 min-w-0 flex-1 bg-transparent text-[15px] font-medium text-zinc-900 outline-none placeholder:text-zinc-400 [&::-webkit-search-cancel-button]:hidden"
					oninput={handleSearchInput}
					onfocus={handleSearchFocus}
					onblur={handleSearchBlur}
					onkeydown={handleSearchKeydown}
				/>

				{#if query.trim().length > 0}
					<button
						type="button"
						onmousedown={(event) => event.preventDefault()}
						onclick={clearSearchQuery}
						class="flex h-7 w-7 touch-manipulation items-center justify-center rounded-full bg-zinc-100 text-zinc-500 transition-colors hover:bg-zinc-200"
						aria-label="Очистить поиск"
					>
						<X class="h-4 w-4" />
					</button>
				{/if}
			</div>

			{#if dropdownOpen}
				<div class="dropdown-surface absolute inset-x-0 top-full z-30 mt-2">
					{#if showRecents}
						<div class="flex items-center justify-between px-4 pt-3 pb-1">
							<p class="text-[11px] font-semibold tracking-wide text-zinc-400 uppercase">
								Недавние
							</p>
							<button
								type="button"
								onmousedown={(event) => event.preventDefault()}
								onclick={clearRecentHistory}
								class="touch-manipulation text-[11px] text-zinc-400 transition-colors hover:text-zinc-700"
							>
								Очистить
							</button>
						</div>

						<ul class="pb-1.5">
							{#each recents as address, index (`${address.displayName}-${address.lat}-${address.lng}`)}
								{@const dist = distanceTo(address)}
								<li>
									<button
										type="button"
										onmousedown={(event) => event.preventDefault()}
										onclick={() => selectAddress(address)}
										class="flex w-full touch-manipulation items-start gap-3 px-4 py-2.5 text-left transition-colors hover:bg-zinc-50
											{index < recents.length - 1 ? 'border-b border-zinc-100' : ''}"
									>
										<Clock3 class="mt-0.5 h-4 w-4 shrink-0 text-zinc-300" />
										<div class="min-w-0 flex-1">
											<p class="truncate text-[14px] font-medium text-zinc-700">
												{[address.street, address.house].filter(Boolean).join(', ') ||
													address.displayName}
											</p>
											<div class="mt-0.5 flex items-center justify-between gap-2">
												<p class="truncate text-[12px] text-zinc-400">{address.city}</p>
												{#if dist}
													<span class="shrink-0 text-[11px] font-medium text-zinc-500">{dist}</span>
												{/if}
											</div>
										</div>
									</button>
								</li>
							{/each}
						</ul>
					{/if}

					{#if showResults}
						<ul class="max-h-[280px] overflow-y-auto py-1.5">
							{#each results as address, index (`${address.displayName}-${address.lat}-${address.lng}`)}
								{@const dist = distanceTo(address)}
								<li>
									<button
										type="button"
										onmousedown={(event) => event.preventDefault()}
										onclick={() => selectAddress(address)}
										class="flex w-full touch-manipulation items-start gap-3 px-4 py-3 text-left transition-colors hover:bg-zinc-50
											{index < results.length - 1 ? 'border-b border-zinc-100' : ''}"
									>
										<MapPin class="mt-0.5 h-4 w-4 shrink-0 text-zinc-900" />
										<div class="min-w-0 flex-1">
											<p class="truncate text-[14px] font-semibold text-zinc-900">
												{[address.street, address.house].filter(Boolean).join(', ') ||
													address.displayName}
											</p>
											<div class="mt-0.5 flex items-center justify-between gap-2">
												<p class="truncate text-[12px] text-zinc-400">
													{[address.city, address.district].filter(Boolean).join(', ')}
												</p>
												{#if dist}
													<span class="shrink-0 text-[11px] font-medium text-zinc-500">{dist}</span>
												{/if}
											</div>
										</div>
									</button>
								</li>
							{/each}
						</ul>
					{/if}

					{#if showEmptyState}
						<div class="px-4 py-6 text-center">
							<p class="text-sm text-zinc-500">Ничего не найдено</p>
							<p class="mt-1 text-xs text-zinc-400">
								Попробуйте изменить запрос или передвинуть карту
							</p>
						</div>
					{/if}
				</div>
			{/if}
		</div>

		{#if isOffline}
			<div
				class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm text-amber-700"
			>
				Нет соединения с интернетом, поиск адреса временно недоступен.
			</div>
		{/if}

		<div class="relative overflow-hidden rounded-3xl border border-zinc-200 bg-zinc-100">
			<div class="relative h-[340px] sm:h-[400px] lg:h-[460px]">
				<Map bind:this={mapRef} onAddressResolved={handleMapAddressResolved} />

				<div
					class="pointer-events-none absolute inset-0 z-20 flex items-center justify-center"
					aria-hidden="true"
				>
					<div class="relative flex flex-col items-center">
						<div
							class="absolute bottom-0 h-1.5 w-4 rounded-full bg-black/20 blur-sm transition-all duration-300
								{locationStore.isMapMoving ? 'translate-y-3 scale-50 opacity-30' : 'scale-100 opacity-100'}"
						></div>
						<div
							class="transition-transform duration-300 {locationStore.isMapMoving
								? '-translate-y-4'
								: 'translate-y-0'}"
						>
							<svg width="38" height="50" viewBox="0 0 40 52" fill="none">
								<path
									d="M20 2C10.06 2 2 10.06 2 20C2 32.5 20 50 20 50C20 50 38 32.5 38 20C38 10.06 29.94 2 20 2Z"
									fill={locationStore.isLoading ? '#a1a1aa' : '#18181b'}
									stroke="white"
									stroke-width="2"
								/>
								<circle cx="20" cy="20" r="7" fill="white" />
							</svg>
						</div>
					</div>
				</div>

				<div class="absolute right-3 bottom-3 z-30 flex flex-col items-end gap-2">
					{#if locateError}
						<div class={mapOverlayStyles.infoBubble}>
							{locateError}
						</div>
					{/if}

					<div class="hidden lg:flex lg:flex-col lg:gap-2">
						<button
							type="button"
							onclick={zoomInMap}
							class={mapOverlayStyles.iconButton}
							aria-label="Приблизить карту"
						>
							<Plus class="h-5 w-5" />
						</button>
						<button
							type="button"
							onclick={zoomOutMap}
							class={mapOverlayStyles.iconButton}
							aria-label="Отдалить карту"
						>
							<Minus class="h-5 w-5" />
						</button>
					</div>

					<button
						type="button"
						onclick={locateUser}
						disabled={isLocating}
						class={mapOverlayStyles.iconButton}
						aria-label="Определить моё местоположение"
					>
						{#if isLocating}
							<Loader2 class="h-5 w-5 animate-spin" />
						{:else}
							<LocateFixed class="h-5 w-5" />
						{/if}
					</button>
				</div>
			</div>
		</div>

		<div class="rounded-2xl border border-zinc-200 bg-white p-4 sm:p-5">
			<p class="mb-2 text-xs tracking-wide text-zinc-400 uppercase">Выбранный адрес</p>

			{#if isMapBusy}
				<div class="flex items-center gap-2 text-sm text-zinc-500">
					<Loader2 class="h-4 w-4 animate-spin" />
					Определяем адрес...
				</div>
			{:else if locationStore.error}
				<div class={formStyles.errorBlock}>
					<AlertCircle class="h-5 w-5 shrink-0 text-red-500" />
					<p class="text-sm font-medium text-red-700">{locationStore.error}</p>
				</div>
			{:else if selectedAddress}
				<div class="space-y-3">
					<div class="flex items-start gap-3">
						<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-zinc-100">
							<Navigation class="h-4 w-4 text-zinc-900" />
						</div>
						<div class="min-w-0 flex-1">
							<p class="text-base font-semibold text-zinc-900">
								{primaryAddressLine || selectedAddress.displayName}
							</p>
							{#if secondaryAddressLine}
								<p class="mt-0.5 text-sm text-zinc-500">{secondaryAddressLine}</p>
							{/if}
						</div>
					</div>

					{#if selectedAddress.confidence !== 'full'}
						<div class="rounded-xl border border-amber-200 bg-amber-50 px-3.5 py-3">
							<p class="text-sm font-semibold text-amber-800">Номер дома не определён</p>
							<p class="mt-0.5 text-xs text-amber-700">
								Передвиньте карту точнее или добавьте номер дома вручную, чтобы адрес стал полным.
							</p>

							{#if !showHouseInput}
								<button
									type="button"
									onclick={() => (showHouseInput = true)}
									class="mt-2 inline-flex touch-manipulation rounded-lg bg-amber-100 px-3 py-1.5 text-xs font-semibold text-amber-900 transition-colors hover:bg-amber-200"
								>
									Добавить номер дома вручную
								</button>
							{:else}
								<div class="mt-2 grid gap-2 sm:flex sm:items-center">
									<House class="h-4 w-4 shrink-0 text-amber-700" />
									<input
										type="text"
										value={manualHouse}
										maxlength="10"
										placeholder="Например: 10А"
										class="h-9 min-w-0 flex-1 rounded-lg border border-amber-300 bg-white px-3 text-sm text-zinc-900 outline-none focus:border-amber-500"
										oninput={(event) =>
											(manualHouse = (event.currentTarget as HTMLInputElement).value)}
										onkeydown={(event) => {
											if (event.key === 'Enter') {
												event.preventDefault();
												applyManualHouse();
											}
											if (event.key === 'Escape') {
												showHouseInput = false;
											}
										}}
									/>
									<button
										type="button"
										onclick={applyManualHouse}
										disabled={!manualHouse.trim()}
										class="h-9 touch-manipulation rounded-lg bg-amber-500 px-3 text-xs font-semibold text-white transition-colors hover:bg-amber-600 disabled:cursor-not-allowed disabled:bg-amber-300"
									>
										Сохранить
									</button>
									<button
										type="button"
										onclick={() => (showHouseInput = false)}
										class="h-9 touch-manipulation rounded-lg border border-amber-300 bg-white px-3 text-xs font-semibold text-amber-800 transition-colors hover:bg-amber-100"
									>
										Отмена
									</button>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			{:else}
				<p class="text-sm text-zinc-500">Переместите карту или выберите адрес в поиске</p>
			{/if}
		</div>

		{#if touched.address && errors.address}
			<div class={formStyles.errorBlock}>
				<AlertCircle class="h-5 w-5 shrink-0 text-red-500" />
				<p class="text-sm font-medium text-red-700">{errors.address}</p>
			</div>
		{/if}
	</div>
</section>
