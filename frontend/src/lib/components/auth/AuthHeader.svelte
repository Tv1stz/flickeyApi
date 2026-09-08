<!-- src/lib/components/auth/AuthHeader.svelte -->
<script lang="ts">
	import { X, ArrowLeft } from 'lucide-svelte';
	import { resolve } from '$app/paths';

	interface Props {
		progress: number;
		showBackButton?: boolean;
		onBack: () => void;
		onClose: () => void;
	}

	let { progress, showBackButton = true, onBack, onClose }: Props = $props();

	const isFirstStep = $derived(!showBackButton);
</script>

<header class="safe-top sticky top-0 z-10 border-b border-zinc-100 bg-white">
	<!-- Navigation -->
	<div class="mx-auto flex h-14 w-full max-w-lg items-center justify-between px-4">
		<button
			type="button"
			onclick={isFirstStep ? onClose : onBack}
			class="-ml-2 flex h-10 w-10 items-center justify-center rounded-full
                   text-zinc-600 transition-colors duration-200
                   hover:bg-zinc-100 active:scale-95 active:bg-zinc-200"
			aria-label={isFirstStep ? 'Закрыть' : 'Назад'}
		>
			<div
				class="transition-transform duration-200"
				style="transform: rotate({isFirstStep ? 0 : 0}deg);"
			>
				{#if isFirstStep}
					<X class="h-5 w-5" />
				{:else}
					<ArrowLeft class="h-5 w-5" />
				{/if}
			</div>
		</button>

		<a href={resolve('/')} class="text-lg font-bold tracking-tight text-zinc-900">
			<img src="/logo.svg" alt="Flickey" class="h-6 w-auto" />
		</a>

		<div class="w-10"></div>
	</div>

	<!-- Progress bar -->
	<div class="h-1 overflow-hidden bg-zinc-100">
		<div
			class="h-full bg-zinc-900"
			style="width: {progress}%; transition: width 0.4s cubic-bezier(0.16, 1, 0.3, 1);"
		></div>
	</div>
</header>
