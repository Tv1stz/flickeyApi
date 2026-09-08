<!-- src/lib/components/listing-page/ContactsModal.svelte -->
<script lang="ts">
	import { Phone, Send, MessageCircle, ChevronRight, Mail } from 'lucide-svelte';
	import type { OwnerContacts } from '$lib/components/card/types';
	import Modal from '$lib/components/ui/Modal.svelte';

	interface Props {
		open: boolean;
		onClose: () => void;
		ownerContacts?: OwnerContacts;
		ownerName?: string;
	}

	let { open, onClose, ownerContacts, ownerName }: Props = $props();

	function formatPhone(phone: string): string {
		const cleaned = phone.replace(/\D/g, '');
		if (cleaned.length === 12) {
			return `+${cleaned.slice(0, 3)} (${cleaned.slice(3, 5)}) ${cleaned.slice(5, 8)}-${cleaned.slice(8, 10)}-${cleaned.slice(10)}`;
		}
		return phone;
	}

	function getPhoneLink(phone: string): string {
		return `tel:+${phone.replace(/\D/g, '')}`;
	}
	function getEmailLink(email: string): string {
		return `mailto:${email.trim()}`;
	}
	function getTelegramLink(username: string): string {
		return `https://t.me/${username.replace('@', '')}`;
	}
	function getViberLink(phone: string): string {
		return `viber://chat?number=%2B${phone.replace(/\D/g, '')}`;
	}
	function getWhatsAppLink(phone: string): string {
		return `https://wa.me/${phone.replace(/\D/g, '')}`;
	}
	function getSignalLink(phone: string): string {
		return `https://signal.me/#p/+${phone.replace(/\D/g, '')}`;
	}

	function handleContactClick(method: { href: string; newTab?: boolean }) {
		if (typeof window === 'undefined') return;
		if (method.newTab) {
			window.open(method.href, '_blank', 'noopener');
			return;
		}
		window.location.href = method.href;
	}

	const contactMethods = $derived.by(() => {
		if (!ownerContacts) return [];
		const methods: Array<{
			type: string;
			label: string;
			sublabel: string;
			href: string;
			icon: typeof Phone;
			newTab?: boolean;
		}> = [];
		if (ownerContacts.phone)
			methods.push({
				type: 'phone',
				label: formatPhone(ownerContacts.phone),
				sublabel: 'Позвонить',
				href: getPhoneLink(ownerContacts.phone),
				icon: Phone
			});
		if (ownerContacts.email)
			methods.push({
				type: 'email',
				label: ownerContacts.email,
				sublabel: 'Email',
				href: getEmailLink(ownerContacts.email),
				icon: Mail
			});
		if (ownerContacts.telegram)
			methods.push({
				type: 'telegram',
				label: '@' + ownerContacts.telegram.replace('@', ''),
				sublabel: 'Telegram',
				href: getTelegramLink(ownerContacts.telegram),
				icon: Send,
				newTab: true
			});
		if (ownerContacts.viber)
			methods.push({
				type: 'viber',
				label: formatPhone(ownerContacts.viber),
				sublabel: 'Viber',
				href: getViberLink(ownerContacts.viber),
				icon: MessageCircle
			});
		if (ownerContacts.whatsapp)
			methods.push({
				type: 'whatsapp',
				label: formatPhone(ownerContacts.whatsapp),
				sublabel: 'WhatsApp',
				href: getWhatsAppLink(ownerContacts.whatsapp),
				icon: MessageCircle,
				newTab: true
			});
		if (ownerContacts.signal)
			methods.push({
				type: 'signal',
				label: formatPhone(ownerContacts.signal),
				sublabel: 'Signal',
				href: getSignalLink(ownerContacts.signal),
				icon: MessageCircle,
				newTab: true
			});
		return methods;
	});
</script>

<Modal {open} {onClose} title={ownerName || 'Контакты хозяина'}>
	<div class="modal-body pt-5 sm:pt-6">
		{#if contactMethods.length === 0}
			<p
				class="rounded-2xl border border-dashed border-zinc-200 px-4 py-5 text-center text-sm text-zinc-500"
			>
				Контакты пока не добавлены
			</p>
		{:else}
			<div class="space-y-1.5">
				{#each contactMethods as method (method.type)}
					{@const ContactIcon = method.icon}
					<button
						type="button"
						onclick={() => handleContactClick(method)}
						class="flex min-h-14 w-full touch-manipulation items-center gap-4 rounded-2xl p-4
                           text-left transition-colors hover:bg-zinc-50 active:bg-zinc-100"
					>
						<div
							class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-zinc-100"
						>
							<ContactIcon size={20} class="text-zinc-600" />
						</div>
						<div class="min-w-0 flex-1">
							<div class="font-semibold text-zinc-900">{method.label}</div>
							<div class="text-sm text-zinc-500">{method.sublabel}</div>
						</div>
						<ChevronRight size={18} class="shrink-0 text-zinc-300" />
					</button>
				{/each}
			</div>
		{/if}
	</div>

	{#snippet footer()}
		<div class="modal-footer">
			<p class="text-center text-xs text-zinc-400">Представьтесь и укажите желаемые даты</p>
		</div>
	{/snippet}
</Modal>
