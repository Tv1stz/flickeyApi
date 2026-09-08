<!-- src/lib/components/host-listing-edit/sections/EditRules.svelte -->
<script lang="ts">
	import CounterField from '$lib/components/form/CounterField.svelte';
	import { cardStyles, formStyles } from '$lib/config/styles';
	import { Clock, Users, FileText, Shield, Baby, Heart, Cigarette, Music } from 'lucide-svelte';
	import type { ListingEditFormData } from '../types';

	interface Props {
		form: Partial<ListingEditFormData>;
		onUpdate: (updates: Partial<ListingEditFormData>) => void;
	}

	let { form, onUpdate }: Props = $props();

	function updateRules(field: string, value: any) {
		const currentRules = form.rules || {
			checkIn: '14:00',
			checkOut: '12:00',
			minNights: 1,
			depositRequired: false,
			documentsProvided: false,
			childrenAllowed: true,
			petsAllowed: false,
			smokingAllowed: false,
			partiesAllowed: false
		};
		onUpdate({ rules: { ...currentRules, [field]: value } });
	}

	const timeOptions = ['10:00', '11:00', '12:00', '13:00', '14:00', '15:00', '16:00', '17:00', '18:00', '19:00', '20:00'];

	const checkboxRules = [
		{ id: 'childrenAllowed', label: 'Разрешено проживание с детьми', icon: Baby },
		{ id: 'petsAllowed', label: 'Разрешено с домашними животными', icon: Heart },
		{ id: 'smokingAllowed', label: 'Разрешено курение в помещении', icon: Cigarette },
		{ id: 'partiesAllowed', label: 'Разрешены шумные вечеринки и мероприятия', icon: Music },
		{ id: 'depositRequired', label: 'Требуется страховой депозит при заезде', icon: Shield },
		{ id: 'documentsProvided', label: 'Предоставляются отчетные документы', icon: FileText }
	];
</script>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Правила проживания</h2>
	<p class={cardStyles.subtitle}>Установите время заезда, выезда и требования к гостям</p>

	<div class={cardStyles.contentGap}>
		<div class={formStyles.fieldGroup}>
			<!-- Time Rules -->
			<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
				<div>
					<label for="check-in-select" class="mb-1.5 flex items-center gap-2 text-sm font-medium text-zinc-900">
						<Clock class="h-4 w-4 text-zinc-500" />
						Время заезда (с)
					</label>
					<select
						id="check-in-select"
						value={form.rules?.checkIn || '14:00'}
						onchange={(e) => updateRules('checkIn', e.currentTarget.value)}
						class="w-full rounded-2xl border border-zinc-300 bg-white px-4 py-3 text-sm text-zinc-900 transition-colors focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
					>
						{#each timeOptions as opt}
							<option value={opt}>{opt}</option>
						{/each}
					</select>
				</div>

				<div>
					<label for="check-out-select" class="mb-1.5 flex items-center gap-2 text-sm font-medium text-zinc-900">
						<Clock class="h-4 w-4 text-zinc-500" />
						Время выезда (до)
					</label>
					<select
						id="check-out-select"
						value={form.rules?.checkOut || '12:00'}
						onchange={(e) => updateRules('checkOut', e.currentTarget.value)}
						class="w-full rounded-2xl border border-zinc-300 bg-white px-4 py-3 text-sm text-zinc-900 transition-colors focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
					>
						{#each timeOptions as opt}
							<option value={opt}>{opt}</option>
						{/each}
					</select>
				</div>
			</div>

			<!-- Minimum Nights -->
			<div>
				<CounterField
					id="minNights"
					label="Минимальный срок бронирования (ночей)"
					min={1}
					max={30}
					value={String(form.rules?.minNights || 1)}
					onInput={(v) => updateRules('minNights', parseInt(v) || 1)}
					onBlur={() => {}}
				/>
			</div>

			<!-- Checkbox Rules -->
			<div class="space-y-3 pt-2">
				<p class="text-xs font-semibold uppercase tracking-wider text-zinc-400">Ограничения и условия</p>
				{#each checkboxRules as item}
					<label
						class="flex cursor-pointer items-center justify-between rounded-2xl border border-zinc-200 bg-white p-4 transition-colors hover:bg-zinc-50"
					>
						<div class="flex items-center gap-3">
							<div class="flex h-9 w-9 items-center justify-center rounded-xl bg-zinc-100 text-zinc-600">
								<item.icon class="h-4 w-4" />
							</div>
							<span class="text-sm font-medium text-zinc-900">{item.label}</span>
						</div>
						<input
							type="checkbox"
							checked={Boolean(form.rules?.[item.id as keyof NonNullable<typeof form.rules>])}
							onchange={(e) => updateRules(item.id, e.currentTarget.checked)}
							class="h-5 w-5 rounded-md border-zinc-300 text-zinc-900 focus:ring-zinc-900"
						/>
					</label>
				{/each}
			</div>
		</div>
	</div>
</section>
