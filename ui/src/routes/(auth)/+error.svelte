<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
</script>

<div class="main">
	<div class="err-wrap">
		<div class="err-code">{page.status}</div>
		<div class="err-title">{page.status === 404 ? 'Page not found' : 'Something went wrong'}</div>
		<div class="err-msg">{page.error?.message ?? 'An unexpected error occurred.'}</div>
		<button class="err-btn" onclick={() => goto('/cluster')}>← Back to Cluster</button>
	</div>
</div>

<style>
	.main { height: 100%; display: flex; align-items: center; justify-content: center; background: var(--bg-base); }
	.err-wrap { display: flex; flex-direction: column; align-items: center; gap: 12px; text-align: center; }
	.err-code { font-family: 'Inter Tight', sans-serif; font-size: 72px; font-weight: 800; letter-spacing: -0.04em; color: var(--red); line-height: 1; }
	.err-title { font-family: 'Inter Tight', sans-serif; font-size: 20px; font-weight: 700; color: var(--text-1); }
	.err-msg { font-family: 'JetBrains Mono', monospace; font-size: 12px; color: var(--text-3); max-width: 400px; }
	.err-btn { margin-top: 8px; font-family: 'JetBrains Mono', monospace; font-size: 11px; font-weight: 600; padding: 8px 16px; background: var(--bg-surface); border: 1px solid var(--border-main); color: var(--text-2); cursor: pointer; }
	.err-btn:hover { background: var(--bg-hover); color: var(--text-1); }
</style>
