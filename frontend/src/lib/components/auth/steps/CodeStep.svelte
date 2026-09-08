<!-- src/lib/components/auth/steps/CodeStep.svelte -->
<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import OTPInput from '$lib/components/form/OTPInput.svelte';

	interface Props {
		loading: boolean;
		resendTimer: number;
		onCodeChange: (code: string) => void;
		onCodeComplete: (code: string) => void;
		onResend: () => void;
		onSubmit: () => void;
		code: string;
	}

	let { loading, resendTimer, onCodeChange, onCodeComplete, onResend, onSubmit, code }: Props =
		$props();

	let otpComponent = $state<{ focus: () => void; clear: () => void } | undefined>();

	const canSubmit = $derived(code.length === 6 && !loading);

	export function focus() {
		otpComponent?.focus();
	}

	export function clear() {
		otpComponent?.clear();
	}
</script>

<div class="space-y-8">
	<OTPInput
		bind:this={otpComponent}
		id="auth-otp"
		length={6}
		disabled={loading}
		onChange={onCodeChange}
		onComplete={onCodeComplete}
	/>

	<div class="text-center">
		{#if resendTimer > 0}
			<p class="text-sm text-zinc-500">
				Отправить повторно через
				<span class="font-semibold text-zinc-900 tabular-nums">
					{resendTimer} сек
				</span>
			</p>
		{:else}
			<button
				type="button"
				onclick={onResend}
				disabled={loading}
				class="text-sm font-medium text-zinc-900 underline decoration-zinc-300
                       underline-offset-4 transition-colors
                       duration-200 hover:decoration-zinc-900
                       disabled:text-zinc-400 disabled:no-underline"
			>
				Отправить код повторно
			</button>
		{/if}
	</div>

	<Button
		variant="solid"
		tone="primary"
		size="xl"
		radius="2xl"
		fullWidth
		disabled={!canSubmit}
		{loading}
		loadingText="Проверка..."
		onclick={onSubmit}
	>
		Подтвердить
	</Button>
</div>
