<!-- src/lib/components/ui/ToggleSwitch.svelte -->
<script lang="ts">
	interface Props {
		checked?: boolean;
		disabled?: boolean;
		size?: 'sm' | 'md' | 'lg';
		label?: string;
		onToggle?: (value: boolean) => void;
	}

	let { checked = false, disabled = false, size = 'md', label, onToggle }: Props = $props();

	const config = {
		sm: { track: 'h-[22px] w-[40px]', thumb: 'h-[16px] w-[16px]', on: 'translate-x-[18px]' },
		md: { track: 'h-[28px] w-[50px]', thumb: 'h-[22px] w-[22px]', on: 'translate-x-[22px]' },
		lg: { track: 'h-[32px] w-[56px]', thumb: 'h-[26px] w-[26px]', on: 'translate-x-[24px]' }
	};

	const current = $derived(config[size]);

	function handleClick(e: MouseEvent) {
		e.stopPropagation();
		if (!disabled) onToggle?.(!checked);
	}
</script>

<button
	type="button"
	role="switch"
	aria-checked={checked}
	aria-label={label}
	{disabled}
	onclick={handleClick}
	class="group relative shrink-0 rounded-full transition-colors duration-200 ease-in-out
		focus:outline-none focus-visible:ring-2 focus-visible:ring-zinc-900 focus-visible:ring-offset-2
		{current.track}
		{disabled
		? 'cursor-not-allowed opacity-40'
		: checked
			? 'bg-zinc-900 hover:bg-zinc-700 active:bg-zinc-950'
			: 'bg-zinc-300 hover:bg-zinc-400 active:bg-zinc-500'}"
>
	<span
		class="absolute top-1/2 left-[3px] -translate-y-1/2 rounded-full bg-white shadow-md
			transition-all duration-200 ease-in-out
			group-active:scale-90
			{current.thumb}
			{checked ? current.on : 'translate-x-0'}"
	></span>
</button>
