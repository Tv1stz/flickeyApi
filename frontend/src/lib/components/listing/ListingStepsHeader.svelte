<!-- src/lib/components/listing/ListingStepsHeader.svelte -->
<script lang="ts">
	interface Props {
		currentStep?: number;
		totalSteps?: number;
		stepLabels?: string[];
		isEditMode?: boolean;
		onStepClick?: (step: number) => void;
	}

	let {
		currentStep = 1,
		totalSteps = 4,
		stepLabels = [],
		isEditMode = false,
		onStepClick
	}: Props = $props();

	const safeTotal = $derived(Math.max(1, totalSteps));
	const safeCurrent = $derived(Math.max(1, Math.min(currentStep, safeTotal)));

	// Прогресс от 0 до 100%
	// На шаге 1 из 4 → 0%, на 2/4 → 33%, на 3/4 → 66%, на 4/4 → 100%
	const progressPercent = $derived(
		safeTotal <= 1 ? 100 : Math.round(((safeCurrent - 1) / (safeTotal - 1)) * 100)
	);
	const currentLabel = $derived(stepLabels[safeCurrent - 1] ?? '');
</script>

<!--
	Единая плавная полоска прогресса — как у Airbnb.
	Переходит от 0% (шаг 1) до 100% (последний шаг) плавно.
-->
<div
	role="progressbar"
	aria-label="Прогресс создания объявления"
	aria-valuemin={0}
	aria-valuemax={100}
	aria-valuenow={progressPercent}
	aria-valuetext={currentLabel
		? `Шаг ${safeCurrent} из ${safeTotal}: ${currentLabel}`
		: `Шаг ${safeCurrent} из ${safeTotal}`}
	class="relative h-[3px] w-full bg-zinc-100"
>
	<div
		class="absolute inset-y-0 left-0 bg-zinc-900 transition-[width] duration-500 ease-out"
		style:width="{progressPercent}%"
	></div>
</div>
