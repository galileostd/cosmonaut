<script lang="ts">
	import { onMount } from 'svelte';
	import Sidebar from '$lib/components/layout/Sidebar.svelte';
	import Topbar from '$lib/components/layout/Topbar.svelte';
	import { theme } from '$lib/stores/theme.svelte';
	import { sidebar } from '$lib/stores/sidebar.svelte';

	let { children } = $props();
	onMount(() => theme.init());
</script>

<div class="app-shell" data-theme={theme.current} style="--sb: {sidebar.width}px">
	<Sidebar />
	<Topbar />
	<main class="main-content">
		{@render children()}
	</main>
</div>

<style>
	.app-shell {
		display: grid;
		grid-template-columns: var(--sb, 220px) 1fr;
		grid-template-rows: 48px 1fr;
		height: 100vh;
		overflow: hidden;
		background: var(--bg-base);
		color: var(--fg);
		font-family: var(--font-sans);
		-webkit-font-smoothing: antialiased;
		transition: grid-template-columns 180ms ease;
	}
	.main-content {
		overflow-y: auto;
		grid-column: 2;
		grid-row: 2;
		padding: 16px 20px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		min-width: 0;
	}
</style>