<script lang="ts">
	import TextInputField from '$lib/components/form/TextInputField.svelte';
	import SelectField from '$lib/components/form/SelectField.svelte';
	import ToggleSwitch from '$lib/components/ui/ToggleSwitch.svelte';
	import { Baby, Dog, Cigarette, PartyPopper, Shield, FileText } from 'lucide-svelte';
	import { cardStyles } from '$lib/config/styles';

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
		onToggleBoolean: (
			field:
				| 'allowChildren'
				| 'allowPets'
				| 'smokingAllowed'
				| 'partiesAllowed'
				| 'securityDepositRequired'
				| 'reportingDocuments',
			value: boolean
		) => void;
	}

	let { form, errors, touched, onInput, onBlur, onToggleBoolean }: Props = $props();

	function buildTimeOptions() {
		const options = [];
		for (let m = 0; m <= 23 * 60; m += 30) {
			const hh = String(Math.floor(m / 60)).padStart(2, '0');
			const mm = String(m % 60).padStart(2, '0');
			options.push({ value: `${hh}:${mm}`, label: `${hh}:${mm}` });
		}
		return options;
	}

	const timeOptions = buildTimeOptions();

	const rules = [
		{ field: 'allowChildren' as const, Icon: Baby, label: 'Можно с детьми' },
		{ field: 'allowPets' as const, Icon: Dog, label: 'Можно с питомцами' },
		{ field: 'smokingAllowed' as const, Icon: Cigarette, label: 'Можно курить' },
		{ field: 'partiesAllowed' as const, Icon: PartyPopper, label: 'Можно вечеринки' },
		{ field: 'securityDepositRequired' as const, Icon: Shield, label: 'Требуется залог' },
		{ field: 'reportingDocuments' as const, Icon: FileText, label: 'Отчётные документы' }
	];
</script>

<section class={cardStyles.sectionDivided}>
	<h2 class={cardStyles.title}>Установите цену</h2>
	<p class={cardStyles.subtitle}>Вы сможете изменить её в любое время</p>

	<div class={cardStyles.contentGap}>
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-5">
			<TextInputField
				id="basePrice"
				label="Цена за ночь"
				required
				numeric
				placeholder="120"
				suffix="BYN"
				value={form.basePrice}
				error={errors.basePrice}
				touched={touched.basePrice}
				valid={touched.basePrice && !errors.basePrice && form.basePrice !== ''}
				onInput={(v) => onInput('basePrice', v)}
				onBlur={() => onBlur('basePrice')}
			/>

			<TextInputField
				id="minNights"
				label="Минимум ночей"
				required
				numeric
				placeholder="1"
				value={form.minNights}
				error={errors.minNights}
				touched={touched.minNights}
				valid={touched.minNights && !errors.minNights && form.minNights !== ''}
				onInput={(v) => onInput('minNights', v)}
				onBlur={() => onBlur('minNights')}
			/>
		</div>
	</div>
</section>

<section class={cardStyles.sectionDivided}>
	<h2 class={cardStyles.title}>Время заезда и выезда</h2>
	<p class={cardStyles.subtitle}>Когда гости смогут заселиться и когда должны выехать</p>

	<div class={cardStyles.contentGap}>
		<div class="grid grid-cols-2 gap-4 sm:gap-5">
			<div class="space-y-2">
				<label for="checkInFrom" class="block text-sm font-medium text-zinc-700">Заезд с</label>
				<SelectField
					id="checkInFrom"
					label="Заезд с"
					compact
					required
					value={form.checkInFrom}
					options={timeOptions}
					error={errors.checkInFrom}
					touched={touched.checkInFrom}
					onInput={(v) => onInput('checkInFrom', v)}
					onBlur={() => onBlur('checkInFrom')}
				/>
			</div>

			<div class="space-y-2">
				<label for="checkOutBefore" class="block text-sm font-medium text-zinc-700">Выезд до</label>
				<SelectField
					id="checkOutBefore"
					label="Выезд до"
					compact
					required
					value={form.checkOutBefore}
					options={timeOptions}
					error={errors.checkOutBefore}
					touched={touched.checkOutBefore}
					onInput={(v) => onInput('checkOutBefore', v)}
					onBlur={() => onBlur('checkOutBefore')}
				/>
			</div>
		</div>
	</div>
</section>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Правила размещения</h2>
	<p class={cardStyles.subtitle}>Что разрешено и что требуется от гостей</p>

	<div class={cardStyles.contentGap}>
		<div class="space-y-3">
			{#each rules as rule (rule.field)}
				{@const isActive = form[rule.field]}
				{@const RuleIcon = rule.Icon}
				<div
					class="flex w-full cursor-pointer touch-manipulation items-center justify-between gap-4 rounded-3xl border p-4 transition-all duration-200 select-none sm:p-5
						{isActive ? 'border-zinc-900 bg-zinc-50' : 'border-zinc-200 bg-white hover:border-zinc-300'}"
					role="switch"
					aria-checked={isActive}
					aria-label={rule.label}
					tabindex="0"
					onclick={() => onToggleBoolean(rule.field, !isActive)}
					onkeydown={(e) => {
						if (e.key === 'Enter' || e.key === ' ') {
							e.preventDefault();
							onToggleBoolean(rule.field, !isActive);
						}
					}}
				>
					<div class="flex items-center gap-4">
						<div
							class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg
								{isActive ? 'bg-zinc-900 text-white' : 'bg-zinc-100 text-zinc-600'}"
						>
							<RuleIcon class="h-5 w-5" />
						</div>
						<span class="text-base font-medium text-zinc-900">{rule.label}</span>
					</div>

					<ToggleSwitch
						checked={isActive}
						label={rule.label}
						onToggle={(v) => onToggleBoolean(rule.field, v)}
					/>
				</div>
			{/each}
		</div>
	</div>
</section>
