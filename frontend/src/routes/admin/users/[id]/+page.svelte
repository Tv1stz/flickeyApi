<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { adminUsersApi, type UserDetailResponse } from '$lib/api/admin/users';
	import UserBlockDialog from '$lib/components/admin/UserBlockDialog.svelte';
	import AdminActionMenu from '$lib/components/admin/AdminActionMenu.svelte';
	import type { EnforcementType } from '$lib/types/admin';
	import { toast } from '$lib/stores/toastStore';
	import {
		formatDate,
		formatCurrency,
		translateHousingType,
		translateListingStatus,
		cn
	} from '$lib/utils';
	import {
		ArrowLeft,
		ShieldAlert,
		ShieldCheck,
		Building2,
		History,
		Loader2,
		CheckCircle2,
		AlertTriangle,
		Phone,
		Mail,
		Calendar,
		ExternalLink,
		Unlock,
		Copy,
		Check,
		Clock,
		User as UserIcon,
		UserX
	} from 'lucide-svelte';

	let userId = $derived(page.params.id);
	let data = $state<UserDetailResponse | null>(null);
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let copied = $state(false);

	let isDialogOpen = $state(false);
	let isSubmitting = $state(false);

	async function loadUserDetail(id: string) {
		if (!id) return;
		isLoading = true;
		error = null;
		try {
			const res = await adminUsersApi.getUserDetail(id);
			data = res;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Не удалось загрузить профиль пользователя';
		} finally {
			isLoading = false;
		}
	}

	$effect(() => {
		if (userId) {
			loadUserDetail(userId);
		}
	});

	function copyId() {
		if (!userId) return;
		navigator.clipboard.writeText(userId);
		copied = true;
		toast.success('ID пользователя скопирован');
		setTimeout(() => (copied = false), 2000);
	}

	function handleConfirmEnforcement(payload: {
		enforcement_type: EnforcementType;
		reason: string;
		note: string;
		duration_days?: number;
	}) {
		if (!data) return;
		isSubmitting = true;
		adminUsersApi
			.enforceUser(data.user.id, payload)
			.then(() => {
				isDialogOpen = false;
				toast.success(
					payload.enforcement_type === 'unblock'
						? 'Пользователь успешно разблокирован'
						: 'Санкции успешно применены'
				);
				if (userId) loadUserDetail(userId);
			})
			.catch((err) => {
				toast.error('Ошибка применения мер', err.message || 'Не удалось применить решение');
			})
			.finally(() => {
				isSubmitting = false;
			});
	}

	function handleQuickUnblock() {
		if (!data) return;
		adminUsersApi
			.unblockUser(data.user.id, 'Административная отмена ограничений')
			.then(() => {
				toast.success('Пользователь разблокирован');
				if (userId) loadUserDetail(userId);
			})
			.catch((err) => {
				toast.error('Ошибка разблокировки', err.message || 'Не удалось разблокировать');
			});
	}

	function handleRoleChange(newRole: string) {
		if (!data) return;
		adminUsersApi
			.updateUserRole(data.user.id, newRole)
			.then(() => {
				toast.success(`Роль пользователя изменена на: ${newRole}`);
				if (userId) loadUserDetail(userId);
			})
			.catch((err) => {
				toast.error('Ошибка изменения роли', err.message || 'Не удалось изменить роль');
			});
	}
</script>

<svelte:head>
	<title>
		{data?.user ? `${data.user.first_name || data.user.phone || 'Пользователь'} — Профиль` : 'Профиль пользователя'} — Flickey Admin
	</title>
</svelte:head>

<div class="space-y-6">
	<!-- Top Sticky Action Bar -->
	<div class="sticky top-20 z-30 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 rounded-3xl border border-zinc-200/80 bg-white/95 p-4 sm:px-6 shadow-sm backdrop-blur-md dark:border-border dark:bg-card/95">
		<!-- Left Breadcrumb & Title -->
		<div class="flex items-center gap-3 min-w-0">
			<a
				href="/admin/users"
				class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-zinc-200/80 bg-white text-zinc-600 transition hover:bg-zinc-50 hover:text-zinc-900 active:scale-95 dark:border-border dark:bg-card dark:text-muted-foreground"
				title="Назад в реестр"
			>
				<ArrowLeft class="h-4 w-4" />
			</a>

			<div class="min-w-0">
				<div class="flex items-center gap-2">
					<span class="text-xs font-semibold text-zinc-500 dark:text-muted-foreground">Пользователь</span>
					{#if data?.user}
						<span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold {data.user.role === 'admin' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' : data.user.role === 'host' ? 'bg-zinc-900 text-white dark:bg-white dark:text-zinc-900' : 'bg-zinc-100 text-zinc-800 dark:bg-muted dark:text-foreground'}">
							{data.user.role === 'admin' ? 'Администратор' : data.user.role === 'host' ? 'Хозяин' : 'Гость'}
						</span>
					{/if}
				</div>
				<h1 class="text-sm sm:text-base font-bold text-zinc-900 truncate dark:text-foreground">
					{data ? [data.user.first_name, data.user.last_name].filter(Boolean).join(' ') || data.user.phone || 'Профиль' : 'Загрузка...'}
				</h1>
			</div>
		</div>

		<!-- Right Actions Hierarchy -->
		{#if data}
			<div class="flex items-center gap-2 shrink-0">
				{#if data.user.status === 'suspended' || data.user.status === 'banned'}
					<button
						type="button"
						onclick={handleQuickUnblock}
						class="inline-flex items-center gap-1.5 rounded-full bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs transition hover:bg-emerald-700 active:scale-95 cursor-pointer"
					>
						<Unlock class="h-3.5 w-3.5" />
						<span>Разблокировать</span>
					</button>
				{/if}

				<button
					type="button"
					onclick={() => (isDialogOpen = true)}
					class="inline-flex items-center gap-1.5 rounded-full border border-amber-300 bg-amber-50 px-3.5 py-2 text-xs font-semibold text-amber-800 transition hover:bg-amber-100 active:scale-95 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-300 cursor-pointer"
				>
					<ShieldAlert class="h-3.5 w-3.5" />
					<span>Санкции</span>
				</button>

				<AdminActionMenu
					items={[
						{
							label: data.user.role === 'host' ? 'Сделать Гостем' : 'Сделать Хозяином',
							icon: UserIcon,
							onclick: () => handleRoleChange(data?.user.role === 'host' ? 'guest' : 'host')
						},
						{
							label: data.user.role === 'admin' ? 'Снять права администратора' : 'Назначить администратором',
							onclick: () => handleRoleChange(data?.user.role === 'admin' ? 'host' : 'admin')
						},
						{
							label: 'Публичный профиль',
							icon: ExternalLink,
							divider: true,
							onclick: () => window.open(`/users/${data?.user.id}`, '_blank')
						},
						{
							label: 'Копировать ID',
							icon: Copy,
							divider: true,
							onclick: copyId
						}
					]}
				/>
			</div>
		{/if}
	</div>

	{#if isLoading}
		<div class="rounded-3xl border border-zinc-200/80 bg-white p-16 text-center shadow-2xs dark:border-border dark:bg-card">
			<Loader2 class="h-8 w-8 text-zinc-900 animate-spin mx-auto mb-3 dark:text-white" />
			<p class="text-xs font-medium text-zinc-500 dark:text-muted-foreground">Загрузка данных профиля...</p>
		</div>
	{:else if error}
		<div class="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-800 dark:border-rose-900 dark:bg-rose-950/60 dark:text-rose-200 flex items-center justify-between">
			<span>{error}</span>
			<button
				type="button"
				onclick={() => { if (userId) loadUserDetail(userId); }}
				class="font-semibold underline hover:no-underline cursor-pointer"
			>
				Повторить
			</button>
		</div>
	{:else if data}
		{@const first = data.user.first_name || ''}
		{@const last = data.user.last_name || ''}
		{@const fullName = [first, last].filter(Boolean).join(' ') || null}
		{@const initial = (fullName ? fullName[0] : (data.user.phone || '?').replace('+', '')[0] || '?').toUpperCase()}

		<!-- Main User Profile Summary Card -->
		<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 sm:p-8 shadow-2xs dark:border-border dark:bg-card space-y-6">
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-zinc-100 dark:border-border">
				<div class="flex items-center gap-4">
					<div class="flex h-16 w-16 shrink-0 items-center justify-center rounded-3xl font-extrabold text-xl bg-zinc-900 text-white dark:bg-white dark:text-zinc-900 shadow-md">
						{initial}
					</div>
					<div>
						<h2 class="text-xl font-bold text-zinc-900 dark:text-foreground">
							{fullName || 'Имя не указано'}
						</h2>
						<div class="flex items-center gap-2 text-xs text-zinc-500 font-mono mt-0.5 dark:text-muted-foreground">
							<span>ID: {data.user.id}</span>
							<button
								type="button"
								onclick={copyId}
								class="hover:text-zinc-900 transition dark:hover:text-foreground cursor-pointer"
								title="Копировать ID"
							>
								{#if copied}
									<Check class="h-3.5 w-3.5 text-emerald-600" />
								{:else}
									<Copy class="h-3.5 w-3.5" />
								{/if}
							</button>
						</div>
					</div>
				</div>

				<!-- Status Pill -->
				<div>
					{#if data.user.status === 'active'}
						<span class="inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-bold bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-950 dark:border-emerald-900 dark:text-emerald-300">
							<ShieldCheck class="h-3.5 w-3.5" /> Активен
						</span>
					{:else if data.user.status === 'suspended'}
						<span class="inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-bold bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-950 dark:border-amber-900 dark:text-amber-300">
							<AlertTriangle class="h-3.5 w-3.5" /> Ограничен
						</span>
					{:else if data.user.status === 'banned'}
						<span class="inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-bold bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-950 dark:border-rose-900 dark:text-rose-300">
							<UserX class="h-3.5 w-3.5" /> Заблокирован
						</span>
					{:else}
						<span class="inline-flex items-center rounded-full px-3 py-1 text-xs font-bold bg-zinc-100 text-zinc-700 border border-zinc-200 dark:bg-muted dark:text-muted-foreground">
							{data.user.status}
						</span>
					{/if}
				</div>
			</div>

			<!-- Meta Grid -->
			<div class="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
				<div class="rounded-2xl border border-zinc-100 bg-zinc-50/80 p-4 dark:border-border dark:bg-muted/30 space-y-1">
					<div class="text-zinc-500 font-semibold flex items-center gap-1.5 dark:text-muted-foreground">
						<Phone class="h-3.5 w-3.5 text-zinc-400" /> Телефон
					</div>
					<div class="font-mono font-bold text-zinc-900 dark:text-foreground">
						{data.user.phone || 'Не указан'}
					</div>
				</div>

				<div class="rounded-2xl border border-zinc-100 bg-zinc-50/80 p-4 dark:border-border dark:bg-muted/30 space-y-1">
					<div class="text-zinc-500 font-semibold flex items-center gap-1.5 dark:text-muted-foreground">
						<Mail class="h-3.5 w-3.5 text-zinc-400" /> Email
					</div>
					<div class="font-medium text-zinc-900 truncate dark:text-foreground">
						{data.user.email || 'Не привязан'}
					</div>
				</div>

				<div class="rounded-2xl border border-zinc-100 bg-zinc-50/80 p-4 dark:border-border dark:bg-muted/30 space-y-1">
					<div class="text-zinc-500 font-semibold flex items-center gap-1.5 dark:text-muted-foreground">
						<Calendar class="h-3.5 w-3.5 text-zinc-400" /> Регистрация
					</div>
					<div class="font-medium text-zinc-900 dark:text-foreground">
						{data.user.created_at ? formatDate(data.user.created_at) : '—'}
					</div>
				</div>
			</div>
		</div>

		<!-- User Owned Listings Section -->
		<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 sm:p-8 shadow-2xs dark:border-border dark:bg-card space-y-4">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2.5">
					<Building2 class="h-5 w-5 text-zinc-700 dark:text-foreground" />
					<h3 class="text-base font-bold text-zinc-900 dark:text-foreground">
						Объявления пользователя
					</h3>
					<span class="flex h-5 min-w-[20px] items-center justify-center rounded-full bg-zinc-900 px-1.5 text-[11px] font-bold text-white dark:bg-white dark:text-zinc-900">
						{data.listings.length}
					</span>
				</div>
			</div>

			{#if data.listings.length === 0}
				<div class="rounded-2xl border border-dashed border-zinc-200 p-8 text-center text-xs text-zinc-500 dark:border-border dark:text-muted-foreground">
					У пользователя пока нет созданных объявлений.
				</div>
			{:else}
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
					{#each data.listings as listing}
						{@const statusInfo = translateListingStatus(listing.status || 'draft')}
						<div class="rounded-2xl border border-zinc-200/80 bg-white p-4 space-y-3 shadow-2xs hover:border-zinc-300 transition-all dark:border-border dark:bg-card">
							{#if listing.media && listing.media.length > 0}
								<img
									src={listing.media[0]}
									alt={listing.name}
									class="w-full h-36 rounded-xl object-cover border border-zinc-100 dark:border-border"
								/>
							{:else}
								<div class="w-full h-36 rounded-xl bg-zinc-100 flex items-center justify-center text-xs text-zinc-400 font-medium dark:bg-muted">
									Без фото
								</div>
							{/if}

							<div class="space-y-1">
								<div class="flex items-center justify-between gap-2">
									<span class="text-[10px] font-bold text-zinc-400 uppercase tracking-wider">
										{translateHousingType(listing.type)}
									</span>
									<span class={cn('px-2 py-0.5 rounded-full text-[10px] font-bold border', statusInfo.color)}>
										{statusInfo.label}
									</span>
								</div>
								<h4 class="font-bold text-xs text-zinc-900 truncate dark:text-foreground">{listing.name}</h4>
								<div class="text-xs font-bold text-zinc-900 dark:text-foreground">
									{formatCurrency(listing.price_per_night, listing.currency)} <span class="text-[10px] font-normal text-zinc-500 dark:text-muted-foreground">/ сутки</span>
								</div>
							</div>

							<div class="pt-2 border-t border-zinc-100 flex items-center justify-between gap-2 dark:border-border">
								<a
									href="/admin/listings/{listing.id}"
									class="w-full py-1.5 rounded-xl border border-zinc-200/80 hover:bg-zinc-50 text-xs font-semibold text-zinc-700 text-center inline-flex items-center justify-center gap-1.5 transition-colors dark:border-border dark:text-muted-foreground dark:hover:text-foreground"
								>
									<span>Инспекция</span>
									<ExternalLink class="h-3 w-3" />
								</a>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Restriction & Enforcement History Timeline -->
		<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 sm:p-8 shadow-2xs dark:border-border dark:bg-card space-y-4">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2.5">
					<History class="h-5 w-5 text-zinc-700 dark:text-foreground" />
					<h3 class="text-base font-bold text-zinc-900 dark:text-foreground">
						История ограничений и санкций
					</h3>
					<span class="flex h-5 min-w-[20px] items-center justify-center rounded-full bg-zinc-900 px-1.5 text-[11px] font-bold text-white dark:bg-white dark:text-zinc-900">
						{data.restrictions.length}
					</span>
				</div>
			</div>

			{#if data.restrictions.length === 0}
				<div class="rounded-2xl border border-dashed border-zinc-200 p-8 text-center text-xs text-zinc-500 dark:border-border dark:text-muted-foreground">
					Дисциплинарные меры к пользователю не применялись.
				</div>
			{:else}
				<div class="space-y-3">
					{#each data.restrictions as r}
						{@const isUnblock = r.enforcement_type === 'unblock'}
						{@const isBlock = r.enforcement_type === 'permanent_block' || r.enforcement_type === 'temporary_block'}
						<div class="p-4 rounded-2xl border border-zinc-100 bg-zinc-50/80 space-y-2 dark:border-border dark:bg-muted/20">
							<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
								<div class="flex items-center gap-2">
									<span class="px-2.5 py-1 rounded-full text-[11px] font-bold uppercase border {isUnblock ? 'bg-emerald-100 text-emerald-800 border-emerald-200 dark:bg-emerald-950 dark:text-emerald-300' : isBlock ? 'bg-rose-100 text-rose-800 border-rose-200 dark:bg-rose-950 dark:text-rose-300' : 'bg-amber-100 text-amber-800 border-amber-200 dark:bg-amber-950 dark:text-amber-300'}">
										{r.enforcement_type}
									</span>
									{#if r.is_active}
										<span class="text-[10px] font-bold text-emerald-600 flex items-center gap-1">
											<CheckCircle2 class="h-3 w-3" /> Активная мера
										</span>
									{/if}
								</div>
								<span class="text-xs text-zinc-400 font-mono">{formatDate(r.created_at)}</span>
							</div>

							<div class="text-xs space-y-1">
								<p class="font-bold text-zinc-900 dark:text-foreground">Основание: <span class="font-normal text-zinc-600 dark:text-muted-foreground">{r.reason}</span></p>
								{#if r.note}
									<p class="font-bold text-zinc-900 dark:text-foreground">Заметка модератора: <span class="font-normal text-zinc-600 dark:text-muted-foreground">{r.note}</span></p>
								{/if}
								{#if r.expires_at}
									<p class="font-bold text-zinc-900 flex items-center gap-1 dark:text-foreground">
										<Clock class="h-3 w-3 text-zinc-400" />
										<span>Действует до: <span class="font-normal text-zinc-600 dark:text-muted-foreground">{formatDate(r.expires_at)}</span></span>
									</p>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>

{#if data}
	{@const targetFirst = data.user.first_name || ''}
	{@const targetLast = data.user.last_name || ''}
	{@const targetFullName = [targetFirst, targetLast].filter(Boolean).join(' ')}
	<UserBlockDialog
		open={isDialogOpen}
		userName={targetFullName ? `${targetFullName} (${data.user.phone || ''})` : data.user.phone || 'Пользователь'}
		{isSubmitting}
		onConfirm={handleConfirmEnforcement}
		onCancel={() => (isDialogOpen = false)}
	/>
{/if}