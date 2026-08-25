<script lang="ts">
	import { Building2, Heart } from 'lucide-svelte';
</script>

<footer class="mt-auto border-t border-border bg-card py-10 text-muted-foreground text-xs">
	<div class="container mx-auto px-4 sm:px-6">
		<div class="grid grid-cols-1 md:grid-cols-4 gap-8 pb-8 border-b border-border">
			<!-- Col 1 -->
			<div class="space-y-3">
				<div class="flex items-center gap-2 font-bold text-foreground text-sm">
					<div class="flex h-6 w-6 items-center justify-center rounded-lg bg-primary text-primary-foreground">
						<Building2 class="h-3.5 w-3.5" />
					</div>
					<span>Flickey Marketplace</span>
				</div>
				<p class="leading-relaxed">
					Современная платформа посуточной аренды квартир, домов и усадеб в Беларуси.
				</p>
			</div>

			<!-- Col 2 -->
			<div>
				<h4 class="font-semibold text-foreground text-xs uppercase tracking-wider mb-3">Гостям</h4>
				<ul class="space-y-2">
					<li><a href="/" class="hover:text-foreground transition-colors">Поиск жилья</a></li>
					<li><a href="/" class="hover:text-foreground transition-colors">Правила бронирования</a></li>
					<li><a href="/" class="hover:text-foreground transition-colors">Безопасность</a></li>
				</ul>
			</div>

			<!-- Col 3 -->
			<div>
				<h4 class="font-semibold text-foreground text-xs uppercase tracking-wider mb-3">Хостам</h4>
				<ul class="space-y-2">
					<li><a href="/host/new" class="hover:text-foreground transition-colors">Сдать жилье</a></li>
					<li><a href="/host" class="hover:text-foreground transition-colors">Панель управления</a></li>
					<li><a href="/host" class="hover:text-foreground transition-colors">Верификация объектов</a></li>
				</ul>
			</div>

			<!-- Col 4 -->
			<div>
				<h4 class="font-semibold text-foreground text-xs uppercase tracking-wider mb-3">О проекте</h4>
				<p class="leading-relaxed mb-2">
					Сделано на <strong>Svelte 5</strong>, <strong>SvelteKit 2</strong> и <strong>Go API</strong>.
				</p>
				<p class="text-[11px] text-muted-foreground">
					© {new Date().getFullYear()} Flickey. Все права защищены.
				</p>
			</div>
		</div>

		<div class="pt-6 flex flex-col sm:flex-row items-center justify-between gap-4 text-[11px]">
			<p>Сервис посуточной аренды жилья Flickey</p>
			<div class="flex items-center gap-1">
				<span>Разработано с</span>
				<Heart class="h-3 w-3 text-destructive fill-destructive" />
				<span>для комфортных путешествий</span>
			</div>
		</div>
	</div>
</footer>
