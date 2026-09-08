<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth/auth.svelte';
	import { adminUsersApi } from '$lib/api/admin/users';
	import type { User } from '$lib/types/auth';
	import type { EnforcementType } from '$lib/types/admin';
	import UserBlockDialog from '$lib/components/admin/UserBlockDialog.svelte';
	import AdminActionMenu from '$lib/components/admin/AdminActionMenu.svelte';
	import { toast } from '$lib/stores/toastStore';
	import { formatDate } from '$lib/utils';
	import {
		Users,
		ShieldAlert,
		ShieldCheck,
		UserCheck,
		UserX,
		Eye,
		Search,
		Loader2,
		Copy,
		Check,
		Mail,
		Phone,
		Calendar,
		AlertTriangle,
		Unlock,
		RotateCw,
		X,
		ArrowRight
	} from 'lucide-svelte';

	let users = $state<User[]>([]);
	let activeTab = $state<'all' | 'host' | 'guest' | 'admin' | 'restricted'>('all');
	let searchQuery = $state('');
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let copiedId = $state<string | null>(null);

	// User Block Dialog state
	let isDialogOpen = $state(false);
	let targetUser = $state<User | null>(null);
	let isSubmitting = $state(false);

	async function loadUsers() {
		isLoading = true;
		error = null;
		try {
			const res = await adminUsersApi.getUsers();
			if (Array.isArray(res)) {
				users = res;
			} else if (res && typeof res === 'object' && Array.isArray((res as { data?: User[] }).data)) {
				users = (res as { data: User[] }).data;
			} else {
				users = [];
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Не удалось загрузить реестр пользователей';
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		loadUsers();
	});

	let metrics = $derived({
		total: users.length,
		hosts: users.filter((u) => u.role === 'host').length,
		guests: users.filter((u) => u.role === 'guest').length,
		admins: users.filter((u) => u.role === 'admin').length,
		restricted: users.filter((u) => u.status === 'suspended' || u.status === 'banned').length
	});

	let filteredUsers = $derived(
		users.filter((u) => {
			if (activeTab === 'host' && u.role !== 'host') return false;
			if (activeTab === 'guest' && u.role !== 'guest') return false;
			if (activeTab === 'admin' && u.role !== 'admin') return false;
			if (activeTab === 'restricted' && u.status !== 'suspended' && u.status !== 'banned') return false;

			if (!searchQuery.trim()) return true;
			const q = searchQuery.toLowerCase().trim();
			const phoneMatch = Boolean(u.phone?.toLowerCase().includes(q));
			const firstName = u.first_name || '';
			const lastName = u.last_name || '';
			const fullName = `${firstName} ${lastName}`.trim();
			const nameMatch =
				Boolean(firstName.toLowerCase().includes(q)) ||
				Boolean(lastName.toLowerCase().includes(q)) ||
				Boolean(fullName.toLowerCase().includes(q));
			const emailMatch = Boolean(u.email?.toLowerCase().includes(q));
			const idMatch = Boolean(u.id?.toLowerCase().includes(q));

			return phoneMatch || nameMatch || emailMatch || idMatch;
		})
	);

	function copyToClipboard(e: MouseEvent, text: string, id: string) {
		e.stopPropagation();
		navigator.clipboard.writeText(text);
		copiedId = id;
		toast.success('Скопировано в буфер обмена');
		setTimeout(() => {
			if (copiedId === id) copiedId = null;
		}, 2000);
	}

	function openBlockDialog(u: User) {
		targetUser = u;
		isDialogOpen = true;
	}

	function handleConfirmEnforcement(payload: {
		enforcement_type: EnforcementType;
		reason: string;
		note: string;
		duration_days?: number;
	}) {
		if (!targetUser) return;
		isSubmitting = true;
		adminUsersApi
			.enforceUser(targetUser.id, payload)
			.then(() => {
				isDialogOpen = false;
				toast.success(
					payload.enforcement_type === 'unblock'
						? 'Пользователь успешно разблокирован'
						: 'Санкции успешно применены'
				);
				loadUsers();
			})
			.catch((err) => {
				toast.error('Ошибка применения мер', err.message || 'Не удалось обновить статус');
			})
			.finally(() => {
				isSubmitting = false;
			});
	}

	function handleRoleChange(u: User, newRole: string) {
		adminUsersApi
			.updateUserRole(u.id, newRole)
			.then(() => {
				toast.success(`Роль пользователя изменена на: ${newRole}`);
				loadUsers();
			})
			.catch((err) => {
				toast.error('Ошибка смены роли', err.message || 'Не удалось сменить роль');
			});
	}

	function handleQuickUnblock(u: User) {
		adminUsersApi
			.unblockUser(u.id, 'Административная отмена ограничений')
			.then(() => {
				toast.success('Пользователь разблокирован');
				loadUsers();
			})
			.catch((err) => {
				toast.error('Ошибка разблокировки', err.message || 'Не удалось разблокировать');
			});
	}

	function navigateToUser(id: string) {
		goto(`/admin/users/${id}`);
	}
</script>

<svelte:head>
	<title>Реестр пользователей — Flickey Admin</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header & Search -->
	<div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
		<div>
			<div class="flex items-center gap-2.5">
				<h1 class="text-2xl font-bold tracking-tight text-zinc-900 dark:text-foreground">
					Пользователи
				</h1>
				<span class="flex h-6 min-w-[24px] items-center justify-center rounded-full bg-zinc-900 px-2 text-xs font-bold text-white dark:bg-white dark:text-zinc-900">
					{filteredUsers.length}
				</span>
			</div>
			<p class="mt-1 text-sm text-zinc-500 dark:text-muted-foreground">
				Управление аккаунтами, ролями, правами доступа и административными ограничениями.
			</p>
		</div>

		<!-- Search Input -->
		<div class="relative w-full md:w-80">
			<Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-400" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Поиск по имени, телефону, email, ID..."
				class="w-full rounded-full border border-zinc-200/80 bg-white py-2 pl-9 pr-9 text-xs text-zinc-900 placeholder:text-zinc-400 focus:border-zinc-900 focus:outline-none focus:ring-1 focus:ring-zinc-900 dark:border-border dark:bg-card dark:text-foreground shadow-2xs"
			/>
			{#if searchQuery}
				<button
					type="button"
					onclick={() => (searchQuery = '')}
					class="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-200"
				>
					<X class="h-3.5 w-3.5" />
				</button>
			{/if}
		</div>
	</div>

	<!-- Filter Tabs with Count Pills -->
	<div class="flex items-center gap-1.5 overflow-x-auto pb-1 scrollbar-none">
		{#each [
			{ id: 'all', label: 'Все', count: metrics.total },
			{ id: 'host', label: 'Хозяева', count: metrics.hosts },
			{ id: 'guest', label: 'Гости', count: metrics.guests },
			{ id: 'admin', label: 'Администраторы', count: metrics.admins },
			{ id: 'restricted', label: 'С ограничениями', count: metrics.restricted }
		] as tab}
			{@const active = activeTab === tab.id}
			<button
				type="button"
				onclick={() => (activeTab = tab.id as any)}
				class="inline-flex items-center gap-1.5 whitespace-nowrap rounded-full px-4 py-2 text-xs font-semibold transition-all cursor-pointer {active
					? 'bg-zinc-900 text-white shadow-xs dark:bg-white dark:text-zinc-900'
					: 'border border-zinc-200/80 bg-white text-zinc-600 hover:bg-zinc-50 hover:text-zinc-900 hover:border-zinc-300 dark:border-border dark:bg-card dark:text-muted-foreground dark:hover:text-foreground'}"
			>
				<span>{tab.label}</span>
				<span class="rounded-full px-1.5 py-0.2 text-[10px] font-bold {active ? 'bg-white/20 text-white dark:bg-zinc-900/20 dark:text-zinc-900' : 'bg-zinc-100 text-zinc-600 dark:bg-muted dark:text-muted-foreground'}">
					{tab.count}
				</span>
			</button>
		{/each}
	</div>

	<!-- Content Table -->
	{#if isLoading}
		<div class="rounded-3xl border border-zinc-200/80 bg-white p-16 text-center shadow-2xs dark:border-border dark:bg-card">
			<Loader2 class="h-8 w-8 text-zinc-900 animate-spin mx-auto mb-3 dark:text-white" />
			<p class="text-xs font-medium text-zinc-500 dark:text-muted-foreground">Загрузка пользователей...</p>
		</div>
	{:else if error}
		<div class="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-800 dark:border-rose-900 dark:bg-rose-950/60 dark:text-rose-200 flex items-center justify-between">
			<span>{error}</span>
			<button
				type="button"
				onclick={loadUsers}
				class="font-semibold underline hover:no-underline cursor-pointer"
			>
				Повторить
			</button>
		</div>
	{:else if filteredUsers.length === 0}
		<div class="rounded-3xl border border-dashed border-zinc-200 bg-white p-16 text-center shadow-2xs dark:border-border dark:bg-card space-y-3">
			<div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-400 mx-auto dark:bg-muted">
				<Users class="h-7 w-7 text-zinc-400" />
			</div>
			<h3 class="text-base font-bold text-zinc-900 dark:text-foreground">
				Пользователи не найдены
			</h3>
			<p class="text-xs text-zinc-500 max-w-sm mx-auto dark:text-muted-foreground">
				{searchQuery ? 'По вашему запросу не найдено ни одного пользователя.' : 'В выбранной вкладке нет пользователей.'}
			</p>
			{#if searchQuery}
				<button
					type="button"
					onclick={() => (searchQuery = '')}
					class="mt-2 inline-flex items-center gap-1.5 rounded-full bg-zinc-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-zinc-800 dark:bg-white dark:text-zinc-900"
				>
					Сбросить поиск
				</button>
			{/if}
		</div>
	{:else}
		<div class="rounded-3xl border border-zinc-200/80 bg-white overflow-hidden shadow-2xs dark:border-border dark:bg-card">
			<div class="overflow-x-auto">
				<table class="w-full text-left border-collapse">
					<thead>
						<tr class="border-b border-zinc-100 bg-zinc-50/70 text-[11px] font-bold text-zinc-500 uppercase tracking-wider dark:border-border dark:bg-muted/40 dark:text-muted-foreground">
							<th class="py-4 px-5">Пользователь</th>
							<th class="py-4 px-4">Контакты</th>
							<th class="py-4 px-4">Роль</th>
							<th class="py-4 px-4">Статус</th>
							<th class="py-4 px-4">Дата регистрации</th>
							<th class="py-4 px-5 text-right">Действия</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-zinc-100 text-xs dark:divide-border">
						{#each filteredUsers as u (u.id)}
							{@const first_name = u.first_name || ''}
							{@const last_name = u.last_name || ''}
							{@const full_name = [first_name, last_name].filter(Boolean).join(' ') || null}
							{@const initial = (full_name ? full_name[0] : (u.phone || '?').replace('+', '')[0] || '?').toUpperCase()}
							<tr
								onclick={() => navigateToUser(u.id)}
								class="group cursor-pointer transition-colors hover:bg-zinc-50/80 dark:hover:bg-muted/30"
							>
								<!-- User Identity -->
								<td class="py-4 px-5">
									<div class="flex items-center gap-3.5">
										<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl font-bold text-sm shadow-2xs {u.role === 'admin' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' : u.role === 'host' ? 'bg-zinc-900 text-white dark:bg-white dark:text-zinc-900' : 'bg-zinc-100 text-zinc-700 dark:bg-muted dark:text-foreground'}">
											{initial}
										</div>
										<div class="min-w-0">
											<div class="font-bold text-zinc-900 group-hover:text-zinc-600 transition-colors truncate max-w-[200px] dark:text-foreground">
												{full_name || 'Без имени'}
											</div>
											<div class="flex items-center gap-1.5 text-[11px] text-zinc-400 font-mono mt-0.5">
												<span>ID: {u.id?.slice(0, 8)}</span>
												<button
													type="button"
													onclick={(e) => copyToClipboard(e, u.id, u.id)}
													class="hover:text-zinc-700 transition-colors cursor-pointer dark:hover:text-zinc-200"
													title="Копировать ID"
												>
													{#if copiedId === u.id}
														<Check class="h-3 w-3 text-emerald-600" />
													{:else}
														<Copy class="h-3 w-3" />
													{/if}
												</button>
											</div>
										</div>
									</div>
								</td>

								<!-- Contact Details -->
								<td class="py-4 px-4">
									<div class="space-y-0.5">
										<div class="font-mono font-medium text-zinc-800 flex items-center gap-1.5 dark:text-zinc-200">
											<Phone class="h-3 w-3 text-zinc-400 shrink-0" />
											<span>{u.phone || '—'}</span>
										</div>
										{#if u.email}
											<div class="text-[11px] text-zinc-500 flex items-center gap-1.5 truncate max-w-[180px] dark:text-muted-foreground">
												<Mail class="h-3 w-3 shrink-0 text-zinc-400" />
												<span class="truncate">{u.email}</span>
											</div>
										{/if}
									</div>
								</td>

								<!-- Role Badge -->
								<td class="py-4 px-4">
									<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-bold {u.role === 'admin' ? 'bg-amber-100 text-amber-800 border border-amber-200 dark:bg-amber-950 dark:text-amber-300' : u.role === 'host' ? 'bg-zinc-100 text-zinc-800 border border-zinc-200 dark:bg-muted dark:text-foreground' : 'bg-zinc-50 text-zinc-600 border border-zinc-200/80 dark:bg-muted/40 dark:text-muted-foreground'}">
										{u.role === 'admin' ? 'Администратор' : u.role === 'host' ? 'Хозяин' : 'Гость'}
									</span>
								</td>

								<!-- Status Badge -->
								<td class="py-4 px-4">
									{#if u.status === 'active'}
										<span class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[11px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-950 dark:border-emerald-900 dark:text-emerald-300">
											<ShieldCheck class="h-3 w-3" /> Активен
										</span>
									{:else if u.status === 'suspended'}
										<span class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[11px] font-bold bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-950 dark:border-amber-900 dark:text-amber-300">
											<AlertTriangle class="h-3 w-3" /> Ограничен
										</span>
									{:else if u.status === 'banned'}
										<span class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[11px] font-bold bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-950 dark:border-rose-900 dark:text-rose-300">
											<UserX class="h-3 w-3" /> Заблокирован
										</span>
									{:else}
										<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-bold bg-zinc-100 text-zinc-700 border border-zinc-200 dark:bg-muted dark:text-muted-foreground">
											{u.status}
										</span>
									{/if}
								</td>

								<!-- Registration Date -->
								<td class="py-4 px-4 text-zinc-500 whitespace-nowrap dark:text-muted-foreground">
									{u.created_at ? formatDate(u.created_at) : '—'}
								</td>

								<!-- Clean Action Menu -->
								<td class="py-4 px-5 text-right" onclick={(e) => e.stopPropagation()}>
									<div class="flex items-center justify-end gap-1.5">
										<AdminActionMenu
											items={[
												{
													label: 'Открыть профиль',
													icon: Eye,
													variant: 'primary',
													onclick: () => navigateToUser(u.id)
												},
												...(u.status === 'suspended' || u.status === 'banned'
													? [
															{
																label: 'Разблокировать пользователя',
																icon: Unlock,
																variant: 'success' as const,
																onclick: () => handleQuickUnblock(u)
															}
														]
													: []),
												{
													label: 'Применить санкции',
													icon: ShieldAlert,
													variant: 'warning',
													onclick: () => openBlockDialog(u)
												},
												{
													label: u.role === 'host' ? 'Переключить в Гостя' : 'Сделать Хозяином',
													divider: true,
													onclick: () => handleRoleChange(u, u.role === 'host' ? 'guest' : 'host')
												},
												{
													label: u.role === 'admin' ? 'Снять права админа' : 'Назначить администратором',
													onclick: () => handleRoleChange(u, u.role === 'admin' ? 'host' : 'admin')
												},
												{
													label: 'Копировать ID',
													icon: Copy,
													divider: true,
													onclick: () => {
														navigator.clipboard.writeText(u.id);
														toast.success('ID скопирован');
													}
												}
											]}
										/>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

{#if targetUser}
	{@const targetFirst = targetUser.first_name || ''}
	{@const targetLast = targetUser.last_name || ''}
	{@const targetFullName = [targetFirst, targetLast].filter(Boolean).join(' ')}
	<UserBlockDialog
		open={isDialogOpen}
		userName={targetFullName ? `${targetFullName} (${targetUser.phone || ''})` : targetUser.phone || 'Пользователь'}
		{isSubmitting}
		onConfirm={handleConfirmEnforcement}
		onCancel={() => (isDialogOpen = false)}
	/>
{/if}