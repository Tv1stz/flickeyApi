<script lang="ts">
	import { page } from '$app/state';
	import { adminUsersApi, type UserDetailResponse } from '$lib/api/admin/users';
	import UserBlockDialog from '$lib/components/admin/UserBlockDialog.svelte';
	import type { EnforcementType } from '$lib/types/admin';
	import { formatDate, formatCurrency, translateHousingType, translateListingStatus, cn } from '$lib/utils';
	import {
		ArrowLeft,
		ShieldAlert,
		ShieldCheck,
		User,
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
		Clock
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
			error = err instanceof Error ? err.message : 'Ошибка загрузки профиля пользователя';
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
		setTimeout(() => (copied = false), 2000);
	}

	function handleConfirmEnforcement(payload: { enforcement_type: EnforcementType; reason: string; note: string; duration_days?: number }) {
		if (!data) return;
		isSubmitting = true;
		adminUsersApi
			.enforceUser(data.user.id, payload)
			.then(() => {
				isDialogOpen = false;
				if (userId) loadUserDetail(userId);
			})
			.catch((err) => {
				alert(err.message || 'Ошибка выполнения действия');
			})
			.finally(() => {
				isSubmitting = false;
			});
	}

	function handleQuickUnblock() {
		if (!data) return;
		if (!confirm('Снять все ограничения и разблокировать пользователя ' + data.user.phone + '?')) return;
		adminUsersApi
			.unblockUser(data.user.id, 'Снятие дисциплинарных ограничений из карточки профиля')
			.then(() => {
				if (userId) loadUserDetail(userId);
			})
			.catch((err) => {
				alert(err.message || 'Ошибка разблокировки');
			});
	}

	function handleRoleChange(newRole: string) {
		if (!data) return;
		if (!confirm('Изменить роль пользователя на "' + newRole.toUpperCase() + '"?')) return;
		adminUsersApi
			.updateUserRole(data.user.id, newRole)
			.then(() => {
				if (userId) loadUserDetail(userId);
			})
			.catch((err) => {
				alert(err.message || 'Ошибка изменения роли');
			});
	}
</script>

<svelte:head>
	<title>
		{data?.user ? ('Пользователь ' + data.user.phone) : 'Профиль пользователя'} — Flickey Admin
	</title>
</svelte:head>

<div class="space-y-6 max-w-6xl mx-auto">
	<!-- Top Breadcrumbs & Actions -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<a
			href="/admin/users"
			class="inline-flex items-center gap-2 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors"
		>
			<ArrowLeft class="h-4 w-4" />
			<span>Назад к реестру пользователей</span>
		</a>

		{#if data}
			<div class="flex items-center gap-2 self-start sm:self-auto">
				{#if data.user.status === 'suspended' || data.user.status === 'banned'}
					<button
						type="button"
						onclick={handleQuickUnblock}
						class="px-3.5 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-semibold text-xs transition-colors flex items-center gap-1.5 shadow-xs cursor-pointer"
					>
						<Unlock class="h-4 w-4" />
						<span>Снять ограничения</span>
					</button>
				{/if}

				<button
					type="button"
					onclick={() => (isDialogOpen = true)}
					class="px-4 py-2 rounded-xl bg-amber-600 hover:bg-amber-700 text-white font-semibold text-xs transition-colors flex items-center gap-1.5 shadow-xs cursor-pointer"
				>
					<ShieldAlert class="h-4 w-4" />
					<span>Применить меры</span>
				</button>
			</div>
		{/if}
	</div>

	{#if isLoading}
		<div class="p-16 text-center rounded-2xl border border-border bg-card space-y-3">
			<Loader2 class="h-8 w-8 text-primary animate-spin mx-auto" />
			<p class="text-xs text-muted-foreground font-medium">Загрузка досье пользователя...</p>
		</div>
	{:else if error}
		<div class="p-5 rounded-2xl bg-rose-50 border border-rose-200 text-rose-800 dark:bg-rose-950 dark:border-rose-900 text-xs flex items-center gap-3">
			<AlertTriangle class="h-5 w-5 shrink-0" />
			<div>
				<h4 class="font-bold text-sm">Ошибка загрузки пользователя</h4>
				<p class="mt-0.5">{error}</p>
			</div>
		</div>
	{:else if data}
		{@const firstName = data.user.first_name || ''}
		{@const lastName = data.user.last_name || ''}
		{@const fullName = [firstName, lastName].filter(Boolean).join(' ') || null}
		{@const initial = (fullName ? fullName[0] : (data.user.phone || '?').replace('+', '')[0] || '?').toUpperCase()}

		<!-- Main Dossier Card -->
		<div class="p-6 sm:p-8 rounded-3xl bg-card border border-border shadow-xs space-y-6">
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-border">
				<div class="flex items-center gap-4">
					<div class="w-16 h-16 rounded-2xl flex items-center justify-center font-extrabold text-2xl shadow-inner shrink-0 {data.user.role === 'admin' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' : data.user.role === 'host' ? 'bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-300' : 'bg-primary/10 text-primary'}">
						{initial}
					</div>
					<div>
						<h2 class="text-xl font-extrabold text-foreground">
							{fullName || 'Имя не заполнено'}
						</h2>
						<div class="flex flex-wrap items-center gap-2 mt-1">
							<span class="text-xs font-mono text-muted-foreground">{data.user.phone}</span>
							<span class="text-muted-foreground text-xs">•</span>
							<button
								type="button"
								onclick={copyId}
								class="inline-flex items-center gap-1 text-[11px] font-mono text-muted-foreground hover:text-foreground bg-muted/50 px-2 py-0.5 rounded-md cursor-pointer"
								title="Скопировать ID"
							>
								<span>ID: {data.user.id.slice(0, 8)}...</span>
								{#if copied}
									<Check class="h-3 w-3 text-emerald-600" />
								{:else}
									<Copy class="h-3 w-3" />
								{/if}
							</button>
						</div>
					</div>
				</div>

				<div class="flex items-center gap-3">
					<div>
						<div class="text-[10px] uppercase font-bold text-muted-foreground mb-1">Роль в системе</div>
						<select
							value={data.user.role}
							onchange={(e) => handleRoleChange((e.target as HTMLSelectElement).value)}
							class="px-3 py-1.5 rounded-xl border border-input bg-background text-xs font-bold text-foreground focus:ring-2 focus:ring-primary focus:outline-none cursor-pointer shadow-xs"
						>
							<option value="guest">Guest (Гость)</option>
							<option value="host">Host (Хозяин)</option>
							<option value="admin">Admin (Администратор)</option>
						</select>
					</div>

					<div>
						<div class="text-[10px] uppercase font-bold text-muted-foreground mb-1">Статус</div>
						{#if data.user.status === 'active'}
							<span class="inline-flex items-center gap-1 px-3 py-1.5 rounded-xl text-xs font-bold bg-emerald-100 text-emerald-800 border border-emerald-200 dark:bg-emerald-950 dark:text-emerald-300">
								<ShieldCheck class="h-3.5 w-3.5" /> Активен
							</span>
						{:else if data.user.status === 'suspended'}
							<span class="inline-flex items-center gap-1 px-3 py-1.5 rounded-xl text-xs font-bold bg-amber-100 text-amber-800 border border-amber-200 dark:bg-amber-950 dark:text-amber-300">
								<AlertTriangle class="h-3.5 w-3.5" /> Ограничен
							</span>
						{:else if data.user.status === 'banned'}
							<span class="inline-flex items-center gap-1 px-3 py-1.5 rounded-xl text-xs font-bold bg-rose-100 text-rose-800 border border-rose-200 dark:bg-rose-950 dark:text-rose-300">
								<ShieldAlert class="h-3.5 w-3.5" /> Заблокирован
							</span>
						{:else}
							<span class="inline-flex items-center gap-1 px-3 py-1.5 rounded-xl text-xs font-bold bg-slate-100 text-slate-800 border border-slate-200">
								{data.user.status}
							</span>
						{/if}
					</div>
				</div>
			</div>

			<!-- Meta information grid -->
			<div class="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
				<div class="p-3.5 rounded-2xl bg-muted/30 border border-border/60 space-y-1">
					<div class="text-muted-foreground font-semibold flex items-center gap-1.5">
						<Mail class="h-3.5 w-3.5" /> Email
					</div>
					<div class="font-medium text-foreground">{data.user.email || 'Не привязан'}</div>
				</div>

				<div class="p-3.5 rounded-2xl bg-muted/30 border border-border/60 space-y-1">
					<div class="text-muted-foreground font-semibold flex items-center gap-1.5">
						<Calendar class="h-3.5 w-3.5" /> Дата регистрации
					</div>
					<div class="font-medium text-foreground">{data.user.created_at ? formatDate(data.user.created_at) : '—'}</div>
				</div>

				<div class="p-3.5 rounded-2xl bg-muted/30 border border-border/60 space-y-1">
					<div class="text-muted-foreground font-semibold flex items-center gap-1.5">
						<Building2 class="h-3.5 w-3.5" /> Объектов недвижимости
					</div>
					<div class="font-bold text-foreground text-sm">{data.listings.length}</div>
				</div>
			</div>
		</div>

		<!-- User Owned Listings Section -->
		<div class="p-6 sm:p-8 rounded-3xl bg-card border border-border shadow-xs space-y-4">
			<div class="flex items-center justify-between">
				<h3 class="text-base font-bold text-foreground flex items-center gap-2">
					<Building2 class="h-5 w-5 text-primary" />
					<span>Размещенные объекты недвижимости ({data.listings.length})</span>
				</h3>
			</div>

			{#if data.listings.length === 0}
				<div class="p-8 rounded-2xl border border-dashed border-border text-center text-xs text-muted-foreground">
					У пользователя пока нет созданных объявлений.
				</div>
			{:else}
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
					{#each data.listings as listing}
						{@const statusInfo = translateListingStatus(listing.status || 'draft')}
						<div class="rounded-2xl border border-border bg-background p-4 space-y-3 shadow-xs hover:border-primary/50 transition-colors">
							{#if listing.media && listing.media.length > 0}
								<img
									src={listing.media[0]}
									alt={listing.name}
									class="w-full h-36 rounded-xl object-cover bg-muted"
								/>
							{:else}
								<div class="w-full h-36 rounded-xl bg-muted flex items-center justify-center text-xs text-muted-foreground font-medium">
									Нет фото
								</div>
							{/if}

							<div class="space-y-1">
								<div class="flex items-center justify-between gap-2">
									<span class="text-[10px] font-bold text-muted-foreground uppercase">
										{translateHousingType(listing.type)}
									</span>
									<span class={cn('px-2 py-0.5 rounded-full text-[10px] font-bold border', statusInfo.color)}>
										{statusInfo.label}
									</span>
								</div>
								<h4 class="font-bold text-xs text-foreground line-clamp-1">{listing.name}</h4>
								<div class="text-xs font-extrabold text-foreground">
									{formatCurrency(listing.price_per_night, listing.currency)} <span class="text-[10px] font-normal text-muted-foreground">/ сутки</span>
								</div>
							</div>

							<div class="pt-2 border-t border-border flex items-center justify-between gap-2">
								<a
									href="/admin/listings/{listing.id}"
									class="w-full py-1.5 rounded-xl border border-border hover:bg-muted text-xs font-semibold text-foreground text-center inline-flex items-center justify-center gap-1 transition-colors"
								>
									<ExternalLink class="h-3.5 w-3.5" />
									<span>Модерация</span>
								</a>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Restriction & Enforcement History Timeline -->
		<div class="p-6 sm:p-8 rounded-3xl bg-card border border-border shadow-xs space-y-4">
			<div class="flex items-center justify-between">
				<h3 class="text-base font-bold text-foreground flex items-center gap-2">
					<History class="h-5 w-5 text-primary" />
					<span>История дисциплинарных мер ({data.restrictions.length})</span>
				</h3>
			</div>

			{#if data.restrictions.length === 0}
				<div class="p-8 rounded-2xl border border-dashed border-border text-center text-xs text-muted-foreground">
					Дисциплинарные взыскания и ограничения ранее не накладывались.
				</div>
			{:else}
				<div class="space-y-3">
					{#each data.restrictions as r}
						{@const isUnblock = r.enforcement_type === 'unblock'}
						{@const isBlock = r.enforcement_type === 'permanent_block' || r.enforcement_type === 'temporary_block'}
						<div class="p-4 rounded-2xl border border-border bg-muted/20 space-y-2">
							<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
								<div class="flex items-center gap-2">
									<span class="px-2.5 py-1 rounded-full text-[11px] font-extrabold uppercase border {isUnblock ? 'bg-emerald-100 text-emerald-800 border-emerald-200 dark:bg-emerald-950 dark:text-emerald-300' : isBlock ? 'bg-rose-100 text-rose-800 border-rose-200 dark:bg-rose-950 dark:text-rose-300' : 'bg-amber-100 text-amber-800 border-amber-200 dark:bg-amber-950 dark:text-amber-300'}">
										{r.enforcement_type}
									</span>
									{#if r.is_active}
										<span class="text-[10px] font-bold text-emerald-600 flex items-center gap-1">
											<CheckCircle2 class="h-3 w-3" /> Активная мера
										</span>
									{/if}
								</div>
								<span class="text-xs text-muted-foreground font-mono">{formatDate(r.created_at)}</span>
							</div>

							<div class="text-xs space-y-1">
								<p class="font-bold text-foreground">Причина: <span class="font-normal text-muted-foreground">{r.reason}</span></p>
								{#if r.note}
									<p class="font-bold text-foreground">Примечание администратора: <span class="font-normal text-muted-foreground">{r.note}</span></p>
								{/if}
								{#if r.expires_at}
									<p class="font-bold text-foreground flex items-center gap-1">
										<Clock class="h-3 w-3 text-muted-foreground" />
										<span>Действует до: <span class="font-normal text-muted-foreground">{formatDate(r.expires_at)}</span></span>
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
