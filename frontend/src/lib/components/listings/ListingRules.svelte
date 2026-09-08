<script lang="ts">
	import { LogIn, LogOut, Baby, PawPrint, Cigarette, PartyPopper, ShieldAlert, FileText } from 'lucide-svelte';
	import type { ListingRules as ListingRulesType } from '$lib/types/listings';

	interface Props {
		checkinFrom?: string;
		checkoutUntil?: string;
		minNights?: number;
		rules?: ListingRulesType;
		allowChildren?: boolean;
		allowPets?: boolean;
		allowSmoking?: boolean;
		allowParties?: boolean;
		depositRequired?: boolean;
		withInvoicing?: boolean;
	}

	let {
		checkinFrom = '14:00',
		checkoutUntil = '12:00',
		rules,
		allowChildren,
		allowPets,
		allowSmoking,
		allowParties,
		depositRequired,
		withInvoicing
	}: Props = $props();

	function formatTime(val?: string, defaultVal: string = '12:00') {
		if (!val) return defaultVal;
		const parts = val.split(':');
		if (parts.length >= 2) {
			return `${parts[0]}:${parts[1]}`;
		}
		return val;
	}

	let isChildrenAllowed = $derived(rules?.allow_children ?? allowChildren ?? true);
	let isPetsAllowed = $derived(rules?.allow_pets ?? allowPets ?? false);
	let isSmokingAllowed = $derived(rules?.allow_smoking ?? allowSmoking ?? false);
	let isPartiesAllowed = $derived(rules?.allow_parties ?? allowParties ?? false);
	let isDepositRequired = $derived(rules?.deposit_required ?? depositRequired ?? false);
	let isWithInvoicing = $derived(rules?.with_invoicing ?? withInvoicing ?? false);
</script>

<div class="space-y-4 pt-6 border-t border-border/60">
	<h2 class="text-base sm:text-lg font-bold text-foreground">Правила проживания</h2>

	<!-- Checkin / Checkout Divided Card matching screenshot -->
	<div class="grid grid-cols-2 rounded-2xl border border-border/80 overflow-hidden bg-card divide-x divide-border/80">
		<div class="p-3.5 sm:p-4 space-y-1">
			<div class="flex items-center gap-1.5 text-[10px] font-bold tracking-wider text-muted-foreground uppercase">
				<LogIn class="h-3.5 w-3.5" />
				<span>Заезд</span>
			</div>
			<div class="text-xs sm:text-sm font-bold text-foreground">
				с {formatTime(checkinFrom, '14:00')}
			</div>
		</div>

		<div class="p-3.5 sm:p-4 space-y-1">
			<div class="flex items-center gap-1.5 text-[10px] font-bold tracking-wider text-muted-foreground uppercase">
				<LogOut class="h-3.5 w-3.5" />
				<span>Выезд</span>
			</div>
			<div class="text-xs sm:text-sm font-bold text-foreground">
				до {formatTime(checkoutUntil, '12:00')}
			</div>
		</div>
	</div>

	<!-- 2-Column Rules Grid matching screenshot -->
	<div class="grid grid-cols-1 sm:grid-cols-2 gap-y-3 gap-x-6 pt-1 text-xs">
		<!-- Children -->
		<div class="flex items-center gap-2 {isChildrenAllowed ? 'text-foreground/80' : 'text-muted-foreground'}">
			<Baby class="h-4 w-4 {isChildrenAllowed ? 'text-foreground/70' : 'text-muted-foreground/60'} shrink-0" />
			<span>{isChildrenAllowed ? 'Можно с детьми' : 'Без детей'}</span>
		</div>

		<!-- Pets -->
		<div class="flex items-center gap-2 {isPetsAllowed ? 'text-foreground/80' : 'text-muted-foreground'}">
			<PawPrint class="h-4 w-4 {isPetsAllowed ? 'text-foreground/70' : 'text-muted-foreground/60'} shrink-0" />
			<span>{isPetsAllowed ? 'Можно с питомцами' : 'Без питомцев'}</span>
		</div>

		<!-- Smoking -->
		<div class="flex items-center gap-2 {isSmokingAllowed ? 'text-foreground/80' : 'text-muted-foreground'}">
			<Cigarette class="h-4 w-4 {isSmokingAllowed ? 'text-foreground/70' : 'text-muted-foreground/60'} shrink-0" />
			<span>{isSmokingAllowed ? 'Курение разрешено' : 'Курение запрещено'}</span>
		</div>

		<!-- Parties -->
		<div class="flex items-center gap-2 {isPartiesAllowed ? 'text-foreground/80' : 'text-muted-foreground'}">
			<PartyPopper class="h-4 w-4 {isPartiesAllowed ? 'text-foreground/70' : 'text-muted-foreground/60'} shrink-0" />
			<span>{isPartiesAllowed ? 'Вечеринки разрешены' : 'Без вечеринок'}</span>
		</div>

		<!-- Deposit -->
		<div class="flex items-center gap-2 {isDepositRequired ? 'text-foreground/80' : 'text-muted-foreground'}">
			<ShieldAlert class="h-4 w-4 {isDepositRequired ? 'text-foreground/70' : 'text-muted-foreground/60'} shrink-0" />
			<span>{isDepositRequired ? 'Требуется залог' : 'Без залога'}</span>
		</div>

		<!-- Documents / Invoicing -->
		<div class="flex items-center gap-2 {isWithInvoicing ? 'text-foreground/80' : 'text-muted-foreground'}">
			<FileText class="h-4 w-4 {isWithInvoicing ? 'text-foreground/70' : 'text-muted-foreground/60'} shrink-0" />
			<span>{isWithInvoicing ? 'Отчетные документы' : 'Без документов'}</span>
		</div>
	</div>
</div>
