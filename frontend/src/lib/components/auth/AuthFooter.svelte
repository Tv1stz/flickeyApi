<!-- src/lib/components/auth/AuthFooter.svelte -->
<script lang="ts">
	import { resolve } from '$app/paths';
</script>

<footer class="safe-bottom border-t border-zinc-100 bg-zinc-50/50 px-4 py-4">
	<p class="mx-auto max-w-lg text-center text-[11px] leading-relaxed text-zinc-400">
		Продолжая, вы соглашаетесь с
		<a
			href={resolve('/terms')}
			class="underline underline-offset-2 transition-colors hover:text-zinc-600"
		>
			условиями использования
		</a>
		и
		<a
			href={resolve('/privacy')}
			class="underline underline-offset-2 transition-colors hover:text-zinc-600"
		>
			политикой конфиденциальности
		</a>
	</p>
</footer>
