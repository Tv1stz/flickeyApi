<!-- src/lib/components/listing-page/ListingRules.svelte -->
<script lang="ts">
	import {
		LogIn,
		LogOut,
		Shield,
		FileText,
		Baby,
		PawPrint,
		Cigarette,
		PartyPopper
	} from 'lucide-svelte';
	import type { ListingRules } from '$lib/components/card/types';

	interface Props {
		rules: ListingRules;
	}

	let { rules }: Props = $props();

	type RuleItem = {
		icon: typeof LogIn;
		value: boolean;
		allowed: string;
		notAllowed: string;
	};

	const booleanRules = $derived.by((): RuleItem[] => [
		{
			icon: Baby,
			value: Boolean(rules.childrenAllowed),
			allowed: 'Можно с детьми',
			notAllowed: 'Без детей'
		},
		{
			icon: PawPrint,
			value: Boolean(rules.petsAllowed),
			allowed: 'Можно с питомцами',
			notAllowed: 'Без питомцев'
		},
		{
			icon: Cigarette,
			value: Boolean(rules.smokingAllowed),
			allowed: 'Курение разрешено',
			notAllowed: 'Не курить'
		},
		{
			icon: PartyPopper,
			value: Boolean(rules.partiesAllowed),
			allowed: 'Вечеринки разрешены',
			notAllowed: 'Без вечеринок'
		},
		{
			icon: Shield,
			value: Boolean(rules.depositRequired),
			allowed: 'Требуется залог',
			notAllowed: 'Залог не нужен'
		},
		{
			icon: FileText,
			value: Boolean(rules.documentsProvided),
			allowed: 'Документы предоставляются',
			notAllowed: 'Без документов'
		}
	]);
</script>

<div>
	<h2 class="mb-6 text-[20px] font-semibold tracking-tight text-zinc-900">Правила проживания</h2>

	<!-- Check-in / Check-out -->
	<div class="mb-6 grid grid-cols-2 gap-4">
		<div class="flex flex-col gap-1 rounded-3xl border border-zinc-400/40 bg-zinc-50/20 px-4 py-4">
			<div class="flex items-center gap-1.5">
				<LogIn size={15} strokeWidth={2} class="text-zinc-400" />
				<span class="text-[12px] font-medium tracking-wide text-zinc-400 uppercase">Заезд</span>
			</div>
			<p class="text-[18px] font-semibold text-zinc-900">с {rules.checkIn}</p>
		</div>

		<div class="flex flex-col gap-1 rounded-3xl border border-zinc-400/40 bg-zinc-50/20 px-4 py-4">
			<div class="flex items-center gap-1.5">
				<LogOut size={15} strokeWidth={2} class="text-zinc-400" />
				<span class="text-[12px] font-medium tracking-wide text-zinc-400 uppercase">Выезд</span>
			</div>
			<p class="text-[18px] font-semibold text-zinc-900">до {rules.checkOut}</p>
		</div>
	</div>

	<!-- Boolean rules -->
	<ul class="grid grid-cols-2 gap-x-4 gap-y-3">
		{#each booleanRules as rule (rule.allowed)}
			{@const RuleIcon = rule.icon}
			<li class="flex items-center gap-2.5">
				<RuleIcon
					size={18}
					strokeWidth={1.75}
					class="shrink-0 {rule.value ? 'text-zinc-600' : 'text-zinc-300'}"
				/>
				<span class="text-[14px] {rule.value ? 'text-zinc-800' : 'text-zinc-400 line-through'}">
					{rule.value ? rule.allowed : rule.notAllowed}
				</span>
			</li>
		{/each}
	</ul>
</div>
