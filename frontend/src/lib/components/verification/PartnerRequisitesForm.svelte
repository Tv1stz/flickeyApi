<!-- src/lib/components/verification/PartnerRequisitesForm.svelte -->
<script lang="ts">
	import {
		Building2,
		User,
		FileText,
		UploadCloud,
		CheckCircle2,
		AlertCircle,
		Loader2,
		X,
		ShieldCheck,
		ExternalLink
	} from 'lucide-svelte';
	import type {
		ProviderType,
		SubmitVerificationPayload,
		VerificationResponse
	} from '$lib/types/verification';
	import {
		UNP_COMPANY_REGEX,
		UNP_INDIVIDUAL_REGEX,
		IBAN_BELARUS_REGEX,
		BIC_REGEX,
		PERSONAL_ID_REGEX
	} from '$lib/types/verification';
	import { mediaApi } from '$lib/api/media';
	import { verificationApi } from '$lib/api/verification';
	import { toast } from '$lib/stores/toastStore';

	interface Props {
		initialData?: VerificationResponse | null;
		onSuccess?: (res: VerificationResponse) => void;
	}

	let { initialData = null, onSuccess }: Props = $props();

	// Tab selection
	type FormTab = 'company' | 'individual';
	let activeTab = $state<FormTab>('company');

	// Individual fields
	let indFirstName = $state('');
	let indLastName = $state('');
	let indEmail = $state('');
	let indPhone = $state('+375');
	let indPersonalId = $state('');
	let indBirthDate = $state('');
	let indBankName = $state('');
	let indUNP = $state('');
	let indBIC = $state('');
	let indIBAN = $state('BY');

	// Company fields
	let compLegalName = $state('');
	let compBrandName = $state('');
	let compWebsite = $state('');
	let compEmail = $state('');
	let compPhone = $state('+375');
	let compPostalCode = $state('');
	let compCity = $state('');
	let compStreetBuilding = $state('');
	let compActualSame = $state(true);
	let compActualPostalCode = $state('');
	let compActualCity = $state('');
	let compActualStreetBuilding = $state('');
	let compBankName = $state('');
	let compUNP = $state('');
	let compBIC = $state('');
	let compEGR = $state('');
	let compIBAN = $state('BY');
	let compContactName = $state('');
	let compContactEmail = $state('');
	let compContactPosition = $state('');
	let compContactPhone = $state('+375');

	// Attached documents
	interface AttachedDoc {
		id: string;
		name: string;
		mediaId: string;
		fileKey: string;
		url: string;
	}
	let attachedDocs = $state<AttachedDoc[]>([]);
	let isUploadingDoc = $state(false);

	// Consent checkbox (Закон РБ № 99-З)
	let consentGiven = $state(false);

	// Submission state
	let isSubmitting = $state(false);
	let formErrors = $state<Record<string, string>>({});

	// Initialize from existing data if present
	$effect(() => {
		if (initialData && initialData.status !== 'none') {
			activeTab = (initialData.provider_type === 'individual' || initialData.provider_type === 'self_employed')
				? 'individual'
				: 'company';
			const req = initialData.requisites || {};

			if (initialData.provider_type === 'individual' || initialData.provider_type === 'self_employed') {
				indFirstName = req.first_name || '';
				indLastName = req.last_name || '';
				indEmail = req.email || '';
				indPhone = req.phone || '+375';
				indPersonalId = req.personal_id || '';
				indBirthDate = req.birth_date || '';
				indBankName = req.bank_name || '';
				indUNP = initialData.unp || '';
				indBIC = req.bic || '';
				indIBAN = req.iban || 'BY';
			} else {
				compLegalName = initialData.legal_name || '';
				compBrandName = req.brand_name || '';
				compWebsite = req.website || '';
				compEmail = req.email || '';
				compPhone = req.phone || '+375';
				compPostalCode = req.legal_address?.postal_code || '';
				compCity = req.legal_address?.city || '';
				compStreetBuilding = req.legal_address?.street_building || '';
				compActualSame = req.actual_address_same !== false;
				compActualPostalCode = req.actual_address?.postal_code || '';
				compActualCity = req.actual_address?.city || '';
				compActualStreetBuilding = req.actual_address?.street_building || '';
				compBankName = req.bank_name || '';
				compUNP = initialData.unp || '';
				compBIC = req.bic || '';
				compEGR = req.egr_number || '';
				compIBAN = req.iban || 'BY';
				compContactName = req.contact_person?.full_name || '';
				compContactEmail = req.contact_person?.email || '';
				compContactPosition = req.contact_person?.position || '';
				compContactPhone = req.contact_person?.phone || '+375';
			}
			consentGiven = true;
		}
	});

	// Document upload handler
	async function handleDocUpload(files: FileList | null) {
		if (!files || files.length === 0 || isUploadingDoc) return;
		const file = files[0];

		if (file.size > 25 * 1024 * 1024) {
			toast.error('Файл слишком большой', 'Максимальный размер документа 25 МБ.');
			return;
		}

		isUploadingDoc = true;
		try {
			const res = await mediaApi.uploadFile(file);
			attachedDocs = [
				...attachedDocs,
				{
					id: crypto.randomUUID(),
					name: file.name,
					mediaId: res.mediaId,
					fileKey: res.fileKey,
					url: mediaApi.getMediaUrl(res.fileKey)
				}
			];
			toast.success('Документ прикреплен', file.name);
		} catch (err: any) {
			console.error('Ошибка загрузки документа:', err);
			toast.error('Ошибка загрузки документа', err?.message);
		} finally {
			isUploadingDoc = false;
		}
	}

	function removeDoc(id: string) {
		attachedDocs = attachedDocs.filter((d) => d.id !== id);
	}

	// Validation
	function validateForm(): boolean {
		const errors: Record<string, string> = {};

		if (activeTab === 'individual') {
			if (!indFirstName.trim()) errors.indFirstName = 'Укажите имя';
			if (!indLastName.trim()) errors.indLastName = 'Укажите фамилию';
			if (!indEmail.trim() || !indEmail.includes('@')) errors.indEmail = 'Укажите корректный email';
			if (!indPhone.trim() || indPhone.length < 9) errors.indPhone = 'Укажите номер телефона';

			const cleanPID = indPersonalId.trim().toUpperCase();
			if (!cleanPID || !PERSONAL_ID_REGEX.test(cleanPID)) {
				errors.indPersonalId = 'Идентификационный номер должен содержать 14 символов (буквы и цифры)';
			}
			if (!indBirthDate.trim()) errors.indBirthDate = 'Укажите дату рождения';
			if (!indBankName.trim()) errors.indBankName = 'Укажите наименование банка';

			const cleanUNP = indUNP.trim().toUpperCase();
			if (!cleanUNP || !UNP_INDIVIDUAL_REGEX.test(cleanUNP)) {
				errors.indUNP = 'УНП физлица / самозанятого должен содержать ровно 9 буквенно-цифровых символов';
			}

			const cleanBIC = indBIC.trim().toUpperCase();
			if (!cleanBIC || !BIC_REGEX.test(cleanBIC)) {
				errors.indBIC = 'БИК должен состоять из 8 или 11 символов';
			}

			const cleanIBAN = indIBAN.trim().toUpperCase().replace(/\s+/g, '');
			if (!cleanIBAN || !IBAN_BELARUS_REGEX.test(cleanIBAN)) {
				errors.indIBAN = 'IBAN в Беларуси должен начинаться с BY и содержать 28 знаков';
			}
		} else {
			if (!compLegalName.trim()) errors.compLegalName = 'Укажите юридическое наименование';
			if (!compEmail.trim() || !compEmail.includes('@')) errors.compEmail = 'Укажите корректный email';
			if (!compPhone.trim() || compPhone.length < 9) errors.compPhone = 'Укажите телефон организации';
			if (!compPostalCode.trim()) errors.compPostalCode = 'Укажите почтовый индекс';
			if (!compCity.trim()) errors.compCity = 'Укажите город';
			if (!compStreetBuilding.trim()) errors.compStreetBuilding = 'Укажите улицу и дом';

			if (!compActualSame) {
				if (!compActualPostalCode.trim()) errors.compActualPostalCode = 'Укажите почтовый индекс';
				if (!compActualCity.trim()) errors.compActualCity = 'Укажите город';
				if (!compActualStreetBuilding.trim()) errors.compActualStreetBuilding = 'Укажите улицу и дом';
			}

			if (!compBankName.trim()) errors.compBankName = 'Укажите наименование банка';

			const cleanUNP = compUNP.trim();
			if (!cleanUNP || !UNP_COMPANY_REGEX.test(cleanUNP)) {
				errors.compUNP = 'УНП организации / ИП должен состоять строго из 9 цифр';
			}

			const cleanBIC = compBIC.trim().toUpperCase();
			if (!cleanBIC || !BIC_REGEX.test(cleanBIC)) {
				errors.compBIC = 'БИК должен состоять из 8 или 11 символов';
			}

			if (!compEGR.trim()) errors.compEGR = 'Укажите номер регистрации в ЕГР';

			const cleanIBAN = compIBAN.trim().toUpperCase().replace(/\s+/g, '');
			if (!cleanIBAN || !IBAN_BELARUS_REGEX.test(cleanIBAN)) {
				errors.compIBAN = 'IBAN в Беларуси должен начинаться с BY и содержать 28 знаков';
			}

			// Contact person
			if (!compContactName.trim()) errors.compContactName = 'Укажите ФИО контактного лица';
			if (!compContactEmail.trim() || !compContactEmail.includes('@')) errors.compContactEmail = 'Укажите email контактного лица';
			if (!compContactPhone.trim() || compContactPhone.length < 9) errors.compContactPhone = 'Укажите телефон контактного лица';
		}

		if (attachedDocs.length === 0) {
			errors.documents = 'Прикрепите хотя бы один подтверждающий документ';
		}

		if (!consentGiven) {
			errors.consent = 'Необходимо дать согласие на обработку персональных данных (Закон РБ № 99-З)';
		}

		formErrors = errors;
		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validateForm()) {
			toast.error('Проверьте поля формы', 'Заполните все обязательные поля корректно.');
			return;
		}

		isSubmitting = true;
		try {
			let payload: SubmitVerificationPayload;

			if (activeTab === 'individual') {
				const cleanIBAN = indIBAN.trim().toUpperCase().replace(/\s+/g, '');
				payload = {
					provider_type: 'individual',
					legal_name: `${indLastName.trim()} ${indFirstName.trim()}`,
					unp: indUNP.trim().toUpperCase(),
					requisites: {
						first_name: indFirstName.trim(),
						last_name: indLastName.trim(),
						email: indEmail.trim(),
						phone: indPhone.trim(),
						personal_id: indPersonalId.trim().toUpperCase(),
						birth_date: indBirthDate.trim(),
						bank_name: indBankName.trim(),
						bic: indBIC.trim().toUpperCase(),
						iban: cleanIBAN,
						currency: 'BYN'
					},
					documents: attachedDocs.map((d) => d.mediaId)
				};
			} else {
				const cleanIBAN = compIBAN.trim().toUpperCase().replace(/\s+/g, '');
				payload = {
					provider_type: 'legal_entity',
					legal_name: compLegalName.trim(),
					unp: compUNP.trim(),
					requisites: {
						brand_name: compBrandName.trim(),
						website: compWebsite.trim(),
						email: compEmail.trim(),
						phone: compPhone.trim(),
						legal_address: {
							postal_code: compPostalCode.trim(),
							city: compCity.trim(),
							street_building: compStreetBuilding.trim()
						},
						actual_address_same: compActualSame,
						actual_address: compActualSame
							? undefined
							: {
									postal_code: compActualPostalCode.trim(),
									city: compActualCity.trim(),
									street_building: compActualStreetBuilding.trim()
								},
						bank_name: compBankName.trim(),
						bic: compBIC.trim().toUpperCase(),
						egr_number: compEGR.trim(),
						iban: cleanIBAN,
						currency: 'BYN',
						contact_person: {
							full_name: compContactName.trim(),
							email: compContactEmail.trim(),
							position: compContactPosition.trim(),
							phone: compContactPhone.trim()
						}
					},
					documents: attachedDocs.map((d) => d.mediaId)
				};
			}

			const res = await verificationApi.submitVerification(payload);
			toast.success(
				'Документы отправлены',
				'Реквизиты переданы на проверку администрации. Статус обновится автоматически.'
			);
			onSuccess?.(res);
		} catch (err: any) {
			console.error('Ошибка отправки реквизитов:', err);
			toast.error('Ошибка отправки', err?.message || 'Не удалось отправить реквизиты.');
		} finally {
			isSubmitting = false;
		}
	}
</script>

<div class="space-y-6">
	<!-- Tab Switcher -->
	<div class="flex items-center gap-2 rounded-2xl bg-zinc-100 p-1.5 dark:bg-muted">
		<button
			type="button"
			class="flex flex-1 items-center justify-center gap-2 rounded-xl py-2.5 px-4 text-xs font-bold transition-all {activeTab === 'company'
				? 'bg-white text-zinc-900 shadow-xs dark:bg-card dark:text-foreground'
				: 'text-zinc-500 hover:text-zinc-900 dark:text-muted-foreground dark:hover:text-foreground'}"
			onclick={() => (activeTab = 'company')}
		>
			<Building2 class="h-4 w-4" />
			<span>Компания (ИП, ООО, юрлицо)</span>
		</button>
		<button
			type="button"
			class="flex flex-1 items-center justify-center gap-2 rounded-xl py-2.5 px-4 text-xs font-bold transition-all {activeTab === 'individual'
				? 'bg-white text-zinc-900 shadow-xs dark:bg-card dark:text-foreground'
				: 'text-zinc-500 hover:text-zinc-900 dark:text-muted-foreground dark:hover:text-foreground'}"
			onclick={() => (activeTab = 'individual')}
		>
			<User class="h-4 w-4" />
			<span>Физическое лицо (Самозанятый / НПД)</span>
		</button>
	</div>

	<!-- Form Fields Section -->
	<div class="space-y-5">
		{#if activeTab === 'individual'}
			<!-- INDIVIDUAL / SELF-EMPLOYED -->
			<div class="rounded-2xl border border-zinc-100 bg-zinc-50/60 p-4 dark:border-border/60 dark:bg-muted/20 text-xs text-zinc-600 dark:text-muted-foreground">
				<p class="font-semibold text-zinc-900 dark:text-foreground mb-0.5">Физическое лицо и плательщики налога на проф. доход (НПД)</p>
				<p>Укажите личные паспортные данные и банковские реквизиты счета в белорусских рублях (BYN).</p>
			</div>

			<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
				<div>
					<label for="indLastName" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Фамилия *</label>
					<input
						id="indLastName"
						type="text"
						bind:value={indLastName}
						placeholder="Иванов"
						class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.indLastName ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
					/>
					{#if formErrors.indLastName}<p class="mt-1 text-[11px] text-rose-600">{formErrors.indLastName}</p>{/if}
				</div>

				<div>
					<label for="indFirstName" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Имя *</label>
					<input
						id="indFirstName"
						type="text"
						bind:value={indFirstName}
						placeholder="Иван"
						class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.indFirstName ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
					/>
					{#if formErrors.indFirstName}<p class="mt-1 text-[11px] text-rose-600">{formErrors.indFirstName}</p>{/if}
				</div>
			</div>

			<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
				<div>
					<label for="indEmail" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Email для отчетов *</label>
					<input
						id="indEmail"
						type="email"
						bind:value={indEmail}
						placeholder="partner@example.by"
						class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.indEmail ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
					/>
					{#if formErrors.indEmail}<p class="mt-1 text-[11px] text-rose-600">{formErrors.indEmail}</p>{/if}
				</div>

				<div>
					<label for="indPhone" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Номер телефона *</label>
					<input
						id="indPhone"
						type="tel"
						bind:value={indPhone}
						placeholder="+375 29 123-45-67"
						class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.indPhone ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
					/>
					{#if formErrors.indPhone}<p class="mt-1 text-[11px] text-rose-600">{formErrors.indPhone}</p>{/if}
				</div>
			</div>

			<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
				<div>
					<label for="indPersonalId" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Идентификационный (личный) номер *</label>
					<input
						id="indPersonalId"
						type="text"
						maxlength="14"
						bind:value={indPersonalId}
						placeholder="1234567A001PB1"
						class="w-full font-mono uppercase rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.indPersonalId ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
					/>
					<p class="mt-0.5 text-[10px] text-zinc-400">14 знаков из паспорта / вида на жительство</p>
					{#if formErrors.indPersonalId}<p class="mt-1 text-[11px] text-rose-600">{formErrors.indPersonalId}</p>{/if}
				</div>

				<div>
					<label for="indBirthDate" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Дата рождения *</label>
					<input
						id="indBirthDate"
						type="date"
						bind:value={indBirthDate}
						class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.indBirthDate ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
					/>
					{#if formErrors.indBirthDate}<p class="mt-1 text-[11px] text-rose-600">{formErrors.indBirthDate}</p>{/if}
				</div>
			</div>

			<!-- Bank Requisites -->
			<div class="border-t border-zinc-100 pt-4 dark:border-border/60 space-y-4">
				<h4 class="text-xs font-bold uppercase tracking-wider text-zinc-400 dark:text-zinc-500">Банковские реквизиты (Беларусь)</h4>

				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<div>
						<label for="indBankName" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Наименование банка *</label>
						<input
							id="indBankName"
							type="text"
							bind:value={indBankName}
							placeholder="ОАО «АСБ Беларусбанк», Альфа-Банк..."
							class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.indBankName ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
						/>
						{#if formErrors.indBankName}<p class="mt-1 text-[11px] text-rose-600">{formErrors.indBankName}</p>{/if}
					</div>

					<div>
						<label for="indUNP" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">УНП плательщика *</label>
						<input
							id="indUNP"
							type="text"
							maxlength="9"
							bind:value={indUNP}
							placeholder="9 символов"
							class="w-full font-mono uppercase rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.indUNP ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
						/>
						{#if formErrors.indUNP}<p class="mt-1 text-[11px] text-rose-600">{formErrors.indUNP}</p>{/if}
					</div>
				</div>

				<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
					<div>
						<label for="indBIC" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">БИК банка (BIC) *</label>
						<input
							id="indBIC"
							type="text"
							maxlength="11"
							bind:value={indBIC}
							placeholder="AKBBBY2X"
							class="w-full font-mono uppercase rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.indBIC ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
						/>
						{#if formErrors.indBIC}<p class="mt-1 text-[11px] text-rose-600">{formErrors.indBIC}</p>{/if}
					</div>

					<div class="sm:col-span-2">
						<label for="indIBAN" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Номер счета (IBAN) *</label>
						<input
							id="indIBAN"
							type="text"
							maxlength="34"
							bind:value={indIBAN}
							placeholder="BY00AKBB00000000000000000000"
							class="w-full font-mono uppercase rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.indIBAN ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
						/>
						<p class="mt-0.5 text-[10px] text-zinc-400">Формат BY + 26 символов, валюта: BYN</p>
						{#if formErrors.indIBAN}<p class="mt-1 text-[11px] text-rose-600">{formErrors.indIBAN}</p>{/if}
					</div>
				</div>
			</div>

		{:else}
			<!-- COMPANY / IP -->
			<div class="rounded-2xl border border-zinc-100 bg-zinc-50/60 p-4 dark:border-border/60 dark:bg-muted/20 text-xs text-zinc-600 dark:text-muted-foreground">
				<p class="font-semibold text-zinc-900 dark:text-foreground mb-0.5">Компания — юридические лица и индивидуальные предприниматели (ИП, ООО, ЧУП)</p>
				<p>Заполните реквизиты организации для заключения договора и выставления актов.</p>
			</div>

			<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
				<div class="sm:col-span-2">
					<label for="compLegalName" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Юридическое наименование *</label>
					<input
						id="compLegalName"
						type="text"
						bind:value={compLegalName}
						placeholder="ООО «Уютный дом» или ИП Иванов И.И."
						class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.compLegalName ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
					/>
					{#if formErrors.compLegalName}<p class="mt-1 text-[11px] text-rose-600">{formErrors.compLegalName}</p>{/if}
				</div>

				<div>
					<label for="compBrandName" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Брендовое название (для гостей)</label>
					<input
						id="compBrandName"
						type="text"
						bind:value={compBrandName}
						placeholder="Flickey Apartments"
						class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 border-zinc-200 dark:border-border dark:bg-card"
					/>
				</div>

				<div>
					<label for="compWebsite" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Сайт компании</label>
					<input
						id="compWebsite"
						type="url"
						bind:value={compWebsite}
						placeholder="https://example.by"
						class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 border-zinc-200 dark:border-border dark:bg-card"
					/>
				</div>

				<div>
					<label for="compEmail" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Email для отчетов и актов *</label>
					<input
						id="compEmail"
						type="email"
						bind:value={compEmail}
						placeholder="accounting@company.by"
						class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.compEmail ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
					/>
					{#if formErrors.compEmail}<p class="mt-1 text-[11px] text-rose-600">{formErrors.compEmail}</p>{/if}
				</div>

				<div>
					<label for="compPhone" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Телефон компании *</label>
					<input
						id="compPhone"
						type="tel"
						bind:value={compPhone}
						placeholder="+375 17 123-45-67"
						class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.compPhone ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
					/>
					{#if formErrors.compPhone}<p class="mt-1 text-[11px] text-rose-600">{formErrors.compPhone}</p>{/if}
				</div>
			</div>

			<!-- Legal Address -->
			<div class="border-t border-zinc-100 pt-4 dark:border-border/60 space-y-3">
				<h4 class="text-xs font-bold uppercase tracking-wider text-zinc-400 dark:text-zinc-500">Юридический адрес</h4>
				<div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
					<div>
						<label for="compPostalCode" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Индекс *</label>
						<input
							id="compPostalCode"
							type="text"
							maxlength="6"
							bind:value={compPostalCode}
							placeholder="220000"
							class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.compPostalCode ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
						/>
					</div>
					<div>
						<label for="compCity" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Город *</label>
						<input
							id="compCity"
							type="text"
							bind:value={compCity}
							placeholder="г. Минск"
							class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.compCity ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
						/>
					</div>
					<div>
						<label for="compStreetBuilding" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Улица, дом, офис *</label>
						<input
							id="compStreetBuilding"
							type="text"
							bind:value={compStreetBuilding}
							placeholder="ул. Ленина, д. 10, оф. 4"
							class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.compStreetBuilding ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
						/>
					</div>
				</div>

				<div class="pt-2">
					<label class="flex items-center gap-2 cursor-pointer text-xs font-medium text-zinc-700 dark:text-zinc-300">
						<input
							type="checkbox"
							bind:checked={compActualSame}
							class="rounded border-zinc-300 text-blue-600 focus:ring-blue-500"
						/>
						<span>Фактический адрес совпадает с юридическим</span>
					</label>
				</div>

				{#if !compActualSame}
					<div class="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2">
						<div>
							<label for="compActualPostalCode" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Факт. индекс *</label>
							<input
								id="compActualPostalCode"
								type="text"
								bind:value={compActualPostalCode}
								placeholder="220000"
								class="w-full rounded-xl border px-3 py-2 text-xs border-zinc-200 dark:border-border dark:bg-card"
							/>
						</div>
						<div>
							<label for="compActualCity" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Факт. город *</label>
							<input
								id="compActualCity"
								type="text"
								bind:value={compActualCity}
								placeholder="г. Минск"
								class="w-full rounded-xl border px-3 py-2 text-xs border-zinc-200 dark:border-border dark:bg-card"
							/>
						</div>
						<div>
							<label for="compActualStreetBuilding" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Факт. улица/офис *</label>
							<input
								id="compActualStreetBuilding"
								type="text"
								bind:value={compActualStreetBuilding}
								placeholder="ул. Немига, 5"
								class="w-full rounded-xl border px-3 py-2 text-xs border-zinc-200 dark:border-border dark:bg-card"
							/>
						</div>
					</div>
				{/if}
			</div>

			<!-- Bank & Legal Details -->
			<div class="border-t border-zinc-100 pt-4 dark:border-border/60 space-y-4">
				<h4 class="text-xs font-bold uppercase tracking-wider text-zinc-400 dark:text-zinc-500">Банковские реквизиты и ЕГР</h4>

				<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
					<div>
						<label for="compUNP" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">УНП (9 цифр) *</label>
						<input
							id="compUNP"
							type="text"
							maxlength="9"
							bind:value={compUNP}
							placeholder="190000000"
							class="w-full font-mono rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.compUNP ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
						/>
						{#if formErrors.compUNP}<p class="mt-1 text-[11px] text-rose-600">{formErrors.compUNP}</p>{/if}
					</div>

					<div>
						<label for="compEGR" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Рег. номер в ЕГР *</label>
						<input
							id="compEGR"
							type="text"
							bind:value={compEGR}
							placeholder="190000000"
							class="w-full font-mono rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.compEGR ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
						/>
						{#if formErrors.compEGR}<p class="mt-1 text-[11px] text-rose-600">{formErrors.compEGR}</p>{/if}
					</div>

					<div>
						<label for="compBIC" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">БИК банка (BIC) *</label>
						<input
							id="compBIC"
							type="text"
							maxlength="11"
							bind:value={compBIC}
							placeholder="BPSBBY2X"
							class="w-full font-mono uppercase rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.compBIC ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
						/>
						{#if formErrors.compBIC}<p class="mt-1 text-[11px] text-rose-600">{formErrors.compBIC}</p>{/if}
					</div>
				</div>

				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<div>
						<label for="compBankName" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Наименование банка *</label>
						<input
							id="compBankName"
							type="text"
							bind:value={compBankName}
							placeholder="ОАО «Сбер Банк»"
							class="w-full rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.compBankName ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
						/>
					</div>
					<div>
						<label for="compIBAN" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Номер счета (IBAN) *</label>
						<input
							id="compIBAN"
							type="text"
							maxlength="34"
							bind:value={compIBAN}
							placeholder="BY00BPSB00000000000000000000"
							class="w-full font-mono uppercase rounded-xl border px-3 py-2 text-xs transition focus:outline-none focus:ring-2 focus:ring-blue-500 {formErrors.compIBAN ? 'border-rose-400 bg-rose-50/30' : 'border-zinc-200 dark:border-border dark:bg-card'}"
						/>
						{#if formErrors.compIBAN}<p class="mt-1 text-[11px] text-rose-600">{formErrors.compIBAN}</p>{/if}
					</div>
				</div>
			</div>

			<!-- Contact Person -->
			<div class="border-t border-zinc-100 pt-4 dark:border-border/60 space-y-3">
				<h4 class="text-xs font-bold uppercase tracking-wider text-zinc-400 dark:text-zinc-500">Контактное лицо</h4>
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
					<div>
						<label for="compContactName" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">ФИО *</label>
						<input
							id="compContactName"
							type="text"
							bind:value={compContactName}
							placeholder="Петров Петр Петрович"
							class="w-full rounded-xl border px-3 py-2 text-xs border-zinc-200 dark:border-border dark:bg-card"
						/>
					</div>
					<div>
						<label for="compContactPosition" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Должность *</label>
						<input
							id="compContactPosition"
							type="text"
							bind:value={compContactPosition}
							placeholder="Директор / Управляющий"
							class="w-full rounded-xl border px-3 py-2 text-xs border-zinc-200 dark:border-border dark:bg-card"
						/>
					</div>
					<div>
						<label for="compContactEmail" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Email *</label>
						<input
							id="compContactEmail"
							type="email"
							bind:value={compContactEmail}
							placeholder="petrov@company.by"
							class="w-full rounded-xl border px-3 py-2 text-xs border-zinc-200 dark:border-border dark:bg-card"
						/>
					</div>
					<div>
						<label for="compContactPhone" class="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">Телефон *</label>
						<input
							id="compContactPhone"
							type="tel"
							bind:value={compContactPhone}
							placeholder="+375 29 111-22-33"
							class="w-full rounded-xl border px-3 py-2 text-xs border-zinc-200 dark:border-border dark:bg-card"
						/>
					</div>
				</div>
			</div>
		{/if}

		<!-- Document Attachments Section -->
		<div class="border-t border-zinc-100 pt-5 dark:border-border/60 space-y-3">
			<div class="flex items-center justify-between">
				<div>
					<h4 class="text-xs font-bold text-zinc-900 dark:text-foreground">Скан-копии документов *</h4>
					<p class="text-[11px] text-zinc-500 dark:text-muted-foreground mt-0.5">
						{#if activeTab === 'individual'}
							Скан паспорта (стр. 31, 32-33, страница с регистрацией) или вида на жительство (PDF, JPG, PNG)
						{:else}
							Свидетельство о госрегистрации в ЕГР, устав (1-2 стр.) или доверенность (PDF, JPG, PNG)
						{/if}
					</p>
				</div>
			</div>

			<!-- Attached documents list -->
			{#if attachedDocs.length > 0}
				<div class="space-y-2">
					{#each attachedDocs as doc (doc.id)}
						<div class="flex items-center justify-between rounded-xl border border-zinc-200 bg-zinc-50 p-2.5 text-xs dark:border-border dark:bg-muted/30">
							<div class="flex items-center gap-2 min-w-0">
								<FileText class="h-4 w-4 text-blue-600 shrink-0" />
								<span class="truncate font-medium text-zinc-900 dark:text-foreground">{doc.name}</span>
							</div>
							<button
								type="button"
								onclick={() => removeDoc(doc.id)}
								class="text-zinc-400 hover:text-rose-600 p-1 rounded-lg transition"
							>
								<X class="h-3.5 w-3.5" />
							</button>
						</div>
					{/each}
				</div>
			{/if}

			<!-- Upload Doc Button -->
			<label class="inline-flex items-center gap-2 rounded-xl border border-dashed border-zinc-300 p-3 text-xs font-medium text-zinc-700 hover:border-blue-500 hover:bg-blue-50/20 cursor-pointer transition w-full justify-center dark:border-border dark:text-zinc-300">
				{#if isUploadingDoc}
					<Loader2 class="h-4 w-4 text-blue-600 animate-spin" />
					<span>Загрузка файла...</span>
				{:else}
					<UploadCloud class="h-4 w-4 text-zinc-500" />
					<span>Прикрепить скан документа (PDF, JPG, PNG до 25 МБ)</span>
				{/if}
				<input
					type="file"
					accept="application/pdf,image/jpeg,image/png"
					class="sr-only"
					onchange={(e) => handleDocUpload(e.currentTarget.files)}
					disabled={isUploadingDoc}
				/>
			</label>
			{#if formErrors.documents}<p class="text-[11px] text-rose-600">{formErrors.documents}</p>{/if}
		</div>

		<!-- Law of Republic of Belarus № 99-З Consent Checkbox -->
		<div class="border-t border-zinc-100 pt-4 dark:border-border/60">
			<label class="flex items-start gap-2.5 cursor-pointer text-xs text-zinc-600 dark:text-muted-foreground">
				<input
					type="checkbox"
					bind:checked={consentGiven}
					class="mt-0.5 rounded border-zinc-300 text-blue-600 focus:ring-blue-500"
				/>
				<span>
					Я подтверждаю достоверность предоставленных сведений и даю согласие на обработку персональных данных в соответствии с <strong class="text-zinc-800 dark:text-zinc-200">Законом Республики Беларусь от 07.05.2021 № 99-З «О защите персональных данных»</strong>.
				</span>
			</label>
			{#if formErrors.consent}<p class="mt-1 text-[11px] text-rose-600 pl-6">{formErrors.consent}</p>{/if}
		</div>

		<!-- Submit Button -->
		<div class="pt-2">
			<button
				type="button"
				onclick={handleSubmit}
				disabled={isSubmitting}
				class="w-full flex items-center justify-center gap-2 rounded-2xl bg-zinc-900 py-3.5 px-6 text-xs font-bold text-white shadow-sm transition hover:bg-zinc-800 active:scale-98 disabled:opacity-50 cursor-pointer dark:bg-primary dark:text-primary-foreground"
			>
				{#if isSubmitting}
					<Loader2 class="h-4 w-4 animate-spin" />
					<span>Отправка на проверку...</span>
				{:else}
					<ShieldCheck class="h-4 w-4" />
					<span>Отправить документы на проверку</span>
				{/if}
			</button>
		</div>
	</div>
</div>
