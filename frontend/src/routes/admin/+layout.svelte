<script lang="ts">
	import { auth } from '$lib/auth/auth.svelte';
	import AdminNav from '$lib/components/admin/AdminNav.svelte';
	import { ShieldAlert, Loader2, LogIn } from 'lucide-svelte';

	let { children } = $props();
</script>

{#if !auth.isInitialized || auth.isLoading}
	<div class="min-h-screen bg-background flex flex-col items-center justify-center p-4">
		<Loader2 class="h-8 w-8 text-primary animate-spin mb-3" />
		<p class="text-sm font-medium text-muted-foreground">Проверка прав администратора...</p>
	</div>
{:else if !auth.user || auth.user.role !== 'admin'}
	<div class="min-h-screen bg-background flex flex-col items-center justify-center p-4">
		<div class="p-6 rounded-3xl bg-card border border-border text-center max-w-md space-y-4 shadow-sm">
			<div class="w-12 h-12 rounded-2xl bg-rose-100 text-rose-600 dark:bg-rose-950 dark:text-rose-400 flex items-center justify-center mx-auto">
				<ShieldAlert class="h-6 w-6" />
			</div>
			<div>
				<h2 class="text-lg font-bold text-foreground">Доступ ограничен</h2>
				<p class="text-xs text-muted-foreground mt-1">
					{#if !auth.user}
						Вы не авторизованы. Для доступа к панели управления войдите под аккаунтом администратора (+375290000000).
					{:else}
						Ваш текущий аккаунт ({auth.user.phone}) имеет роль <b>{auth.user.role}</b>. Для доступа требуются права <b>admin</b>.
					{/if}
				</p>
			</div>

			<div class="flex items-center justify-center gap-3 pt-2">
				{#if !auth.user}
					<button
						type="button"
						onclick={() => auth.openAuthModal('phone')}
						class="px-4 py-2 rounded-xl bg-primary text-primary-foreground text-xs font-semibold hover:bg-primary/90 transition-colors flex items-center gap-1.5 cursor-pointer shadow-xs"
					>
						<LogIn class="h-4 w-4" />
						<span>Войти</span>
					</button>
				{/if}
				<a
					href="/"
					class="px-4 py-2 rounded-xl border border-border hover:bg-muted text-xs font-semibold text-foreground transition-colors"
				>
					На главную
				</a>
			</div>
		</div>
	</div>
{:else}
	<div class="min-h-screen bg-background flex flex-col font-sans">
		<AdminNav />
		<main class="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
			{@render children()}
		</main>
	</div>
{/if}
