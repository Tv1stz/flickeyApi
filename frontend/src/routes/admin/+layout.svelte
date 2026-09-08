<script lang="ts">
	import { auth } from '$lib/auth/auth.svelte';
	import AdminNav from '$lib/components/admin/AdminNav.svelte';
	import { ShieldAlert, Loader2, LogIn, ArrowLeft } from 'lucide-svelte';

	let { children } = $props();
</script>

{#if !auth.isInitialized || auth.isLoading}
	<div class="min-h-screen bg-zinc-50/50 flex flex-col items-center justify-center p-4 dark:bg-background">
		<Loader2 class="h-8 w-8 text-zinc-900 animate-spin mb-3 dark:text-white" />
		<p class="text-sm font-medium text-zinc-500 dark:text-muted-foreground">Проверка прав администратора...</p>
	</div>
{:else if !auth.user || auth.user.role !== 'admin'}
	<div class="min-h-screen bg-zinc-50/50 flex flex-col items-center justify-center p-4 dark:bg-background">
		<div class="p-8 rounded-3xl bg-white border border-zinc-200/80 text-center max-w-md space-y-4 shadow-sm dark:bg-card dark:border-border">
			<div class="w-14 h-14 rounded-2xl bg-rose-50 text-rose-600 flex items-center justify-center mx-auto dark:bg-rose-950 dark:text-rose-400">
				<ShieldAlert class="h-7 w-7" />
			</div>
			<div>
				<h2 class="text-lg font-bold text-zinc-900 dark:text-foreground">Доступ ограничен</h2>
				<p class="text-xs text-zinc-500 mt-1.5 dark:text-muted-foreground leading-relaxed">
					{#if !auth.user}
						Для работы в панели управления необходимо войти под учетной записью администратора (+375290000000).
					{:else}
						Ваш текущий аккаунт ({auth.user.phone}) имеет роль <b>{auth.user.role}</b>. Доступ разрешен только для администраторов (<b>admin</b>).
					{/if}
				</p>
			</div>

			<div class="flex items-center justify-center gap-3 pt-2">
				{#if !auth.user}
					<button
						type="button"
						onclick={() => auth.openAuthModal('phone')}
						class="px-5 py-2.5 rounded-full bg-zinc-900 text-white text-xs font-semibold hover:bg-zinc-800 transition-all flex items-center gap-1.5 cursor-pointer shadow-xs active:scale-95 dark:bg-white dark:text-zinc-900 dark:hover:bg-zinc-100"
					>
						<LogIn class="h-4 w-4" />
						<span>Войти</span>
					</button>
				{/if}
				<a
					href="/"
					class="px-5 py-2.5 rounded-full border border-zinc-200/80 bg-white hover:bg-zinc-50 text-xs font-semibold text-zinc-700 transition-all shadow-2xs active:scale-95 dark:border-border dark:bg-card dark:text-muted-foreground dark:hover:text-foreground"
				>
					На главную
				</a>
			</div>
		</div>
	</div>
{:else}
	<div class="min-h-screen bg-zinc-50/50 flex flex-col font-sans dark:bg-background">
		<AdminNav />
		<main class="flex-1 max-w-[1440px] w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8 space-y-8">
			{@render children()}
		</main>
	</div>
{/if}