<script lang="ts">
	import { FileText, Eye, CheckCircle2, XCircle, ExternalLink, Download, X } from 'lucide-svelte';

	interface DocumentItem {
		id: string;
		name: string;
		url: string;
		status?: string;
	}

	interface Props {
		documents: DocumentItem[];
	}

	let { documents = [] }: Props = $props();
	let selectedDoc = $state<DocumentItem | null>(null);

	function isImage(url: string): boolean {
		return /\.(jpe?g|png|webp|gif|svg)(\?.*)?$/i.test(url);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && selectedDoc) {
			selectedDoc = null;
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<h4 class="text-sm font-bold text-zinc-900 flex items-center gap-2 dark:text-foreground">
			<FileText class="h-4 w-4 text-zinc-600 dark:text-zinc-400" />
			<span>Прикрепленные документы ({documents.length})</span>
		</h4>
	</div>

	{#if documents.length === 0}
		<div class="rounded-2xl border border-dashed border-zinc-200 bg-zinc-50/50 p-6 text-center text-xs text-zinc-500 dark:border-border dark:bg-muted/20 dark:text-muted-foreground">
			Документы не прикреплены к данной заявке.
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
			{#each documents as doc}
				<div class="rounded-2xl border border-zinc-200/80 bg-white p-4 shadow-2xs hover:border-zinc-300 transition-colors flex items-center justify-between gap-3 dark:border-border dark:bg-card">
					<div class="flex items-center gap-3 overflow-hidden">
						<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300">
							<FileText class="h-5 w-5" />
						</div>
						<div class="truncate">
							<p class="text-xs font-semibold text-zinc-900 truncate dark:text-foreground">{doc.name}</p>
							<p class="text-[10px] text-zinc-400 uppercase font-medium">{doc.status || 'Загружен'}</p>
						</div>
					</div>

					<div class="flex items-center gap-1 shrink-0">
						<button
							type="button"
							onclick={() => (selectedDoc = doc)}
							class="flex h-8 w-8 items-center justify-center rounded-xl border border-zinc-200/80 hover:bg-zinc-50 text-zinc-600 transition-colors dark:border-border dark:text-muted-foreground dark:hover:text-foreground cursor-pointer"
							title="Просмотр документа"
						>
							<Eye class="h-4 w-4" />
						</button>
						<a
							href={doc.url}
							target="_blank"
							rel="noreferrer"
							class="flex h-8 w-8 items-center justify-center rounded-xl border border-zinc-200/80 hover:bg-zinc-50 text-zinc-600 transition-colors dark:border-border dark:text-muted-foreground dark:hover:text-foreground"
							title="Открыть в новой вкладке"
						>
							<ExternalLink class="h-4 w-4" />
						</a>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	<!-- Document Preview Modal -->
	{#if selectedDoc}
		<div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 overflow-y-auto">
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				class="fixed inset-0 bg-black/60 backdrop-blur-xs transition-opacity"
				onclick={() => (selectedDoc = null)}
			></div>

			<div class="relative z-50 w-full max-w-3xl rounded-3xl bg-white p-6 shadow-2xl border border-zinc-200/80 space-y-4 overflow-hidden dark:bg-card dark:border-border animate-in zoom-in-95">
				<div class="flex items-center justify-between border-b border-zinc-100 pb-3 dark:border-border">
					<div class="flex items-center gap-2.5 truncate pr-4">
						<FileText class="h-5 w-5 text-zinc-700 dark:text-foreground shrink-0" />
						<h3 class="text-sm font-bold text-zinc-900 truncate dark:text-foreground">{selectedDoc.name}</h3>
					</div>

					<button
						type="button"
						onclick={() => (selectedDoc = null)}
						class="rounded-full p-1.5 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-700 transition-colors dark:hover:bg-muted cursor-pointer"
						aria-label="Закрыть"
					>
						<X class="h-4 w-4" />
					</button>
				</div>

				<div class="min-h-[360px] max-h-[70vh] overflow-auto flex items-center justify-center bg-zinc-50 rounded-2xl p-4 dark:bg-muted/30">
					{#if isImage(selectedDoc.url)}
						<img
							src={selectedDoc.url}
							alt={selectedDoc.name}
							class="max-h-[60vh] max-w-full rounded-xl object-contain shadow-sm"
						/>
					{:else}
						<div class="text-center space-y-3 p-8">
							<FileText class="h-12 w-12 text-zinc-400 mx-auto" />
							<p class="text-xs text-zinc-500 max-w-sm">
								Файл в формате документа. Вы можете открыть его в новой вкладке браузера или загрузить для проверки.
							</p>
							<a
								href={selectedDoc.url}
								target="_blank"
								rel="noreferrer"
								class="inline-flex items-center gap-2 rounded-full bg-zinc-900 px-5 py-2 text-xs font-semibold text-white transition hover:bg-zinc-800 dark:bg-white dark:text-zinc-900"
							>
								<span>Открыть документ</span>
								<ExternalLink class="h-3.5 w-3.5" />
							</a>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>