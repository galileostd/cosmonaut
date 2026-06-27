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

	const id = page.params.id ?? '';
	const [namespace, name] = id.includes('--') ? id.split('--') : ['cosmonaut', id];

	let component = $state<Component | null>(null);
	let loading = $state(true);
	let error = $state('');
	let lastUpdated = $state<Date | null>(null);

	let logLines = $state<string[]>([]);
	let logsLoading = $state(false);
	let logsError = $state('');
	let logsTail = $state(100);
	let logsLastUpdated = $state<Date | null>(null);

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

	async function loadLogs() {
		logsLoading = true;
		logsError = '';
		try {
			const res = await fetch(`/api/v1/components/${namespace}/${name}/logs?tail=${logsTail}`, {
				cache: 'no-store',
				headers: { 'Cache-Control': 'no-cache' }
			});
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			const data = await res.json();
			logLines = data.lines ?? [];
			logsLastUpdated = new Date();
		} catch (e: unknown) {
			logsError = e instanceof Error ? e.message : String(e);
		} finally {
			logsLoading = false;
		}
	}

	let interval: ReturnType<typeof setInterval>;
	let logsInterval: ReturnType<typeof setInterval>;

	onMount(() => {
		load();
		interval = setInterval(load, 30000);
		loadLogs();
		logsInterval = setInterval(loadLogs, 15000);
	});

	onDestroy(() => {
		if (interval) clearInterval(interval);
		if (logsInterval) clearInterval(logsInterval);
	});

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

	const HIDDEN_CONFIG = ['token'];

	function visibleConfig(config: Record<string, string>) {
		return Object.entries(config).filter(([k]) => !HIDDEN_CONFIG.includes(k));
	}

	function logClass(line: string): string {
		const l = line.toLowerCase();
		if (l.includes('error') || l.includes('fatal') || l.includes('exception')) return 'log-err';
		if (l.includes('warn')) return 'log-warn';
		if (l.includes('info')) return 'log-info';
		if (l.includes('debug')) return 'log-debug';
		return 'log-default';
	}
</script>

<svelte:head>
	<title>Cosmonaut · {name}</title>
</svelte:head>

<div class="main">
	<div class="breadcrumb">
		<button class="back-btn" onclick={() => void goto('/services')}>
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

		<div class="msg-bar msg-{healthColor(component.health)}">
			{component.message}
		</div>

		<div class="detail-grid">
			<div class="detail-left">
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

				{#if visibleConfig(component.config).length > 0}
					<div class="panel">
						<div class="panel-head">Configuration</div>
						<div class="panel-body">
							{#each visibleConfig(component.config) as [key, value] (key)}
								<div class="kv-row">
									<span class="kv-key">{key}</span>
									<span class="kv-val mono">{value}</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>

			<div class="detail-right">
				<div class="caps-panel">
					<div class="panel-head">
						Capabilities
						<span class="panel-badge">{component.capabilities.length}</span>
					</div>
					<div class="caps-body">
						{#each component.capabilities as cap (cap)}
							<div class="cap-row">
								<span class="cap-dot"></span>
								<span class="cap-name">{cap}</span>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>

		<div class="logs-panel">
			<div class="logs-head">
				<span class="logs-title">Service Logs</span>
				<div class="logs-controls">
					{#if logsLastUpdated}
						<span class="p-meta">Updated {logsLastUpdated.toLocaleTimeString()}</span>
					{/if}
					<select class="tail-select" bind:value={logsTail} onchange={loadLogs}>
						<option value={50}>Last 50</option>
						<option value={100}>Last 100</option>
						<option value={200}>Last 200</option>
						<option value={500}>Last 500</option>
						<option value={0}>All</option>
					</select>
					<button class="p-btn" onclick={loadLogs} disabled={logsLoading}>
						{logsLoading ? '⟳' : '↻'} Refresh
					</button>
				</div>
			</div>

			{#if logsError}
				<div class="logs-error">⚠ {logsError}</div>
			{:else if logsLoading && logLines.length === 0}
				<div class="logs-empty">Loading logs...</div>
			{:else if logLines.length === 0}
				<div class="logs-empty">No log lines returned.</div>
			{:else}
				<div class="logs-body">
					{#each logLines as line, i (i)}
						<div class="log-line {logClass(line)}">
							<span class="log-n">{i + 1}</span>
							<span class="log-text">{line}</span>
						</div>
					{/each}
				</div>
			{/if}
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
	.ns-tag { font-family: 'JetBrains Mono', monospace; font-size: 10px; color: var(--text-3); }

	/* ── Grid: left column defines the row height; right column matches it ── */
	.detail-grid {
		display: grid;
		grid-template-columns: 1fr 360px;
		gap: 16px;
		margin-bottom: 20px;
		align-items: stretch;
	}
	.detail-left { display: flex; flex-direction: column; gap: 16px; }

	/* Right column is a positioning context; the panel is pinned to its bounds
	   so its height is forced to equal the left column's height (the row height).
	   The capabilities list then scrolls inside that fixed height. */
	.detail-right { position: relative; }
	.caps-panel {
		position: absolute;
		inset: 0;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		border: 1px solid var(--border-main);
		background: var(--bg-surface);
	}

	.panel { border: 1px solid var(--border-main); background: var(--bg-surface); }
	.panel-head {
		padding: 0 14px; height: 36px; border-bottom: 1px solid var(--border-main);
		display: flex; align-items: center; justify-content: space-between;
		font-family: 'JetBrains Mono', monospace; font-size: 11px; font-weight: 700;
		letter-spacing: 0.12em; text-transform: uppercase;
		background: var(--bg-hover); color: var(--text-1);
		flex-shrink: 0;
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

	.caps-body { padding: 8px 0; flex: 1; overflow-y: auto; min-height: 0; }
	.cap-row {
		display: flex; align-items: center; gap: 10px;
		padding: 5px 14px; border-bottom: 1px solid var(--border-sub);
	}
	.cap-row:last-child { border-bottom: none; }
	.cap-dot { width: 5px; height: 5px; border-radius: 50%; background: var(--text-3); flex-shrink: 0; }
	.cap-name { font-family: 'JetBrains Mono', monospace; font-size: 12px; color: var(--text-2); }

	/* ── Logs ── */
	.logs-panel { border: 1px solid var(--border-main); background: var(--bg-surface); }
	.logs-head {
		padding: 0 14px; height: 36px;
		border-bottom: 1px solid var(--border-main);
		display: flex; align-items: center; justify-content: space-between;
		background: var(--bg-hover);
	}
	.logs-title {
		font-family: 'JetBrains Mono', monospace; font-size: 11px; font-weight: 700;
		letter-spacing: 0.12em; text-transform: uppercase; color: var(--text-1);
	}
	.logs-controls { display: flex; align-items: center; gap: 8px; }
	.tail-select {
		font-family: 'JetBrains Mono', monospace; font-size: 11px;
		background: var(--bg-inset); border: 1px solid var(--border-main);
		color: var(--text-2); padding: 3px 6px; cursor: pointer;
	}
	.logs-error {
		padding: 12px 14px;
		font-family: 'JetBrains Mono', monospace; font-size: 12px; color: var(--red);
	}
	.logs-empty {
		padding: 24px 14px; text-align: center;
		font-family: 'JetBrains Mono', monospace; font-size: 12px; color: var(--text-3);
	}
	.logs-body {
		font-family: 'JetBrains Mono', monospace; font-size: 12px;
		max-height: 500px; overflow-y: auto;
		background: var(--bg-inset, #0d0d0d);
	}
	.log-line {
		display: flex; gap: 12px;
		padding: 2px 14px; border-bottom: 1px solid color-mix(in srgb, var(--border-sub) 40%, transparent);
		line-height: 1.5;
	}
	.log-line:last-child { border-bottom: none; }
	.log-n {
		flex-shrink: 0; width: 36px; text-align: right;
		color: var(--text-3); font-size: 11px; user-select: none;
	}
	.log-text { color: var(--text-2); white-space: pre-wrap; word-break: break-all; }
	.log-err .log-text { color: var(--red, #C8102E); }
	.log-warn .log-text { color: var(--yellow, #B8862A); }
	.log-info .log-text { color: var(--text-1); }
	.log-debug .log-text { color: var(--text-3); }
	.log-default .log-text { color: var(--text-2); }

	@media (max-width: 1024px) {
		.detail-grid { grid-template-columns: 1fr; }
		.detail-right { position: static; }
		.caps-panel { position: static; }
		.caps-body { max-height: 400px; }
	}
</style>