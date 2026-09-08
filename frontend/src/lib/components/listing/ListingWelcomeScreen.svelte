<script lang="ts">
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import { ArrowRight, Check } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';

	interface Props {
		onStart: () => void;
		onExit: () => void;
		hasDraft?: boolean;
		resumeStep?: number;
	}

	let { onStart, onExit, hasDraft = false, resumeStep = 1 }: Props = $props();


	const ctaText = $derived(
		hasDraft && resumeStep > 1 ? `Продолжить с шага ${resumeStep}` : 'Начать создание'
	);

	const points = [
		'Локация с картой и поиском адреса',
		'Удобная загрузка и сортировка фото',
		'Понятные шаги с обязательными полями'
	] as const;
</script>

<section class="min-h-screen bg-white" in:fade={{ duration: 220 }}>
	<div
		class="mx-auto flex min-h-screen w-full max-w-[1220px] items-center px-6 py-10 sm:px-10 lg:px-14"
	>
		<div class="grid w-full gap-12 lg:grid-cols-[1.05fr_0.95fr] lg:items-center">
			<div class="space-y-9">
				<div class="space-y-4">
					<h1 class="max-w-[18ch] text-4xl font-bold text-zinc-900 sm:text-5xl">
                        Получите первых
                        гостей уже сегодня
					</h1>
					<p class="max-w-[56ch] text-base leading-relaxed text-zinc-600 sm:text-lg">
                        Добавьте свой объект размещения
                        на Flickey
					</p>
				</div>

				<div class="space-y-3">
					{#each points as point (point)}
						<div class="flex items-start gap-3 rounded-2xl border border-zinc-200 px-4 py-3">
							<div
								class="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-zinc-900 text-white"
							>
								<Check class="h-3 w-3" strokeWidth={3} />
							</div>
							<p class="text-sm text-zinc-700 sm:text-base">{point}</p>
						</div>
					{/each}
				</div>

				<div class="flex flex-wrap items-center gap-3">
					<Button onclick={onStart} iconRight={ArrowRight}>
						{ctaText}
					</Button>

					<Button variant="ghost" tone="neutral" onclick={onExit}>Не сейчас</Button>
				</div>
			</div>

			<div class="space-y-4">
					<div class="overflow-visible">
						<img
							src="/ill.png"
							alt="Иллюстрация приветственного экрана"
							class="w-auto"
						/>
					</div>
			</div>
		</div>
	</div>
</section>
