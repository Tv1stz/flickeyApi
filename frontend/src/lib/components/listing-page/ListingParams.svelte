<!-- src/lib/components/listing-page/ListingParams.svelte -->
<script lang="ts">
	import { pluralize, PLURAL_FORMS } from '$lib/utils/pluralize';
	import type { PropertyType } from '$lib/components/card/types';

	interface Props {
		maxGuests: number;
		beds: number;
		bathrooms: number;
		bedrooms: number;
		area?: number;
		floor?: number;
		totalFloors?: number;
		propertyType?: PropertyType;
	}

	let { maxGuests, beds, bathrooms, bedrooms, area, floor, totalFloors, propertyType }: Props =
		$props();

	const isApartment = $derived(propertyType === 'apartment');

	const formatArea = (a: number): string => {
		const formatted = a % 1 === 0 ? `${a}` : a.toFixed(1).replace('.', ',');
		return `${formatted} м²`;
	};

	type ParamItem = {
		key: string;
		label: string;
		value: string;
		icon: string;
	};

	const params = $derived.by((): ParamItem[] => {
		const items: ParamItem[] = [
			{
				key: 'guests',
				label: pluralize(maxGuests, PLURAL_FORMS.guests),
				value: `${maxGuests}`,
				icon: '/icons/users.svg'
			},
			{
				key: 'bedrooms',
				label: pluralize(bedrooms, PLURAL_FORMS.rooms),
				value: `${bedrooms}`,
				icon: '/icons/door-open.svg'
			},
			{
				key: 'beds',
				label: pluralize(beds, PLURAL_FORMS.beds),
				value: `${beds}`,
				icon: '/icons/bed.svg'
			},
			{
				key: 'bathrooms',
				label: pluralize(bathrooms, PLURAL_FORMS.bathrooms),
				value: `${bathrooms}`,
				icon: '/icons/bath.svg'
			}
		];

		if (area)
			items.push({
				key: 'area',
				label: 'площадь',
				value: formatArea(area),
				icon: '/icons/maximize.svg'
			});

		if (totalFloors) {
			if (isApartment && floor) {
				items.push({
					key: 'floor',
					label: 'этаж',
					value: `${floor}/${totalFloors}`,
					icon: '/icons/building.svg'
				});
			} else if (!isApartment) {
				items.push({
					key: 'floors',
					label: pluralize(totalFloors, PLURAL_FORMS.floors),
					value: `${totalFloors}`,
					icon: '/icons/building.svg'
				});
			}
		}

		return items;
	});
</script>

<div>
	<div class="flex flex-wrap items-center gap-x-10 gap-y-4 text-[16px] text-zinc-900">
		{#each params as p (p.key)}
			<div class="flex items-center gap-2.5">
				<img src={p.icon} alt={p.label} width={26} height={26} />
				<span class="font-medium">{p.value} {p.label}</span>
			</div>
		{/each}
	</div>
</div>
