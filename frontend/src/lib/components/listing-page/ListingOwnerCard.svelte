<!-- src/lib/components/listing-page/ListingOwnerCard.svelte -->
<script lang="ts">
	import { CheckCircle, ChevronRight, Calendar } from 'lucide-svelte';
	import { resolve } from '$app/paths';
	import type { Owner } from '$lib/components/card/types';
	import { pluralRu } from '$lib/utils/format';

	interface Props {
		owner: Owner;
		compact?: boolean;
	}

	let { owner, compact = false }: Props = $props();

	const initials = $derived(
		owner.name
			.split(' ')
			.map((word) => word[0])
			.join('')
			.toUpperCase()
			.slice(0, 2)
	);

	const memberSince = $derived(new Date(owner.createdAt).getFullYear());
	const yearsOnPlatform = $derived(new Date().getFullYear() - memberSince);

	const listingsText = $derived(
		`${owner.listingsCount} ${pluralRu(owner.listingsCount, ['объявление', 'объявления', 'объявлений'])}`
	);

	const yearsText = $derived(
		yearsOnPlatform > 0
			? `${yearsOnPlatform} ${pluralRu(yearsOnPlatform, ['год', 'года', 'лет'])} на платформе`
			: 'Недавно на платформе'
	);
</script>

{#if compact}
	<a
		href={resolve('/users/[id]', { id: owner.id })}
		class="group block overflow-hidden rounded-3xl bg-white
               shadow-md ring-1 ring-zinc-200/60
               transition-all duration-300
               hover:-translate-y-0.5 hover:shadow-xl"
	>
		<!-- Аватар + имя -->
		<div
			class="flex flex-col items-center gap-3 bg-gradient-to-b from-white to-zinc-50/50 px-6 pt-7 pb-6 text-center"
		>
			{#if owner.avatar}
				<img
					src={owner.avatar}
					alt={owner.name}
					class="h-20 w-20 rounded-full object-cover shadow-sm ring-4 ring-white"
				/>
			{:else}
				<div
					class="flex h-20 w-20 items-center justify-center rounded-full
                            bg-gradient-to-br from-zinc-700 to-zinc-900
                            text-[28px] font-semibold text-white shadow-sm ring-4 ring-white"
				>
					{initials}
				</div>
			{/if}

			<div>
				<h3 class="text-[20px] leading-tight font-bold tracking-tight text-zinc-900">
					{owner.name}
				</h3>
				{#if owner.isVerified}
					<div
						class="mt-1 flex items-center justify-center gap-1.5
                                text-[12px] font-medium text-emerald-600"
					>
						<CheckCircle size={12} class="shrink-0" />
						Подтверждённый профиль
					</div>
				{:else}
					<div class="mt-1 text-[12px] text-zinc-400">Профиль не подтверждён</div>
				{/if}

				<div class="mt-1.5 flex items-center justify-center gap-1.5 text-[12px] text-zinc-400">
					<Calendar size={12} strokeWidth={1.5} class="shrink-0" />
					На платформе с {memberSince} года
				</div>
			</div>
		</div>

		<!-- Статистика -->
		<div class="grid grid-cols-2 divide-x divide-zinc-100 border-t border-zinc-100">
			<div class="flex flex-col items-center gap-0.5 py-4">
				<span class="text-[26px] leading-none font-extrabold tracking-tight text-zinc-900">
					{owner.listingsCount}
				</span>
				<span class="text-[11px] font-medium text-zinc-400">
					{pluralRu(owner.listingsCount, ['объявление', 'объявления', 'объявлений'])}
				</span>
			</div>

			<div class="flex flex-col items-center gap-0.5 py-4">
				{#if yearsOnPlatform > 0}
					<span class="text-[26px] leading-none font-extrabold tracking-tight text-zinc-900">
						{yearsOnPlatform}
					</span>
					<span class="text-[11px] font-medium text-zinc-400">
						{pluralRu(yearsOnPlatform, ['год', 'года', 'лет'])} на платформе
					</span>
				{:else}
					<span class="px-2 text-center text-[13px] leading-snug font-medium text-zinc-400">
						Недавно на платформе
					</span>
				{/if}
			</div>
		</div>

		<!-- Футер -->
		<div
			class="flex items-center justify-between border-t border-zinc-100
                    bg-zinc-50/50 px-6 py-3.5
                    transition-colors duration-200 group-hover:bg-zinc-50"
		>
			<span
				class="text-[13px] font-semibold text-zinc-900
                         underline-offset-4 group-hover:underline"
			>
				Посмотреть профиль
			</span>
			<ChevronRight
				size={16}
				class="shrink-0 text-zinc-300 transition-all duration-300
                       group-hover:translate-x-1 group-hover:text-zinc-500"
			/>
		</div>
	</a>
{:else}
	<!-- Full version for mobile -->
	<div>
		<h2 class="mb-6 text-[22px] font-semibold tracking-tight text-zinc-900 sm:text-2xl">
			Хозяин объекта
		</h2>

		<a
			href={resolve('/users/[id]', { id: owner.id })}
			class="group block w-full max-w-2xl overflow-hidden rounded-4xl bg-white shadow-lg ring-1 ring-zinc-200/60 transition-all delay-75 duration-300 hover:-translate-y-1 hover:shadow-2xl"
		>
			<div class="flex flex-col sm:flex-row">
				<!-- Left: Avatar & Name -->
				<div
					class="relative flex flex-1 flex-col items-center justify-center gap-y-3 border-b border-zinc-100 bg-gradient-to-b from-white to-zinc-50/50 p-8 text-center sm:border-r sm:border-b-0 sm:p-10"
				>
					{#if owner.avatar}
						<img
							src={owner.avatar}
							alt={owner.name}
							class="h-[110px] w-[110px] rounded-full object-cover shadow-sm ring-4 ring-white"
						/>
					{:else}
						<div
							class="flex h-[110px] w-[110px] items-center justify-center rounded-full bg-gradient-to-br from-zinc-700 to-zinc-900 text-[38px] font-semibold tracking-wide text-white shadow-sm ring-4 ring-white"
						>
							{initials}
						</div>
					{/if}

					<h3 class="text-[26px] leading-tight font-bold tracking-tight text-zinc-900">
						{owner.name}
					</h3>

					{#if owner.isVerified}
						<div class="flex items-center gap-1.5 text-[13px] font-medium text-emerald-600">
							<CheckCircle size={13} class="shrink-0" />
							Подтверждённый профиль
						</div>
					{:else}
						<div class="text-[13px] text-zinc-400">Профиль не подтверждён</div>
					{/if}

					<div class="flex items-center gap-1.5 text-[13px] text-zinc-400">
						<Calendar size={13} strokeWidth={1.5} class="shrink-0" />
						На платформе с {memberSince} года
					</div>
				</div>

				<!-- Right: Stats -->
				<div class="flex flex-1 flex-col justify-center gap-y-5 px-8 py-8 sm:px-10">
					<div class="flex flex-col">
						<div class="flex items-end gap-2">
							<span class="text-[34px] leading-none font-extrabold tracking-tight text-zinc-900">
								{owner.listingsCount}
							</span>
							<span class="mb-0.5 text-[15px] leading-snug font-medium text-zinc-500">
								{listingsText.replace(owner.listingsCount.toString(), '').trim()}
							</span>
						</div>
					</div>

					<div class="w-full border-t border-zinc-100"></div>

					<div class="flex flex-col">
						{#if yearsOnPlatform > 0}
							<div class="flex items-end gap-2">
								<span class="text-[34px] leading-none font-extrabold tracking-tight text-zinc-900">
									{yearsOnPlatform}
								</span>
								<span class="mb-0.5 text-[15px] leading-snug font-medium text-zinc-500">
									{yearsText.replace(yearsOnPlatform.toString(), '').trim()}
								</span>
							</div>
						{:else}
							<div class="flex items-center gap-2">
								<Calendar size={28} strokeWidth={1.5} class="shrink-0 text-zinc-400" />
								<span class="text-[15px] font-medium text-zinc-500">
									{yearsText}
								</span>
							</div>
						{/if}
					</div>

					<div class="w-full border-t border-zinc-100"></div>

					<p class="text-[13px] leading-relaxed text-zinc-400">
						Свяжитесь с хозяином напрямую через контакты в объявлении
					</p>
				</div>
			</div>

			<div
				class="flex items-center justify-between border-t border-zinc-100 bg-zinc-50/50 px-8 py-4 transition-colors duration-200 group-hover:bg-zinc-50"
			>
				<span
					class="text-[14px] font-semibold text-zinc-900 underline-offset-4 group-hover:underline"
				>
					Посмотреть профиль хозяина
				</span>
				<ChevronRight
					size={20}
					class="shrink-0 text-zinc-300 transition-all duration-300 group-hover:translate-x-1 group-hover:text-zinc-500"
				/>
			</div>
		</a>
	</div>
{/if}
