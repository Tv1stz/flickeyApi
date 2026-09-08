<script lang="ts">
	interface Props {
		name: string;
		initials: string;
		modeLabel: string;
		avatar?: string;
		onAvatarClick: () => void;
		variant?: 'desktop' | 'mobile';
	}

	let {
		name,
		initials,
		modeLabel,
		avatar,
		onAvatarClick,
		variant = 'desktop'
	}: Props = $props();

	const isMobile = $derived(variant === 'mobile');
</script>

<div class={isMobile ? 'flex items-center gap-4' : 'flex flex-col items-center text-center'}>
	<button
		type="button"
		class={`relative flex shrink-0 items-center justify-center overflow-hidden rounded-full bg-zinc-900 text-white transition-all duration-200 hover:opacity-90 active:scale-95 ${
			isMobile ? 'h-16 w-16 text-xl font-semibold' : 'h-20 w-20 text-2xl font-semibold'
		}`}
		onclick={onAvatarClick}
		aria-label="Изменить фото профиля"
	>
		{#if avatar}
			<img src={avatar} alt="Фото профиля" class="h-full w-full object-cover" />
		{:else}
			{initials}
		{/if}
	</button>

	<div class={isMobile ? '' : 'mt-4'}>
		<p class="text-base font-semibold text-zinc-900 sm:text-lg">Привет, {name}!</p>
		<p class="mt-0.5 text-sm font-medium text-zinc-500">{modeLabel}</p>
	</div>
</div>
