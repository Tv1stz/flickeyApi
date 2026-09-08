<!-- src/lib/components/host-listing-edit/sections/EditContacts.svelte -->
<script lang="ts">
	import TextInputField from '$lib/components/form/TextInputField.svelte';
	import { cardStyles, formStyles } from '$lib/config/styles';
	import { Phone, Mail, MessageSquare, Globe, Info } from 'lucide-svelte';
	import type { ListingEditFormData } from '../types';

	interface Props {
		form: Partial<ListingEditFormData>;
		onUpdate: (updates: Partial<ListingEditFormData>) => void;
	}

	let { form, onUpdate }: Props = $props();

	function updateContacts(field: string, value: string) {
		const current = form.contacts || {
			phone: '',
			email: '',
			telegram: '',
			whatsapp: ''
		};
		onUpdate({ contacts: { ...current, [field]: value } });
	}

	const contactFields = [
		{ id: 'phone', label: 'Номер телефона', placeholder: '+375 29 123-45-67', icon: Phone, required: true },
		{ id: 'email', label: 'Электронная почта', placeholder: 'example@mail.ru', icon: Mail, required: false },
		{ id: 'telegram', label: 'Telegram (username)', placeholder: '@username', icon: MessageSquare, required: false },
		{ id: 'whatsapp', label: 'WhatsApp', placeholder: '+375 29 123-45-67', icon: Globe, required: false }
	];
</script>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Контактная информация</h2>
	<p class={cardStyles.subtitle}>Укажите способы связи для гостей после бронирования</p>

	<div class={cardStyles.contentGap}>
		<div class={formStyles.fieldGroup}>
			{#each contactFields as field}
				<div>
					<label for={`contact-${field.id}`} class="mb-1.5 flex items-center gap-2 text-sm font-medium text-zinc-900">
						<field.icon class="h-4 w-4 text-zinc-500" />
						{field.label}
						{#if field.required}
							<span class="text-red-500">*</span>
						{/if}
					</label>
					<TextInputField
						id={`contact-${field.id}`}
						label={field.label}
						placeholder={field.placeholder}
						required={field.required}
						value={form.contacts?.[field.id as keyof NonNullable<typeof form.contacts>] || ''}
						onInput={(v) => updateContacts(field.id, v)}
						onBlur={() => {}}
					/>
				</div>
			{/each}
		</div>

		<div class="flex items-start gap-3.5 rounded-2xl border border-blue-100 bg-blue-50/70 p-4">
			<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-blue-100 text-blue-700">
				<Info class="h-4 w-4" />
			</div>
			<div class="text-xs text-blue-900">
				<p class="font-semibold">Безопасность контактов</p>
				<p class="mt-0.5 leading-relaxed text-blue-800">
					Контактная информация отображается гостю только после подтверждения и оплаты бронирования. 
					Рекомендуем указать актуальный номер телефона и Telegram для оперативного ответа.
				</p>
			</div>
		</div>
	</div>
</section>
