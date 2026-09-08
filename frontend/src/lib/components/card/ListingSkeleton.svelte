<div class="animate-pulse" aria-hidden="true">
	<!-- Фото -->
	<div class="overflow-hidden rounded-2xl">
		<div class="relative aspect-[12/13] w-full overflow-hidden bg-zinc-200">
			<div
				class="animate-shimmer absolute inset-0 -translate-x-full
                        bg-gradient-to-r from-transparent via-white/30 to-transparent"
			></div>
		</div>
	</div>

	<!-- Контент -->
	<div class="space-y-2 pt-3 sm:pt-3.5">
		<div class="h-5 w-3/4 rounded-md bg-zinc-200 sm:h-6"></div>
		<div class="h-4 w-1/2 rounded-md bg-zinc-100 sm:h-5"></div>
		<div class="h-4 w-2/5 rounded-md bg-zinc-100"></div>
		<div class="!mt-2.5 h-5 w-1/3 rounded-md bg-zinc-200 sm:h-6"></div>
	</div>
</div>

<style>
	@keyframes shimmer {
		100% {
			transform: translateX(100%);
		}
	}
	.animate-shimmer {
		animation: shimmer 1.5s infinite;
	}
</style>
