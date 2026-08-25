<script lang="ts">
	import { auth } from '$lib/auth/auth.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import ContextSwitcher from './ContextSwitcher.svelte';
	import NotificationBell from '$lib/components/notifications/NotificationBell.svelte';
	import { maskPhone } from '$lib/utils';
	import {
		Compass,
		PlusCircle,
		LayoutDashboard,
		User as UserIcon,
		LogOut,
		Menu,
		X,
		ShieldCheck,
		Building2,
		Zap
	} from 'lucide-svelte';

	let isUserMenuOpen = $state(false);
	let isMobileMenuOpen = $state(false);

	function toggleUserMenu() {
		isUserMenuOpen = !isUserMenuOpen;
	}

	function closeMenus() {
		isUserMenuOpen = false;
		isMobileMenuOpen = false;
	}

	async function handleLogout() {
		closeMenus();
		await auth.logout();
	}
</script>

<header class="sticky top-0 z-40 w-full border-b border-border/80 bg-background/95 backdrop-blur-md">
	<div class="container mx-auto flex h-16 items-center justify-between px-4 sm:px-6">
		<!-- Brand Logo -->
		<div class="flex items-center gap-6">
			<a href="/" class="flex items-center gap-2.5 font-extrabold text-xl tracking-tight text-foreground" onclick={closeMenus}>
				<div class="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
					<Building2 class="h-5 w-5" />
				</div>
				<span class="bg-gradient-to-r from-primary to-emerald-700 bg-clip-text text-transparent">Flickey</span>
			</a>

			<!-- Desktop Nav Links -->
			<nav class="hidden md:flex items-center gap-1 text-sm font-medium">
				<a
					href="/"
					class="flex items-center gap-1.5 rounded-lg px-3 py-2 text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
				>
					<Compass class="h-4 w-4" />
					<span>Каталог</span>
				</a>

				{#if auth.isHost}
					<a
						href="/host"
						class="flex items-center gap-1.5 rounded-lg px-3 py-2 text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
					>
						<LayoutDashboard class="h-4 w-4" />
						<span>Мои объекты</span>
					</a>
				{/if}

				{#if auth.isAdmin}
					<a
						href="/admin"
						class="flex items-center gap-1.5 rounded-lg px-3 py-2 text-amber-700 dark:text-amber-400 hover:bg-amber-50 dark:hover:bg-amber-950/30 transition-colors"
					>
						<ShieldCheck class="h-4 w-4" />
						<span>Админ-панель</span>
					</a>
				{/if}
			</nav>
		</div>

		<!-- Center: Context Switcher (Desktop) -->
		<div class="hidden lg:flex items-center justify-center">
			<ContextSwitcher />
		</div>

		<!-- Right Action Items -->
		<div class="flex items-center gap-3">
			<!-- Create Listing CTA -->
			<a href="/host/new" class="hidden sm:inline-flex">
				<Button size="sm" class="gap-1.5 font-semibold shadow-sm">
					<PlusCircle class="h-4 w-4" />
					<span>Сдать жилье</span>
				</Button>
			</a>

			<!-- Auth Section -->
			{#if auth.isLoading}
				<div class="h-9 w-20 animate-pulse rounded-md bg-muted"></div>
			{:else if auth.isAuthenticated}
				<!-- Notifications Bell -->
				<NotificationBell />

				<!-- User Menu Dropdown -->
				<div class="relative">
					<button
						type="button"
						onclick={toggleUserMenu}
						class="flex items-center gap-2 rounded-full border border-border bg-card p-1.5 pr-3 hover:bg-accent transition-colors cursor-pointer select-none"
						aria-expanded={isUserMenuOpen}
					>
						<div class="flex h-7 w-7 items-center justify-center rounded-full bg-primary/10 text-primary font-bold text-xs">
							{auth.user?.first_name ? auth.user.first_name[0].toUpperCase() : 'U'}
						</div>
						<span class="text-xs font-semibold text-foreground max-w-[120px] truncate hidden md:inline-block">
							{auth.displayName}
						</span>
					</button>

					{#if isUserMenuOpen}
						<!-- Backdrop -->
						<!-- svelte-ignore a11y_click_events_have_key_events -->
						<!-- svelte-ignore a11y_no_static_element_interactions -->
						<div class="fixed inset-0 z-40" onclick={closeMenus}></div>

						<!-- Dropdown Panel -->
						<div
							class="absolute right-0 top-full mt-2 w-64 z-50 rounded-2xl border border-border bg-card p-2 shadow-2xl animate-in zoom-in-95"
						>
							<div class="px-3 py-2 border-b border-border/80 mb-1">
								<p class="text-xs font-bold text-foreground truncate">{auth.displayName}</p>
								<p class="text-[11px] text-muted-foreground font-mono mt-0.5">{maskPhone(auth.user?.phone || '')}</p>
								<div class="mt-1.5 flex items-center gap-1.5">
									<span class="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-bold text-primary">
										Роль: {auth.user?.role}
									</span>
									<span class="inline-flex items-center rounded-full bg-muted px-2 py-0.5 text-[10px] font-medium text-muted-foreground">
										{auth.activeContext === 'host' ? 'Режим: Хост' : 'Режим: Гость'}
									</span>
								</div>
							</div>

							<div class="py-1 text-xs space-y-0.5">
								<a
									href="/profile"
									onclick={closeMenus}
									class="flex items-center gap-2 rounded-lg px-3 py-2 text-foreground hover:bg-accent transition-colors"
								>
									<UserIcon class="h-4 w-4 text-muted-foreground" />
									<span>Профиль аккаунта</span>
								</a>

								{#if auth.isHost}
									<a
										href="/host"
										onclick={closeMenus}
										class="flex items-center gap-2 rounded-lg px-3 py-2 text-foreground hover:bg-accent transition-colors"
									>
										<LayoutDashboard class="h-4 w-4 text-muted-foreground" />
										<span>Кабинет хоста</span>
									</a>
								{/if}

								{#if auth.isAdmin}
									<a
										href="/admin"
										onclick={closeMenus}
										class="flex items-center gap-2 rounded-lg px-3 py-2 text-emerald-700 dark:text-emerald-400 font-bold hover:bg-emerald-50 dark:hover:bg-emerald-950/30 transition-colors"
									>
										<ShieldCheck class="h-4 w-4" />
										<span>Админ-панель Compliance</span>
									</a>
								{/if}

								<a
									href="/host/new"
									onclick={closeMenus}
									class="flex items-center gap-2 rounded-lg px-3 py-2 text-primary font-semibold hover:bg-primary/5 transition-colors"
								>
									<PlusCircle class="h-4 w-4" />
									<span>Создать новое объявление</span>
								</a>
							</div>

							<div class="border-t border-border/80 pt-1 mt-1">
								<button
									type="button"
									onclick={handleLogout}
									class="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-xs font-medium text-destructive hover:bg-destructive/10 transition-colors cursor-pointer"
								>
									<LogOut class="h-4 w-4" />
									<span>Выйти из аккаунта</span>
								</button>
							</div>
						</div>
					{/if}
				</div>
			{:else}
				<div class="flex items-center gap-2">
					<Button
						variant="outline"
						size="sm"
						onclick={() => auth.openAuthModal('phone')}
						class="text-xs font-semibold"
					>
						Войти / Регистрация
					</Button>
				</div>
			{/if}

			<!-- Mobile Menu Toggle Button -->
			<button
				type="button"
				onclick={() => (isMobileMenuOpen = !isMobileMenuOpen)}
				class="md:hidden rounded-lg p-2 text-muted-foreground hover:bg-accent hover:text-foreground cursor-pointer"
				aria-label="Меню навигации"
			>
				{#if isMobileMenuOpen}
					<X class="h-5 w-5" />
				{:else}
					<Menu class="h-5 w-5" />
				{/if}
			</button>
		</div>
	</div>

	<!-- Mobile Menu Drawer -->
	{#if isMobileMenuOpen}
		<div class="md:hidden border-t border-border bg-card p-4 space-y-4 animate-in slide-in-from-top-2 duration-150">
			{#if auth.isAuthenticated}
				<div class="flex justify-center pb-2">
					<ContextSwitcher />
				</div>
			{/if}

			<nav class="space-y-1 text-sm font-medium">
				<a
					href="/"
					onclick={closeMenus}
					class="flex items-center gap-2.5 rounded-lg px-3 py-2.5 hover:bg-accent transition-colors"
				>
					<Compass class="h-4 w-4 text-muted-foreground" />
					<span>Каталог объявлений</span>
				</a>

				<a
					href="/host/new"
					onclick={closeMenus}
					class="flex items-center gap-2.5 rounded-lg px-3 py-2.5 text-primary font-semibold hover:bg-primary/10 transition-colors"
				>
					<PlusCircle class="h-4 w-4" />
					<span>Сдать жилье (+ объявление)</span>
				</a>

				{#if auth.isHost}
					<a
						href="/host"
						onclick={closeMenus}
						class="flex items-center gap-2.5 rounded-lg px-3 py-2.5 hover:bg-accent transition-colors"
					>
						<LayoutDashboard class="h-4 w-4 text-muted-foreground" />
						<span>Мои объекты (Хост)</span>
					</a>
				{/if}

				{#if auth.isAuthenticated}
					<a
						href="/profile"
						onclick={closeMenus}
						class="flex items-center gap-2.5 rounded-lg px-3 py-2.5 hover:bg-accent transition-colors"
					>
						<UserIcon class="h-4 w-4 text-muted-foreground" />
						<span>Мой профиль</span>
					</a>
				{/if}

				{#if auth.isAdmin}
					<a
						href="/admin"
						onclick={closeMenus}
						class="flex items-center gap-2.5 rounded-lg px-3 py-2.5 text-amber-600 hover:bg-amber-50 transition-colors"
					>
						<ShieldCheck class="h-4 w-4" />
						<span>Админ-панель</span>
					</a>
				{/if}
			</nav>

			{#if !auth.isAuthenticated}
				<div class="pt-2 border-t border-border">
					<Button
						class="w-full justify-center"
						onclick={() => {
							closeMenus();
							auth.openAuthModal('phone');
						}}
					>
						Войти или зарегистрироваться
					</Button>
				</div>
			{/if}
		</div>
	{/if}
</header>
