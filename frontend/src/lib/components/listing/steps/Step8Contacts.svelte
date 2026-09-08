<!-- Step8Contacts.svelte -->
<script lang="ts">
	import { fade, fly, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { cardStyles, formStyles } from '$lib/config/styles';
	import { Phone, Mail, MessageCircle, Plus, X, Check, AlertCircle } from 'lucide-svelte';
	import type { ContactMethod, ContactMethodType } from '$lib/components/card/types';
	import type {
		ListingFormValues,
		ListingErrors,
		ListingTouched,
		ValidatedListingField
	} from '$lib/validation/listingValidation';

	import {
		filterContactInput,
		formatContactDisplay,
		validateContactValue,
		getContactHint,
		getContactInputMode,
		getContactMaxLength,
		isKeyAllowed
	} from '$lib/utils/contactFormatters';

	interface Props {
		form: ListingFormValues;
		errors: ListingErrors;
		touched: ListingTouched;
		onBlur: (field: ValidatedListingField) => void;
		onUpdateContactMethods: (methods: ContactMethod[]) => void;
	}

	let { form, errors, touched, onBlur, onUpdateContactMethods }: Props = $props();

	interface MessengerConfig {
		type: ContactMethodType;
		label: string;
		icon: typeof Mail;
		color: string;
		bgColor: string;
	}

	const MESSENGER_CONFIGS: MessengerConfig[] = [
		{ type: 'email', label: 'Email', icon: Mail, color: 'text-rose-600', bgColor: 'bg-rose-100' },
		{
			type: 'telegram',
			label: 'Telegram',
			icon: MessageCircle,
			color: 'text-sky-600',
			bgColor: 'bg-sky-100'
		},
		{
			type: 'whatsapp',
			label: 'WhatsApp',
			icon: MessageCircle,
			color: 'text-emerald-600',
			bgColor: 'bg-emerald-100'
		},
		{
			type: 'viber',
			label: 'Viber',
			icon: MessageCircle,
			color: 'text-violet-600',
			bgColor: 'bg-violet-100'
		},
		{
			type: 'signal',
			label: 'Signal',
			icon: MessageCircle,
			color: 'text-blue-600',
			bgColor: 'bg-blue-100'
		}
	];

	const messengerConfigMap = new Map(MESSENGER_CONFIGS.map((c) => [c.type, c]));

	let focusedField = $state<ContactMethodType | null>(null);
	let contactErrors = $state<Record<ContactMethodType, string>>({
		email: '',
		telegram: '',
		whatsapp: '',
		viber: '',
		signal: ''
	});

	const addedTypes = $derived(new Set(form.contactMethods.map((m) => m.type)));
	const availableToAdd = $derived(MESSENGER_CONFIGS.filter((m) => !addedTypes.has(m.type)));

	function getConfig(type: ContactMethodType): MessengerConfig {
		return messengerConfigMap.get(type)!;
	}

	function getFieldState(type: ContactMethodType, value: string) {
		const error = contactErrors[type];
		const hasValue = value.trim().length > 0;
		const isValid = hasValue && !error;
		const hasError = hasValue && Boolean(error);
		return { hasValue, isValid, hasError, error, isFocused: focusedField === type };
	}

	function updateMethod(type: ContactMethodType, value: string) {
		onUpdateContactMethods(form.contactMethods.map((m) => (m.type === type ? { ...m, value } : m)));
	}

	function addMethod(type: ContactMethodType) {
		onUpdateContactMethods([...form.contactMethods, { type, value: '' }]);
		contactErrors[type] = '';
		setTimeout(() => document.getElementById(`contact-${type}`)?.focus(), 100);
	}

	function removeMethod(type: ContactMethodType) {
		onUpdateContactMethods(form.contactMethods.filter((m) => m.type !== type));
		contactErrors[type] = '';
	}

	function handleInput(type: ContactMethodType, event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const cursorPos = input.selectionStart || 0;
		const filtered = filterContactInput(type, input.value);
		const formatted = formatContactDisplay(type, filtered);
		if (input.value !== formatted) {
			const diff = formatted.length - input.value.length;
			input.value = formatted;
			const newPos = Math.max(0, Math.min(formatted.length, cursorPos + diff));
			requestAnimationFrame(() => input.setSelectionRange(newPos, newPos));
		}
		updateMethod(type, formatted);
		contactErrors[type] = validateContactValue(type, formatted);
	}

	function handlePaste(type: ContactMethodType, event: ClipboardEvent) {
		event.preventDefault();
		const input = event.currentTarget as HTMLInputElement;
		const text = event.clipboardData?.getData('text') || '';
		const formatted = formatContactDisplay(type, filterContactInput(type, text));
		const truncated = formatted.slice(0, getContactMaxLength(type));
		input.value = truncated;
		input.setSelectionRange(truncated.length, truncated.length);
		updateMethod(type, truncated);
		contactErrors[type] = validateContactValue(type, truncated);
	}

	function handleKeydown(type: ContactMethodType, event: KeyboardEvent) {
		if (event.ctrlKey || event.metaKey) return;
		const input = event.currentTarget as HTMLInputElement;
		if (!isKeyAllowed(type, event.key, input.value, input.selectionStart || 0)) {
			event.preventDefault();
		}
	}

	function handleBlur(type: ContactMethodType) {
		focusedField = null;
		const method = form.contactMethods.find((m) => m.type === type);
		if (method) {
			const formatted = formatContactDisplay(type, method.value);
			if (formatted !== method.value) updateMethod(type, formatted);
			contactErrors[type] = validateContactValue(type, formatted);
		}
		onBlur('contactMethods');
	}
</script>

<!-- Section: Main Phone -->
<section class={cardStyles.sectionDivided}>
	<h2 class={cardStyles.title}>Основной номер телефона</h2>
	<p class={cardStyles.subtitle}>Этот номер привязан к вашему аккаунту и будет виден гостям</p>

	<div class={cardStyles.contentGap}>
		<div class="flex items-center gap-4 rounded-3xl border-1 border-zinc-900 bg-zinc-50 p-4 sm:p-5">
			<div
				class="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-zinc-900 sm:h-14 sm:w-14"
			>
				<Phone class="h-6 w-6 text-white" />
			</div>

			<div class="min-w-0 flex-1">
				<p class="mb-0.5 text-xs font-medium text-zinc-500 uppercase">Подтверждённый номер</p>
				<p class="text-lg font-bold text-zinc-900 sm:text-xl">
					{form.phone || '+375 29 XXX XX XX'}
				</p>
			</div>

			<div class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-green-500">
				<Check class="h-4 w-4 text-white" strokeWidth={3} />
			</div>
		</div>
	</div>
</section>

<!-- Section: Additional Contacts -->
<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Дополнительные способы связи</h2>
	<p class={cardStyles.subtitle}>Добавьте мессенджеры или email для удобства гостей</p>

	<div class={cardStyles.contentGap}>
		<!-- Added Contacts -->
		{#if form.contactMethods.length > 0}
			<div class="space-y-4">
				{#each form.contactMethods as method, index (method.type)}
					{@const config = getConfig(method.type)}
					{@const state = getFieldState(method.type, method.value)}
					{@const ContactIcon = config.icon}

					<div
						class="group"
						in:fly={{ y: 20, duration: 300, delay: index * 50, easing: cubicOut }}
						out:fly={{ x: -20, duration: 200 }}
					>
						<div class="flex items-center gap-3 sm:gap-4">
							<div
								class="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl sm:h-14 sm:w-14 {config.bgColor}"
							>
								<ContactIcon class="h-6 w-6 {config.color}" />
							</div>

							<div class="relative min-w-0 flex-1">
								<input
									id={`contact-${method.type}`}
									type="text"
									inputmode={getContactInputMode(method.type)}
									autocomplete="off"
									autocapitalize="off"
									spellcheck="false"
									maxlength={getContactMaxLength(method.type)}
									class="block h-12 w-full touch-manipulation rounded-xl border-1 bg-white
                                            px-4 text-base text-zinc-900 transition-all
                                            duration-200 placeholder:text-zinc-400 focus:outline-none
                                            sm:h-14
                                            {state.hasError
										? 'border-red-300 focus:border-red-500'
										: state.isValid
											? 'border-green-300 focus:border-green-500'
											: 'border-zinc-200 focus:border-zinc-900'}"
									placeholder={`Введите ${config.label.toLowerCase()}`}
									value={method.value}
									oninput={(e) => handleInput(method.type, e)}
									onpaste={(e) => handlePaste(method.type, e)}
									onkeydown={(e) => handleKeydown(method.type, e)}
									onfocus={() => (focusedField = method.type)}
									onblur={() => handleBlur(method.type)}
								/>

								{#if state.isValid}
									<div
										class="absolute top-1/2 right-4 -translate-y-1/2"
										in:scale={{ duration: 150 }}
									>
										<div class="flex h-6 w-6 items-center justify-center rounded-full bg-green-500">
											<Check class="h-4 w-4 text-white" strokeWidth={3} />
										</div>
									</div>
								{:else if state.hasError}
									<div
										class="absolute top-1/2 right-4 -translate-y-1/2"
										in:scale={{ duration: 150 }}
									>
										<AlertCircle class="h-5 w-5 text-red-500" />
									</div>
								{/if}
							</div>

							<button
								type="button"
								class="flex h-10 w-10 shrink-0 touch-manipulation items-center justify-center
                                        rounded-xl text-zinc-400 transition-all
                                        duration-200 hover:bg-red-50 hover:text-red-600"
								onclick={() => removeMethod(method.type)}
								aria-label="Удалить {config.label}"
							>
								<X class="h-5 w-5" />
							</button>
						</div>

						<div class="mt-2 ml-16 min-h-[20px] sm:ml-[72px]">
							{#if state.hasError}
								<p class="text-sm font-medium text-red-600" in:fade={{ duration: 150 }}>
									{state.error}
								</p>
							{:else if state.isFocused || !state.hasValue}
								<p class="text-sm text-zinc-400" in:fade={{ duration: 150 }}>
									{getContactHint(method.type)}
								</p>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}

		<!-- Validation Error -->
		{#if touched.contactMethods && errors.contactMethods}
			<div class={formStyles.errorBlock} in:fade={{ duration: 200 }}>
				<AlertCircle class="h-5 w-5 shrink-0 text-red-500" />
				<p class="text-sm font-medium text-red-700">{errors.contactMethods}</p>
			</div>
		{/if}

		<!-- Add Buttons -->
		{#if availableToAdd.length > 0}
			<div class="flex flex-wrap gap-2">
				{#each availableToAdd as messenger (messenger.type)}
					{@const MessengerIcon = messenger.icon}
					<button
						type="button"
						class="inline-flex touch-manipulation items-center gap-2.5 rounded-full border-1
                                border-zinc-200 bg-white px-4 py-2.5 transition-all
                                duration-200 hover:border-zinc-500 active:bg-zinc-50"
						onclick={() => addMethod(messenger.type)}
					>
						<div class="flex h-6 w-6 items-center justify-center rounded-md {messenger.bgColor}">
							<MessengerIcon class="h-3.5 w-3.5 {messenger.color}" />
						</div>
						<span class="text-sm font-medium text-zinc-700">{messenger.label}</span>
						<Plus class="h-4 w-4 text-zinc-400" />
					</button>
				{/each}
			</div>
		{/if}

		<!-- Empty State -->
		{#if form.contactMethods.length === 0 && availableToAdd.length === MESSENGER_CONFIGS.length}
			<div
				class="rounded-3xl border-1 border-dashed border-zinc-300 bg-zinc-50 px-6 py-10 text-center"
				in:fade={{ duration: 300 }}
			>
				<div
					class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-zinc-200"
				>
					<MessageCircle class="h-7 w-7 text-zinc-500" />
				</div>
				<p class="mb-1 text-base font-semibold text-zinc-700">
					Дополнительные контакты не добавлены
				</p>
				<p class="text-sm text-zinc-500">
					Нажмите на кнопку выше, чтобы добавить мессенджер или email
				</p>
			</div>
		{/if}
	</div>
</section>
