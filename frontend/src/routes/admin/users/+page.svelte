<script lang="ts">
	import { auth } from '$lib/auth/auth.svelte';
	import { adminUsersApi } from '$lib/api/admin/users';
	import type { User } from '$lib/types/auth';
	import type { EnforcementType } from '$lib/types/admin';
	import UserBlockDialog from '$lib/components/admin/UserBlockDialog.svelte';
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
		ArrowUpDown,
		Mail,
		Phone,
		Calendar,
		Building2,
		AlertTriangle,
		Unlock,
		RefreshCw
	} from 'lucide-svelte';

	let users = $state<User[]>([]);
	let activeTab = $state<'all' | 'host' | 'guest' | 'admin' | 'active' | 'restricted' | 'pending'>('all');
	let searchQuery = $state('');
	let sortBy = $state<'date_desc' | 'date_asc' | 'name' | 'role'>('date_desc');
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
			error = err instanceof Error ? err.message : 'Ошибка загрузки реестра пользователей';
		} finally {
			isLoading = false;
		}
	}

	$effect(() => {
		if (auth.user?.id) {
			loadUsers();
		}
	});

	// Metrics calculations
	let metrics = $derived({
		total: users.length,
		active: users.filter((u) => u.status === 'active').length,
		hosts: users.filter((u) => u.role === 'host').length,
		guests: users.filter((u) => u.role === 'guest').length,
		admins: users.filter((u) => u.role === 'admin').length,
		restricted: users.filter((u) => u.status === 'suspended' || u.status === 'banned').length,
		pending: users.filter((u) => u.status === 'pending_profile').length
	});

	// Filtered and Sorted Users
	let filteredUsers = $derived(
		users
			.filter((u) => {
				if (activeTab === 'host' && u.role !== 'host') return false;
				if (activeTab === 'guest' && u.role !== 'guest') return false;
				if (activeTab === 'admin' && u.role !== 'admin') return false;
				if (activeTab === 'active' && u.status !== 'active') return false;
				if (activeTab === 'restricted' && u.status !== 'suspended' && u.status !== 'banned') return false;
				if (activeTab === 'pending' && u.status !== 'pending_profile') return false;

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
			.sort((a, b) => {
				const dateA = new Date(a.created_at || 0).getTime() || 0;
				const dateB = new Date(b.created_at || 0).getTime() || 0;

				if (sortBy === 'date_desc') return dateB - dateA;
				if (sortBy === 'date_asc') return dateA - dateB;
				if (sortBy === 'name') {
					const nameA = [a.first_name, a.last_name].filter(Boolean).join(' ') || a.phone || '';
					const nameB = [b.first_name, b.last_name].filter(Boolean).join(' ') || b.phone || '';
					return nameA.localeCompare(nameB);
				}
				if (sortBy === 'role') return (a.role || '').localeCompare(b.role || '');
				return 0;
			})
	);

	function copyToClipboard(text: string, id: string) {
		navigator.clipboard.writeText(text);
		copiedId = id;
		setTimeout(() => {
			if (copiedId === id) copiedId = null;
		}, 2000);
	}

	function openBlockDialog(u: User) {
		targetUser = u;
		isDialogOpen = true;
	}

	function handleQuickUnblock(u: User) {
		if (!confirm('Снять все ограничения и разблокировать пользователя ' + (u.phone || u.id) + '?')) return;
		adminUsersApi
			.unblockUser(u.id, 'Снятие ограничений через реестр пользователей')
			.then(() => {
				loadUsers();
			})
			.catch((err) => {
				alert(err.message || 'Ошибка разблокировки пользователя');
			});
	}

	function handleConfirmEnforcement(data: { enforcement_type: EnforcementType; reason: string; note: string; duration_days?: number }) {
		if (!targetUser) return;
		isSubmitting = true;
		adminUsersApi
			.enforceUser(targetUser.id, data)
			.then(() => {
				isDialogOpen = false;
				targetUser = null;
				loadUsers();
			})
			.catch((err) => {
				alert(err.message || 'Ошибка применения дисциплинарной меры');
			})
			.finally(() => {
				isSubmitting = false;
			});
	}

	function handleRoleChange(user: User, newRole: string) {
		if (!confirm('Вы действительно хотите изменить роль пользователя ' + (user.phone || user.id) + ' на "' + newRole.toUpperCase() + '"?')) return;
		adminUsersApi
			.updateUserRole(user.id, newRole)
			.then(() => {
				loadUsers();
			})
			.catch((err) => {
				alert(err.message || 'Ошибка изменения роли');
			});
	}
</script>

<svelte:head>
	<title>Управление пользователями — Flickey Admin</title>
</svelte:head>

<div class="space-y-6">
	<!-- Top Title Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-foreground tracking-tight flex items-center gap-2.5">
				<Users class="h-6 w-6 text-primary" />
				<span>Реестр пользователей</span>
			</h1>
			<p class="text-xs sm:text-sm text-muted-foreground mt-0.5">
				Управление учетными записями, ролями, дисциплинарными мерами и безопасностью платформы.
			</p>
		</div>

		<button
			type="button"
			onclick={loadUsers}
			disabled={isLoading}
			class="inline-flex items-center gap-2 px-4 py-2 rounded-xl border border-border bg-card hover:bg-muted text-xs font-semibold text-foreground transition-colors disabled:opacity-50 self-start sm:self-auto cursor-pointer shadow-xs"
		>
			<RefreshCw class="h-3.5 w-3.5 {isLoading ? 'animate-spin text-primary' : ''}" />
			<span>Обновить</span>
		</button>
	</div>

	<!-- KPI Metrics Grid -->
	<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
		<button
			type="button"
			onclick={() => (activeTab = 'all')}
			class="p-4 rounded-2xl border text-left transition-all cursor-pointer {activeTab === 'all' ? 'border-primary bg-primary/5 ring-2 ring-primary/20' : 'border-border bg-card hover:border-border/80'}"
		>
			<div class="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">Всего</div>
			<div class="text-2xl font-extrabold text-foreground mt-1">{metrics.total}</div>
		</button>

		<button
			type="button"
			onclick={() => (activeTab = 'active')}
			class="p-4 rounded-2xl border text-left transition-all cursor-pointer {activeTab === 'active' ? 'border-emerald-500 bg-emerald-50/50 dark:bg-emerald-950/20 ring-2 ring-emerald-500/20' : 'border-border bg-card hover:border-border/80'}"
		>
			<div class="text-[11px] font-semibold text-emerald-600 dark:text-emerald-400 uppercase tracking-wider flex items-center gap-1">
				<UserCheck class="h-3 w-3" /> Активные
			</div>
			<div class="text-2xl font-extrabold text-emerald-600 dark:text-emerald-400 mt-1">{metrics.active}</div>
		</button>

		<button
			type="button"
			onclick={() => (activeTab = 'host')}
			class="p-4 rounded-2xl border text-left transition-all cursor-pointer {activeTab === 'host' ? 'border-sky-500 bg-sky-50/50 dark:bg-sky-950/20 ring-2 ring-sky-500/20' : 'border-border bg-card hover:border-border/80'}"
		>
			<div class="text-[11px] font-semibold text-sky-600 dark:text-sky-400 uppercase tracking-wider flex items-center gap-1">
				<Building2 class="h-3 w-3" /> Хосты
			</div>
			<div class="text-2xl font-extrabold text-sky-600 dark:text-sky-400 mt-1">{metrics.hosts}</div>
		</button>

		<button
			type="button"
			onclick={() => (activeTab = 'guest')}
			class="p-4 rounded-2xl border text-left transition-all cursor-pointer {activeTab === 'guest' ? 'border-violet-500 bg-violet-50/50 dark:bg-violet-950/20 ring-2 ring-violet-500/20' : 'border-border bg-card hover:border-border/80'}"
		>
			<div class="text-[11px] font-semibold text-violet-600 dark:text-violet-400 uppercase tracking-wider">Гости</div>
			<div class="text-2xl font-extrabold text-violet-600 dark:text-violet-400 mt-1">{metrics.guests}</div>
		</button>

		<button
			type="button"
			onclick={() => (activeTab = 'restricted')}
			class="p-4 rounded-2xl border text-left transition-all cursor-pointer {activeTab === 'restricted' ? 'border-rose-500 bg-rose-50/50 dark:bg-rose-950/20 ring-2 ring-rose-500/20' : 'border-border bg-card hover:border-border/80'}"
		>
			<div class="text-[11px] font-semibold text-rose-600 dark:text-rose-400 uppercase tracking-wider flex items-center gap-1">
				<ShieldAlert class="h-3 w-3" /> Ограничены
			</div>
			<div class="text-2xl font-extrabold text-rose-600 dark:text-rose-400 mt-1">{metrics.restricted}</div>
		</button>

		<button
			type="button"
			onclick={() => (activeTab = 'admin')}
			class="p-4 rounded-2xl border text-left transition-all cursor-pointer {activeTab === 'admin' ? 'border-amber-500 bg-amber-50/50 dark:bg-amber-950/20 ring-2 ring-amber-500/20' : 'border-border bg-card hover:border-border/80'}"
		>
			<div class="text-[11px] font-semibold text-amber-600 dark:text-amber-400 uppercase tracking-wider">Админы</div>
			<div class="text-2xl font-extrabold text-amber-600 dark:text-amber-400 mt-1">{metrics.admins}</div>
		</button>
	</div>

	<!-- Filter Tabs Navigation -->
	<div class="flex items-center gap-2 border-b border-border overflow-x-auto pb-px">
		<button
			type="button"
			onclick={() => (activeTab = 'all')}
			class="px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors whitespace-nowrap cursor-pointer {activeTab === 'all' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}"
		>
			Все ({metrics.total})
		</button>
		<button
			type="button"
			onclick={() => (activeTab = 'host')}
			class="px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors whitespace-nowrap cursor-pointer {activeTab === 'host' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}"
		>
			Хосты ({metrics.hosts})
		</button>
		<button
			type="button"
			onclick={() => (activeTab = 'guest')}
			class="px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors whitespace-nowrap cursor-pointer {activeTab === 'guest' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}"
		>
			Гости ({metrics.guests})
		</button>
		<button
			type="button"
			onclick={() => (activeTab = 'restricted')}
			class="px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors whitespace-nowrap cursor-pointer {activeTab === 'restricted' ? 'border-rose-500 text-rose-600 dark:text-rose-400' : 'border-transparent text-muted-foreground hover:text-foreground'}"
		>
			С ограничениями ({metrics.restricted})
		</button>
		<button
			type="button"
			onclick={() => (activeTab = 'pending')}
			class="px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors whitespace-nowrap cursor-pointer {activeTab === 'pending' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}"
		>
			Ожидают профиль ({metrics.pending})
		</button>
		<button
			type="button"
			onclick={() => (activeTab = 'admin')}
			class="px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors whitespace-nowrap cursor-pointer {activeTab === 'admin' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}"
		>
			Администраторы ({metrics.admins})
		</button>
	</div>

	<!-- Search & Sort Controls Bar -->
	<div class="p-4 rounded-2xl bg-card border border-border flex flex-col sm:flex-row items-center justify-between gap-3 shadow-xs">
		<div class="relative w-full sm:flex-1">
			<Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Поиск по номеру телефона, имени, email или ID..."
				class="w-full pl-10 pr-8 py-2 rounded-xl border border-input bg-background text-xs text-foreground focus:ring-2 focus:ring-primary focus:outline-none placeholder:text-muted-foreground/70"
			/>
			{#if searchQuery}
				<button
					type="button"
					onclick={() => (searchQuery = '')}
					class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground text-xs cursor-pointer"
				>
					✕
				</button>
			{/if}
		</div>

		<div class="flex items-center gap-2 w-full sm:w-auto shrink-0 justify-between sm:justify-start">
			<div class="flex items-center gap-1.5 text-xs text-muted-foreground shrink-0">
				<ArrowUpDown class="h-3.5 w-3.5" />
				<span>Сортировка:</span>
			</div>
			<select
				bind:value={sortBy}
				class="px-3 py-2 rounded-xl border border-input bg-background text-xs font-medium text-foreground focus:ring-2 focus:ring-primary focus:outline-none cursor-pointer"
			>
				<option value="date_desc">Сначала новые</option>
				<option value="date_asc">Сначала старые</option>
				<option value="name">По имени (А-Я)</option>
				<option value="role">По роли</option>
			</select>
		</div>
	</div>

	<!-- Content Area -->
	{#if isLoading}
		<div class="p-16 rounded-2xl border border-border bg-card text-center space-y-3">
			<Loader2 class="h-8 w-8 text-primary animate-spin mx-auto" />
			<p class="text-xs text-muted-foreground font-medium">Загрузка реестра пользователей...</p>
		</div>
	{:else if error}
		<div class="p-5 rounded-2xl bg-rose-50 border border-rose-200 text-rose-800 dark:bg-rose-950 dark:border-rose-900 dark:text-rose-200 text-xs flex items-center justify-between gap-3">
			<div class="flex items-center gap-3">
				<AlertTriangle class="h-5 w-5 shrink-0" />
				<div>
					<h4 class="font-bold text-sm">Ошибка при получении данных</h4>
					<p class="mt-0.5">{error}</p>
				</div>
			</div>
			<button
				type="button"
				onclick={loadUsers}
				class="px-3 py-1.5 rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-semibold text-xs shrink-0 cursor-pointer"
			>
				Повторить
			</button>
		</div>
	{:else if filteredUsers.length === 0}
		<div class="p-16 rounded-2xl border border-dashed border-border bg-card text-center space-y-3">
			<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
				<Users class="h-6 w-6" />
			</div>
			<h3 class="text-sm font-bold text-foreground">Пользователи не найдены</h3>
			<p class="text-xs text-muted-foreground max-w-sm mx-auto">
				По вашему запросу и выбранным фильтрам ничего не найдено.
			</p>
			{#if searchQuery}
				<button
					type="button"
					onclick={() => (searchQuery = '')}
					class="px-3.5 py-1.5 rounded-xl border border-border bg-background text-xs font-semibold text-foreground hover:bg-muted transition-colors cursor-pointer"
				>
					Сбросить поиск
				</button>
			{/if}
		</div>
	{:else}
		<!-- Users Table -->
		<div class="rounded-2xl border border-border bg-card overflow-hidden shadow-xs">
			<div class="overflow-x-auto">
				<table class="w-full text-left border-collapse">
					<thead>
						<tr class="border-b border-border bg-muted/40 text-[11px] font-bold text-muted-foreground uppercase tracking-wider">
							<th class="py-3.5 px-4">Пользователь</th>
							<th class="py-3.5 px-4">Контакты</th>
							<th class="py-3.5 px-4">Роль</th>
							<th class="py-3.5 px-4">Статус</th>
							<th class="py-3.5 px-4">Регистрация</th>
							<th class="py-3.5 px-4 text-right">Действия</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-border text-xs">
						{#each filteredUsers as u (u.id)}
							{@const first_name = u.first_name || ''}
							{@const last_name = u.last_name || ''}
							{@const full_name = [first_name, last_name].filter(Boolean).join(' ') || null}
							{@const initial = (full_name ? full_name[0] : (u.phone || '?').replace('+', '')[0] || '?').toUpperCase()}
							<tr class="hover:bg-muted/20 transition-colors">
								<!-- User Identity -->
								<td class="py-3.5 px-4">
									<div class="flex items-center gap-3">
										<div class="flex h-10 w-10 items-center justify-center rounded-xl font-extrabold text-sm shrink-0 {u.role === 'admin' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' : u.role === 'host' ? 'bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-300' : 'bg-primary/10 text-primary'}">
											{initial}
										</div>
										<div class="min-w-0">
											<div class="font-bold text-foreground truncate max-w-[200px]">
												{full_name || 'Имя не указано'}
											</div>
											<div class="flex items-center gap-1.5 text-[11px] text-muted-foreground font-mono mt-0.5">
												<span>ID: {u.id?.slice(0, 8) || '—'}</span>
												{#if u.id}
													<button
														type="button"
														onclick={() => copyToClipboard(u.id, u.id)}
														class="text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
														title="Скопировать полный ID"
													>
														{#if copiedId === u.id}
															<Check class="h-3 w-3 text-emerald-600" />
														{:else}
															<Copy class="h-3 w-3" />
														{/if}
													</button>
												{/if}
											</div>
										</div>
									</div>
								</td>

								<!-- Contact details -->
								<td class="py-3.5 px-4">
									<div class="space-y-1">
										<div class="font-mono font-semibold text-foreground flex items-center gap-1.5">
											<Phone class="h-3 w-3 text-muted-foreground shrink-0" />
											<span>{u.phone || '—'}</span>
										</div>
										{#if u.email}
											<div class="text-[11px] text-muted-foreground flex items-center gap-1.5 truncate max-w-[180px]">
												<Mail class="h-3 w-3 shrink-0" />
												<span class="truncate">{u.email}</span>
											</div>
										{/if}
									</div>
								</td>

								<!-- Role Switcher -->
								<td class="py-3.5 px-4">
									<select
										value={u.role}
										onchange={(e) => handleRoleChange(u, (e.target as HTMLSelectElement).value)}
										class="px-2.5 py-1.5 rounded-xl border border-input bg-background text-[11px] font-bold text-foreground focus:ring-2 focus:ring-primary focus:outline-none cursor-pointer shadow-xs"
									>
										<option value="guest">Guest (Гость)</option>
										<option value="host">Host (Хозяин)</option>
										<option value="admin">Admin (Администратор)</option>
									</select>
								</td>

								<!-- Status Badge -->
								<td class="py-3.5 px-4">
									{#if u.status === 'active'}
										<span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[11px] font-bold bg-emerald-100 text-emerald-800 border border-emerald-200 dark:bg-emerald-950 dark:text-emerald-300 dark:border-emerald-900">
											<ShieldCheck class="h-3 w-3" /> Активен
										</span>
									{:else if u.status === 'suspended'}
										<span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[11px] font-bold bg-amber-100 text-amber-800 border border-amber-200 dark:bg-amber-950 dark:text-amber-300 dark:border-amber-900">
											<AlertTriangle class="h-3 w-3" /> Ограничен
										</span>
									{:else if u.status === 'banned'}
										<span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[11px] font-bold bg-rose-100 text-rose-800 border border-rose-200 dark:bg-rose-950 dark:text-rose-300 dark:border-rose-900">
											<UserX class="h-3 w-3" /> Заблокирован
										</span>
									{:else}
										<span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[11px] font-bold bg-slate-100 text-slate-800 border border-slate-200 dark:bg-slate-900 dark:text-slate-300 dark:border-slate-800">
											{u.status}
										</span>
									{/if}
								</td>

								<!-- Registration Date -->
								<td class="py-3.5 px-4 text-muted-foreground whitespace-nowrap">
									<div class="flex items-center gap-1.5">
										<Calendar class="h-3 w-3 text-muted-foreground/70" />
										<span>{u.created_at ? formatDate(u.created_at) : '—'}</span>
									</div>
								</td>

								<!-- Action Buttons -->
								<td class="py-3.5 px-4 text-right">
									<div class="flex items-center justify-end gap-1.5">
										{#if u.status === 'suspended' || u.status === 'banned'}
											<button
												type="button"
												onclick={() => handleQuickUnblock(u)}
												class="p-2 rounded-xl bg-emerald-100 text-emerald-800 hover:bg-emerald-200 dark:bg-emerald-950 dark:text-emerald-300 transition-colors cursor-pointer"
												title="Разблокировать пользователя"
											>
												<Unlock class="h-4 w-4" />
											</button>
										{/if}

										<button
											type="button"
											onclick={() => openBlockDialog(u)}
											class="p-2 rounded-xl bg-amber-100 text-amber-800 hover:bg-amber-200 dark:bg-amber-950 dark:text-amber-300 transition-colors cursor-pointer"
											title="Применить меры дисциплинарного воздействия"
										>
											<ShieldAlert class="h-4 w-4" />
										</button>

										<a
											href="/admin/users/{u.id}"
											class="p-2 rounded-xl border border-border hover:bg-muted text-muted-foreground hover:text-foreground transition-colors inline-flex items-center justify-center"
											title="Профиль, объявления и история взысканий"
										>
											<Eye class="h-4 w-4" />
										</a>
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
