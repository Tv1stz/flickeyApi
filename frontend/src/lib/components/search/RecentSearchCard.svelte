<script lang="ts">
	import { ArrowRight, Clock } from 'lucide-svelte';
	import type { RecentSearch } from './taxonomy';

	interface Props {
		search: RecentSearch | null;
		onClick?: () => void;
	}

	let { search, onClick }: Props = $props();

	function formatGuests(adults: number, children: number): string {
		const total = adults + children;
		if (total === 0) return 'Любое кол-во гостей';
		if (total === 1) return '1 гость';
		if (total >= 2 && total <= 4) return `${total} гостя`;
		return `${total} гостей`;
	}

	function getPropertyTypeLabel(type: string): string {
		switch (type) {
			case 'apartment': return 'Квартира';
			case 'house': return 'Дом';
			case 'estate': return 'Усадьба';
			default: return 'Любое жилье';
		}
	}
</script>

{#if search}
	<button
		type="button"
		onclick={onClick}
		class="group relative flex w-full items-center gap-4 rounded-[2rem] border border-zinc-200/80 bg-white/90 p-2.5 pr-5
          shadow-[0_8px_20px_rgb(0,0,0,0.04)] backdrop-blur-xl
          transition-all duration-300 ease-out
          hover:-translate-y-0.5 hover:border-zinc-300 hover:bg-white hover:shadow-[0_12px_30px_rgb(0,0,0,0.08)]
          active:scale-[0.98] sm:w-auto sm:min-w-[340px] cursor-pointer text-left"
	>
		<!-- Thumbnail -->
		<div
			class="relative h-14 w-14 shrink-0 overflow-hidden rounded-[1.25rem] bg-zinc-100 sm:h-16 sm:w-16"
		>
			<div class="absolute inset-0 z-10 rounded-[1.25rem] ring-1 ring-inset ring-black/5 transition-colors group-hover:ring-black/10"></div>
			<img
				src="https://images.unsplash.com/photo-1560448204-e02f11c3d0e2?w=200&h=200&fit=crop"
				alt={search.params.location}
				class="h-full w-full object-cover"
			/>
		</div>

		<!-- Content -->
		<div class="min-w-0 flex-1 text-left">
			<!-- Надзаголовок -->
			<div class="flex items-center gap-1 text-zinc-400">
				<Clock size={12} strokeWidth={2.5} />
				<span class="text-[12px] font-bold tracking-tight">Продолжить поиск</span>
			</div>

			<!-- Главный заголовок -->
			<h4 class="mt-0.5 truncate text-base font-bold tracking-tight text-zinc-900 sm:text-lg">
				{search.params.location || 'Все варианты'}
			</h4>

			<!-- Подзаголовок с характеристиками -->
			<p class="mt-0.5 flex items-center gap-1.5 truncate text-xs font-medium text-zinc-500 sm:text-sm">
				<span>{getPropertyTypeLabel(search.params.propertyType)}</span>
				<span class="h-1 w-1 rounded-full bg-zinc-300"></span>
				<span>{formatGuests(search.params.adults, search.params.children)}</span>
			</p>
		</div>

		<!-- Action Button (Arrow) -->
		<div class="relative flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-zinc-100 text-zinc-400 transition-all duration-300 ease-out group-hover:bg-zinc-900 group-hover:text-white group-hover:shadow-md">
			<ArrowRight size={18} strokeWidth={2.5} class="transition-transform duration-300 ease-out group-hover:translate-x-0.5" />
		</div>
	</button>
{/if}
