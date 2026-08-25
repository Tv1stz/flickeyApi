<script lang="ts">
	import { FileText, Eye, CheckCircle2, XCircle } from 'lucide-svelte';

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
</script>

<div class="space-y-4">
	<h4 class="text-sm font-bold text-foreground flex items-center gap-2">
		<FileText class="h-4 w-4 text-primary" />
		<span>Предоставленные документы верификации ({documents.length})</span>
	</h4>

	{#if documents.length === 0}
		<div class="p-6 rounded-2xl bg-muted/20 border border-dashed border-border text-center text-xs text-muted-foreground">
			Документы не загружены.
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
			{#each documents as doc}
				<div class="p-3.5 rounded-xl border border-border bg-card hover:bg-muted/30 transition-colors flex items-center justify-between">
					<div class="flex items-center gap-2.5 overflow-hidden">
						<div class="p-2 rounded-lg bg-primary/10 text-primary shrink-0">
							<FileText class="h-4 w-4" />
						</div>
						<div class="truncate">
							<p class="text-xs font-semibold text-foreground truncate">{doc.name}</p>
							<p class="text-[10px] text-muted-foreground uppercase">{doc.status || 'Загружен'}</p>
						</div>
					</div>
					<button
						type="button"
						onclick={() => (selectedDoc = doc)}
						class="p-2 rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground transition-colors shrink-0"
						title="Просмотр документа"
					>
						<Eye class="h-4 w-4" />
					</button>
				</div>
			{/each}
		</div>
	{/if}

	{#if selectedDoc}
		<div class="fixed inset-0 z-50 bg-black/70 backdrop-blur-xs flex items-center justify-center p-4">
			<div class="bg-card border border-border rounded-2xl shadow-2xl max-w-2xl w-full p-6 space-y-4 overflow-hidden">
				<div class="flex items-center justify-between border-b border-border pb-4">
					<h3 class="text-sm font-bold text-foreground truncate">{selectedDoc.name}</h3>
					<button type="button" onclick={() => (selectedDoc = null)} class="text-muted-foreground hover:text-foreground text-xl font-bold">×</button>
				</div>
				<div class="min-h-[300px] flex items-center justify-center bg-muted/30 rounded-xl p-4">
					<p class="text-xs text-muted-foreground text-center">
						Безопасный просмотр документа: <br />
						<a href={selectedDoc.url} target="_blank" rel="noreferrer" class="text-primary underline font-medium mt-2 inline-block">
							Открыть оригинал документа
						</a>
					</p>
				</div>
			</div>
		</div>
	{/if}
</div>
