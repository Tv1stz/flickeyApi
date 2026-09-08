<!-- src/lib/components/navigation/ModeSwitcher.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { isHostUser } from '$lib/auth/permissions';
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import { toast } from '$lib/stores/toastStore';
	import { ArrowLeftRight } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';

	let currentPath = $derived($page.url.pathname);
	let isProfilePage = $derived(currentPath === '/profile');
	let canSwitch = $derived(authStore.canSwitchModes);
	let hostMode = $derived(authStore.isHostMode);
	let user = $derived(authStore.user);

	function switchMode() {
		if (!isHostUser(user) || !canSwitch) {
			toast.info('Режим хозяина недоступен', 'Сначала начните размещать жилье');
			return;
		}
		const newMode = authStore.viewMode === 'guest' ? 'host' : 'guest';
		authStore.setViewMode(newMode);
		if (newMode === 'host') {
			toast.success('Режим хозяина', 'Управляйте объявлениями и статистикой');
			goto(resolve('/host/listings'));
		} else {
			toast.success('Режим гостя', 'Ищите и бронируйте жильё');
			goto(resolve('/'));
		}
	}
</script>

{#if canSwitch && isProfilePage}
	<div class="pointer-events-none fixed right-0 bottom-0 left-0 z-[41] lg:hidden">
		<div class="flex items-end justify-center px-4 pb-[calc(var(--tab-bar-height,68px)+12px)]">
			<Button
				variant="solid"
				tone="primary"
				size="lg"
				radius="pill"
				onclick={switchMode}
				class="pointer-events-auto shadow-lg shadow-black/10"
			>
				<ArrowLeftRight size={16} strokeWidth={1.75} class="mr-1.5" />
				{hostMode ? 'Режим гостя' : 'Режим хозяина'}
			</Button>
		</div>
	</div>
{/if}
