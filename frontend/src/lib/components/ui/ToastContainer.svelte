<!-- src/lib/components/ui/ToastContainer.svelte -->
<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	import { cubicOut, backOut } from 'svelte/easing';
	import { toast, type Toast, type ConfirmToast } from '$lib/stores/toastStore';
	import { CircleCheckBig, CircleX, TriangleAlert, Info, X, CircleAlert } from 'lucide-svelte';
	import Button from './Button.svelte';
	import { onDestroy } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';

	let toasts: Toast[] = $state([]);
	let confirmToastData: ConfirmToast | null = $state(null);

	const timerData = new SvelteMap<
		string,
		{
			remainingTime: number;
			totalDuration: number;
			isPaused: boolean;
			lastUpdateTime: number;
			dismissTimeout: ReturnType<typeof setTimeout> | null;
		}
	>();

	const unsubscribeToasts = toast.subscribe((value) => {
		toasts = value;

		for (const t of value) {
			if (!timerData.has(t.id) && t.duration > 0) {
				timerData.set(t.id, {
					remainingTime: t.duration,
					totalDuration: t.duration,
					isPaused: false,
					lastUpdateTime: Date.now(),
					dismissTimeout: null
				});
				startTimer(t.id);
			}
		}

		const currentIds = new Set(value.map((t) => t.id));
		for (const id of timerData.keys()) {
			if (!currentIds.has(id)) cleanup(id);
		}
	});

	const unsubscribeConfirm = toast.confirmToast.subscribe((v) => (confirmToastData = v));

	function cleanup(id: string) {
		const d = timerData.get(id);
		if (d) {
			if (d.dismissTimeout) clearTimeout(d.dismissTimeout);
			timerData.delete(id);
		}
	}

	function startTimer(id: string) {
		const d = timerData.get(id);
		if (!d || d.isPaused) return;

		d.lastUpdateTime = Date.now();
		if (d.dismissTimeout) clearTimeout(d.dismissTimeout);
		d.dismissTimeout = setTimeout(() => toast.dismiss(id), d.remainingTime);
	}

	function pauseToast(id: string) {
		const d = timerData.get(id);
		if (!d || d.isPaused) return;
		d.isPaused = true;
		d.remainingTime = Math.max(0, d.remainingTime - (Date.now() - d.lastUpdateTime));
		if (d.dismissTimeout) {
			clearTimeout(d.dismissTimeout);
			d.dismissTimeout = null;
		}
	}

	function resumeToast(id: string) {
		const d = timerData.get(id);
		if (!d || !d.isPaused) return;
		d.isPaused = false;
		if (d.remainingTime > 0) startTimer(id);
		else toast.dismiss(id);
	}

	onDestroy(() => {
		unsubscribeToasts();
		unsubscribeConfirm();
		for (const id of timerData.keys()) cleanup(id);
	});

	const typeConfig = {
		success: {
			icon: CircleCheckBig,
			bgGradient: 'from-emerald-500 to-green-500',
			borderColor: 'border-emerald-200',
			bgLight: 'bg-emerald-50',
			textColor: 'text-emerald-800',
			progressColor: 'bg-emerald-500'
		},
		error: {
			icon: CircleX,
			bgGradient: 'from-red-500 to-rose-500',
			borderColor: 'border-red-200',
			bgLight: 'bg-red-50',
			textColor: 'text-red-800',
			progressColor: 'bg-red-500'
		},
		warning: {
			icon: TriangleAlert,
			bgGradient: 'from-amber-500 to-orange-500',
			borderColor: 'border-amber-200',
			bgLight: 'bg-amber-50',
			textColor: 'text-amber-800',
			progressColor: 'bg-amber-500'
		},
		info: {
			icon: Info,
			bgGradient: 'from-blue-500 to-indigo-500',
			borderColor: 'border-blue-200',
			bgLight: 'bg-blue-50',
			textColor: 'text-blue-800',
			progressColor: 'bg-blue-500'
		}
	};

	const confirmTypeConfig = {
		danger: {
			icon: CircleAlert,
			iconBg: 'bg-gradient-to-br from-red-500 to-rose-600',
			confirmTone: 'danger' as const
		},
		warning: {
			icon: TriangleAlert,
			iconBg: 'bg-gradient-to-br from-amber-500 to-orange-500',
			confirmTone: 'primary' as const
		},
		info: {
			icon: Info,
			iconBg: 'bg-gradient-to-br from-blue-500 to-indigo-500',
			confirmTone: 'primary' as const
		}
	};
</script>

<div
	class="pointer-events-none fixed top-4 left-1/2 -translate-x-1/2 z-[300]
           flex flex-col items-center gap-2.5 w-[calc(100%-2rem)] max-w-md"
	aria-live="polite"
	aria-label="Уведомления"
>
	{#each toasts as t (t.id)}
		{@const config = typeConfig[t.type]}
		{@const ToastIcon = config.icon}

		<div
			class="pointer-events-auto relative w-full overflow-hidden rounded-xl border sm:rounded-2xl
                   {config.borderColor} {config.bgLight} shadow-xl backdrop-blur-sm"
			in:fly={{ y: -20, duration: 250, easing: backOut }}
			out:fly={{ y: -20, duration: 180, easing: cubicOut }}
			animate:flip={{ duration: 250 }}
			role="alert"
			onmouseenter={() => pauseToast(t.id)}
			onmouseleave={() => resumeToast(t.id)}
			ontouchstart={() => pauseToast(t.id)}
			ontouchend={() => resumeToast(t.id)}
		>
			<div class="p-3 sm:p-4">
				<div class="flex items-start gap-2.5 sm:gap-3">
					<div
						class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br
                                sm:h-9 sm:w-9 sm:rounded-xl {config.bgGradient} shadow-lg"
					>
						<ToastIcon class="h-4 w-4 text-white sm:h-5 sm:w-5" />
					</div>
					<div class="min-w-0 flex-1 pt-0.5">
						<p class="text-xs font-semibold sm:text-sm {config.textColor} line-clamp-2">
							{t.title}
						</p>
						{#if t.message}
							<p
								class="mt-0.5 line-clamp-2 text-[11px] leading-relaxed text-zinc-600 sm:mt-1 sm:text-xs"
							>
								{t.message}
							</p>
						{/if}
					</div>
					{#if t.dismissible}
						<button
							type="button"
							class="flex h-6 w-6 shrink-0 touch-manipulation items-center justify-center
                                   rounded-md text-zinc-400 transition-all duration-200
                                   hover:bg-white/50 hover:text-zinc-600 active:bg-white/70
                                   sm:h-7 sm:w-7 sm:rounded-lg"
							onclick={() => toast.dismiss(t.id)}
							aria-label="Закрыть"
						>
							<X class="h-3.5 w-3.5 sm:h-4 sm:w-4" />
						</button>
					{/if}
				</div>
			</div>
			{#if t.duration > 0}
				<div class="absolute right-0 bottom-0 left-0 h-0.5 bg-black/5 sm:h-1">
					<div
						class="h-full {config.progressColor} toast-progress-bar"
						style="--toast-duration: {t.duration}ms; animation-play-state: {timerData.get(t.id)?.isPaused ? 'paused' : 'running'};"
					></div>
				</div>
			{/if}
		</div>
	{/each}
</div>

{#if confirmToastData}
	{@const config = confirmTypeConfig[confirmToastData.type]}
	{@const ConfirmIcon = config.icon}

	<button
		type="button"
		class="fixed inset-0 z-[300] cursor-default border-none bg-black/50 p-0 backdrop-blur-sm"
		in:fade={{ duration: 200 }}
		out:fade={{ duration: 150 }}
		onclick={() => confirmToastData?.onCancel()}
		onkeydown={(e) => e.key === 'Escape' && confirmToastData?.onCancel()}
		aria-label="Закрыть диалог"
	></button>

	<div
		class="pointer-events-none fixed inset-0 z-[301] flex items-end justify-center p-0 sm:items-center sm:p-4"
	>
		<section
			class="pointer-events-auto w-full rounded-t-3xl bg-white p-5 shadow-2xl
                   sm:max-w-md sm:rounded-3xl sm:p-6"
			in:fly={{ y: 100, duration: 300, easing: backOut }}
			out:fly={{ y: 100, duration: 200, easing: cubicOut }}
			role="alertdialog"
			tabindex="-1"
			aria-modal="true"
		>
			<div class="mb-4 flex justify-center sm:hidden">
				<div class="h-1 w-10 rounded-full bg-zinc-300"></div>
			</div>
			<div class="mb-4 flex justify-center sm:mb-5">
				<div
					class="flex h-12 w-12 items-center justify-center rounded-xl
                            sm:h-14 sm:w-14 sm:rounded-2xl {config.iconBg} shadow-lg"
				>
					<ConfirmIcon class="h-6 w-6 text-white sm:h-7 sm:w-7" />
				</div>
			</div>
			<div class="mb-5 text-center sm:mb-6">
				<h3 class="mb-1.5 text-base font-bold text-zinc-900 sm:mb-2 sm:text-lg">
					{confirmToastData.title}
				</h3>
				<p class="text-xs leading-relaxed text-zinc-500 sm:text-sm">
					{confirmToastData.message}
				</p>
			</div>
			<div class="flex flex-col-reverse gap-2 sm:flex-row sm:gap-3">
				<Button
					variant="outline"
					tone="neutral"
					size="lg"
					radius="xl"
					fullWidth
					onclick={() => confirmToastData?.onCancel()}
				>
					{confirmToastData.cancelText}
				</Button>
				<Button
					variant="solid"
					tone={config.confirmTone}
					size="lg"
					radius="xl"
					fullWidth
					onclick={() => confirmToastData?.onConfirm()}
				>
					{confirmToastData.confirmText}
				</Button>
			</div>
			<div class="h-[env(safe-area-inset-bottom)] sm:hidden"></div>
		</section>
	</div>
{/if}

<style>
	@keyframes toastProgress {
		from {
			width: 100%;
		}
		to {
			width: 0%;
		}
	}
	.toast-progress-bar {
		animation: toastProgress var(--toast-duration, 4000ms) linear forwards;
	}
</style>
