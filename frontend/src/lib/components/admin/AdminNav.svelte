<script lang="ts">
	import { page } from '$app/state';
	import { LayoutDashboard, Building2, ShieldCheck, Users, AlertTriangle, FileText } from 'lucide-svelte';

	const navItems = [
		{ href: '/admin', label: 'Обзор', icon: LayoutDashboard, exact: true },
		{ href: '/admin/listings', label: 'Модерация жилья', icon: Building2, exact: false },
		{ href: '/admin/verification', label: 'Верификация', icon: ShieldCheck, exact: false },
		{ href: '/admin/users', label: 'Пользователи', icon: Users, exact: false },
		{ href: '/admin/reports', label: 'Жалобы', icon: AlertTriangle, exact: false },
		{ href: '/admin/audit', label: 'Аудит действий', icon: FileText, exact: false }
	];

	function isActive(href: string, exact: boolean): boolean {
		const path = page.url.pathname;
		if (exact) {
			return path === href;
		}
		return path.startsWith(href);
	}
</script>

<nav class="bg-card border-b border-border shadow-xs">
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="flex items-center justify-between h-16">
			<div class="flex items-center gap-6">
				<div class="flex items-center gap-2">
					<div class="w-8 h-8 rounded-lg bg-emerald-600 text-white flex items-center justify-center font-bold text-lg">
						F
					</div>
					<span class="font-bold text-lg text-foreground tracking-tight">Flickey Admin</span>
					<span class="text-[10px] font-semibold uppercase bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300 px-2 py-0.5 rounded-full">
						Compliance
					</span>
				</div>

				<div class="hidden md:flex items-center space-x-1">
					{#each navItems as item}
						{@const active = isActive(item.href, item.exact)}
						{@const Icon = item.icon}
						<a
							href={item.href}
							class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors {active
								? 'bg-primary text-primary-foreground font-semibold shadow-xs'
								: 'text-muted-foreground hover:text-foreground hover:bg-muted/50'}"
						>
							<Icon class="h-4 w-4 shrink-0" />
							<span>{item.label}</span>
						</a>
					{/each}
				</div>
			</div>
		</div>

		<!-- Mobile navigation sub-bar -->
		<div class="md:hidden flex items-center space-x-1 overflow-x-auto pb-3 pt-1 scrollbar-none">
			{#each navItems as item}
				{@const active = isActive(item.href, item.exact)}
				{@const Icon = item.icon}
				<a
					href={item.href}
					class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors {active
						? 'bg-primary text-primary-foreground font-semibold'
						: 'text-muted-foreground hover:text-foreground hover:bg-muted/50'}"
				>
					<Icon class="h-3.5 w-3.5 shrink-0" />
					<span>{item.label}</span>
				</a>
			{/each}
		</div>
	</div>
</nav>
