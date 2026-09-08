<script lang="ts">
	import '../app.css';
	import { onMount, type Snippet } from 'svelte';
	import { page } from '$app/stores';
	import { auth } from '$lib/auth/auth.svelte';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { favoritesStore } from '$lib/stores/favoritesStore.svelte';
	import { searchStore } from '$lib/stores/searchStore.svelte';
	import { throttle } from '$lib/utils/dom';
	import DesktopNav from '$lib/components/navigation/DesktopNav.svelte';
	import MobileTabBar from '$lib/components/navigation/MobileTabBar.svelte';
	import ModeSwitcher from '$lib/components/navigation/ModeSwitcher.svelte';
	import ToastContainer from '$lib/components/ui/ToastContainer.svelte';
	import AuthDialog from '$lib/components/auth/AuthDialog.svelte';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();

	let scrolled = $state(false);
	let currentPath = $derived($page.url.pathname);
	let desktopNavRef = $state<ReturnType<typeof DesktopNav>>();

	const isAdminPage = $derived(currentPath.startsWith('/admin'));
	const isAuthPage = $derived(currentPath.startsWith('/auth'));
	const isListingDetailPage = $derived(
		/^\/listings\/[^/]+$/.test(currentPath) && !currentPath.endsWith('/new')
	);
	const isListingCreationPage = $derived(
		currentPath === '/listings/new' || currentPath === '/host/new'
	);
	const isHomePage = $derived(currentPath === '/');
	const isSearchPage = $derived(currentPath.startsWith('/search'));

	const showHeaderSearch = $derived(isHomePage || isSearchPage);
	const headerSearchOffset = $derived(showHeaderSearch ? (isSearchPage ? '0px' : '88px') : '0px');

	const handleScroll = throttle(() => {
		if (typeof window === 'undefined') return;
		const currentScrollY = window.scrollY;
		const isScrolled = currentScrollY > 20;
		if (scrolled !== isScrolled) {
			scrolled = isScrolled;
		}
	}, 50);

	onMount(() => {
		auth.init();
		authStore.initialize();
		favoritesStore.initialize();

		const measureNav = () => {
			const nav = document.querySelector('[data-tab-bar]');
			if (nav) {
				const height = Math.ceil(nav.getBoundingClientRect().height);
				document.documentElement.style.setProperty('--tab-bar-height', `${height}px`);
				document.documentElement.style.setProperty('--mobile-nav-height', `${height}px`);
			}
		};
		measureNav();
		window.addEventListener('resize', measureNav);
		return () => {
			window.removeEventListener('resize', measureNav);
		};
	});
</script>

<svelte:window onscroll={handleScroll} />

<svelte:head>
	<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover" />
	<meta name="theme-color" content="#ffffff" />
</svelte:head>

{#if isAdminPage || isAuthPage}
	{@render children()}
{:else}
	<DesktopNav {scrolled} bind:this={desktopNavRef} />
	<ModeSwitcher />
	<MobileTabBar />

	<main
		style={`--header-search-offset: ${headerSearchOffset};`}
		class={isListingDetailPage || isListingCreationPage
			? ''
			: 'pb-[calc(var(--tab-bar-height,68px)+8px)] lg:pt-[calc(96px+var(--header-search-offset))] lg:pb-0'}
	>
		{@render children()}
	</main>
{/if}

<AuthDialog />
<ToastContainer />

