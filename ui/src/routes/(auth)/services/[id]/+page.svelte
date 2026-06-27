<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';

	interface Component {
		name: string;
		namespace: string;
		plugin: string;
		type: string;
		endpoint: string;
		config: Record<string, string>;
		health: string;
		message: string;
		last_checked: string;
		capabilities: string[];
		created_at: string;
		updated_at: string;
	}

	// URL param: namespace--name
	const id = page.params.id;
	const [namespace, name] = id.includes('--') ? id.split('--') : ['cosmonaut', id];

	let component = $state<Component | null>(null);
	let loading = $state(true);
	let error = $state('');
	let lastUpdated = $state<Date | null>(null);

	async function load() {
		loading = true;
		error = '';
		try {
			const res = await fetch(`/api/v1/components/${namespace}/${name}`, {
				cache: 'no-store',
				headers: { 'Cache-Control': 'no-cache' }
			});
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			component = await res.json();
			lastUpdated = new Date();
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	}

	let interval: ReturnType<typeof setInterval>;
	onMount(() => { load(); interval = setInterval(load, 30000); });
	onDestroy(() => { if (interval) clearInterval(interval); });

	function healthColor(h: string) {
		if (h === 'healthy') return 'ok';
		if (h === 'degraded') return 'warn';
		return 'err';
	}

	function typeLabel(t: string) {
		const map: Record<string, string> = {
			'processing': 'Processing',
			'catalog': 'Catalog',
			'query-engine': 'Query Engine',
			'streaming': 'Streaming',
			'orchestration': 'Orchestration',
		};
		return map[t] ?? t;
	}

	function fmtDate(iso: string) {
		if (!iso) return '—';
		return new Date(iso).toLocaleString('en-US', {
			month: 'short', day: 'numeric', year: 'numeric',
			hour: '2-digit', minute: '2-digit'
		});
	}

	// config keys to hide from display
	const HIDDEN_CONFIG = ['token'];

	function visibleConfig(config: Record<string, string>) {
		return Object.entries(config).filter(([k]) => !HIDDEN_CONFIG.includes(k));
	}
</script>

<svelte:head>
	<title>Cosmonaut · {name}</title>
</svelte:head>

<div class="main">
	<!-- Breadcrumb -->
	<div class="breadcrumb">
		<button class="back-btn" onclick={() => goto('/services')}>
			<svg width="14" height="14" viewBox="0 0 16 16" fill="none">
				<path d="M10 3l-5 5 5 5" stroke="currentColor" stroke-width="1.5"/>
			</svg>
			Services
		</button>
		<span class="sep">/</span>
		<span class="current">{name}</span>
	</div>

	{#if error}
		<div class="error-banner">
			<span>⚠</span> {error}
			<button class="retry-btn" onclick={load}>Retry</button>
		</div>
	{/if}

	{#if loading && !component}
		<div class="loading">Loading...</div>
	{:else if component}
		<!-- Header -->
		<div class="svc-header">
			<div class="svc-icon svc-icon-{component.plugin}">
				{component.plugin.slice(0, 2).toUpperCase()}
			</div>
			<div class="svc-meta">
				<div class="svc-name">{component.name}</div>
				<div class="svc-sub">
					<span class="tag tag-{healthColor(component.health)}">{component.health}</span>
					<span class="type-tag">{typeLabel(component.type)}</span>
					<span class="ns-tag">{component.namespace}</span>
				</div>
			</div>
			<div class="svc-actions">
				{#if lastUpdated}
					<span class="p-meta">Updated {lastUpdated.toLocaleTimeString()}</span>
				{/if}
				<button class="p-btn" onclick={load} disabled={loading}>↻ Refresh</button>
			</div>
		</div>

		<!-- Message bar -->
		<div class="msg-bar msg-{healthColor(component.health)}">
			{component.message}
		</div>

		<!-- Two-column layout -->
		<div class="detail-grid">
			<!-- Left: Info + Config -->
			<div class="detail-left">
				<!-- Service Info -->
				<div class="panel">
					<div class="panel-head">Service Info</div>
					<div class="panel-body">
						<div class="kv-row">
							<span class="kv-key">Plugin</span>
							<span class="kv-val mono">{component.plugin}</span>
						</div>
						<div class="kv-row">
							<span class="kv-key">Type</span>
							<span class="kv-val">{typeLabel(component.type)}</span>
						</div>
						<div class="kv-row">
							<span class="kv-key">Namespace</span>
							<span class="kv-val mono">{component.namespace}</span>
						</div>
						<div class="kv-row">
							<span class="kv-key">Endpoint</span>
							<span class="kv-val mono endpoint">{component.endpoint}</span>
						</div>
						<div class="kv-row">
							<span class="kv-key">Created</span>
							<span class="kv-val mono">{fmtDate(component.created_at)}</span>
						</div>
						<div class="kv-row">
							<span class="kv-key">Last checked</span>
							<span class="kv-val mono">{fmtDate(component.last_checked)}</span>
						</div>
					</div>
				</div>

				<!-- Config -->
				{#if visibleConfig(component.config).length > 0}
					<div class="panel">
						<div class="panel-head">Configuration</div>
						<div class="panel-body">
							{#each visibleConfig(component.config) as [key, value]}
								<div class="kv-row">
									<span class="kv-key">{key}</span>
									<span class="kv-val mono">{value}</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>

			<!-- Right: Capabilities -->
			<div class="detail-right">
				<div class="panel">
					<div class="panel-head">
						Capabilities
						<span class="panel-badge">{component.capabilities.length}</span>
					</div>
					<div class="panel-body caps-body">
						{#each component.capabilities as cap}
							<div class="cap-row">
								<span class="cap-dot"></span>
								<span class="cap-name">{cap}</span>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.main { padding: 20px 24px; overflow-y: auto; background: var(--bg-base); height: 100%; }

	.breadcrumb {
		display: flex; align-items: center; gap: 8px;
		margin-bottom: 20px;
		font-family: 'JetBrains Mono', monospace; font-size: 12px;
	}
	.back-btn {
		display: inline-flex; align-items: center; gap: 6px;
		background: none; border: none; color: var(--text-3);
		cursor: pointer; font-family: inherit; font-size: inherit; padding: 0;
	}
	.back-btn:hover { color: var(--text-1); }
	.sep { color: var(--text-3); }
	.current { color: var(--text-1); font-weight: 600; }

	.loading { padding: 40px; text-align: center; color: var(--text-3); font-family: 'JetBrains Mono', monospace; font-size: 12px; }

	.error-banner {
		padding: 12px 16px; margin-bottom: 20px;
		background: var(--red-deep); border: 1px solid var(--red);
		color: var(--text-1); font-family: 'JetBrains Mono', monospace; font-size: 12px;
		display: flex; align-items: center; gap: 10px;
	}
	.retry-btn {
		margin-left: auto; font-family: 'JetBrains Mono', monospace; font-size: 10px;
		padding: 4px 10px; background: var(--red); border: none; color: #fff;
		cursor: pointer; text-transform: uppercase; font-weight: 700;
	}

	.svc-header {
		display: flex; align-items: center; gap: 16px;
		margin-bottom: 12px; padding-bottom: 16px;
		border-bottom: 1px solid var(--border-main);
	}
	.svc-icon {
		width: 48px; height: 48px;
		display: flex; align-items: center; justify-content: center;
		font-family: 'Inter Tight', sans-serif; font-size: 14px; font-weight: 800;
		background: var(--bg-surface); border: 1px solid var(--border-main);
		color: var(--text-1); flex-shrink: 0;
	}
	.svc-icon-spark { background: var(--red); color: #F4F1E8; border-color: var(--red); }
	.svc-icon-polaris { background: var(--ocher, #B8862A); color: #F4F1E8; border-color: var(--ocher, #B8862A); }
	.svc-icon-trino { background: var(--bg-hover); }

	.svc-meta { flex: 1; }
	.svc-name {
		font-family: 'Inter Tight', sans-serif; font-size: 22px; font-weight: 800;
		letter-spacing: -0.02em; color: var(--text-1); margin-bottom: 6px;
	}
	.svc-sub { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }

	.svc-actions { display: flex; align-items: center; gap: 12px; margin-left: auto; }
	.p-meta { font-family: 'JetBrains Mono', monospace; font-size: 10px; color: var(--text-3); }
	.p-btn {
		font-family: 'JetBrains Mono', monospace; font-size: 11px; font-weight: 600;
		padding: 6px 12px; background: var(--bg-surface); border: 1px solid var(--border-main);
		color: var(--text-2); cursor: pointer;
	}
	.p-btn:hover:not(:disabled) { background: var(--bg-hover); color: var(--text-1); }
	.p-btn:disabled { opacity: 0.5; cursor: wait; }

	.msg-bar {
		padding: 10px 14px; margin-bottom: 20px;
		font-family: 'JetBrains Mono', monospace; font-size: 12px;
		border: 1px solid; border-left-width: 3px;
	}
	.msg-ok { border-color: var(--green); color: var(--green); background: color-mix(in srgb, var(--green) 8%, var(--bg-base)); }
	.msg-err { border-color: var(--red); color: var(--red); background: color-mix(in srgb, var(--red) 8%, var(--bg-base)); }
	.msg-warn { border-color: var(--yellow); color: var(--yellow); background: color-mix(in srgb, var(--yellow) 8%, var(--bg-base)); }

	.tag {
		display: inline-flex; align-items: center;
		font-family: 'JetBrains Mono', monospace; font-size: 10px; font-weight: 700;
		text-transform: uppercase; padding: 2px 6px; border: 1px solid;
	}
	.tag-ok { border-color: var(--green); color: var(--green); }
	.tag-err { border-color: var(--red); color: var(--red); }
	.tag-warn { border-color: var(--yellow); color: var(--yellow); }

	.type-tag {
		font-family: 'JetBrains Mono', monospace; font-size: 10px; font-weight: 600;
		text-transform: uppercase; padding: 2px 6px;
		border: 1px solid var(--border-main); color: var(--text-2); background: var(--bg-inset);
	}
	.ns-tag {
		font-family: 'JetBrains Mono', monospace; font-size: 10px;
		color: var(--text-3);
	}

	.detail-grid {
		display: grid; grid-template-columns: 1fr 360px; gap: 16px;
	}
	.detail-left { display: flex; flex-direction: column; gap: 16px; }
	.detail-right { display: flex; flex-direction: column; gap: 16px; }

	.panel { border: 1px solid var(--border-main); background: var(--bg-surface); }
	.panel-head {
		padding: 0 14px; height: 36px; border-bottom: 1px solid var(--border-main);
		display: flex; align-items: center; justify-content: space-between;
		font-family: 'JetBrains Mono', monospace; font-size: 11px; font-weight: 700;
		letter-spacing: 0.12em; text-transform: uppercase;
		background: var(--bg-hover); color: var(--text-1);
	}
	.panel-badge {
		font-family: 'JetBrains Mono', monospace; font-size: 10px; font-weight: 400;
		color: var(--text-3); background: var(--bg-inset);
		padding: 1px 6px; border: 1px solid var(--border-main);
	}
	.panel-body { padding: 4px 0; }

	.kv-row {
		display: grid; grid-template-columns: 140px 1fr;
		padding: 7px 14px; border-bottom: 1px solid var(--border-sub);
		align-items: start; gap: 12px;
	}
	.kv-row:last-child { border-bottom: none; }
	.kv-key {
		font-family: 'JetBrains Mono', monospace; font-size: 11px; font-weight: 600;
		text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-3);
	}
	.kv-val { font-size: 13px; color: var(--text-1); word-break: break-all; }
	.kv-val.mono { font-family: 'JetBrains Mono', monospace; font-size: 12px; }
	.endpoint { font-size: 11px; color: var(--text-2); }

	.caps-body { padding: 8px 0; max-height: 500px; overflow-y: auto; }
	.cap-row {
		display: flex; align-items: center; gap: 10px;
		padding: 5px 14px; border-bottom: 1px solid var(--border-sub);
	}
	.cap-row:last-child { border-bottom: none; }
	.cap-dot {
		width: 5px; height: 5px; border-radius: 50%;
		background: var(--text-3); flex-shrink: 0;
	}
	.cap-name {
		font-family: 'JetBrains Mono', monospace; font-size: 12px; color: var(--text-2);
	}

	@media (max-width: 1024px) {
		.detail-grid { grid-template-columns: 1fr; }
	}
</style>