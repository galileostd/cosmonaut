<script lang="ts">
	import { page } from '$app/state';
	import { locales, localizeHref } from '$lib/paraglide/runtime';
	import { resolve } from '$app/paths';
	import type { Pathname } from '$app/types';
	import './layout.css';

	let { children } = $props();

	// rotas que não usam o shell autenticado
	const PUBLIC_ROUTES = ['/login', '/setup', '/forgot-password'];
	const isPublic = $derived(PUBLIC_ROUTES.some(r => page.url.pathname.startsWith(r)));
</script>

<svelte:head>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Inter+Tight:wght@700;800&family=JetBrains+Mono:wght@400;500;600;700&display=swap" rel="stylesheet" />
</svelte:head>

<!-- Apenas um contêiner simples, sem grid -->
{#if isPublic}
	{@render children()}
{:else}
	<div data-theme="dark" class="auth-shell">
		{@render children()}
	</div>
{/if}

<!-- paraglide locale links (hidden) -->
<div style="display:none">
	{#each locales as locale (locale)}
		<a href={resolve(localizeHref(page.url.pathname, { locale }) as Pathname)}>{locale}</a>
	{/each}
</div>

<style>
	.auth-shell {
		height: 100vh;
		overflow: hidden;
	}
</style>