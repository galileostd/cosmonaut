<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
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

	interface ComponentsResponse {
		items: Component[];
		total: number;
		limit: number;
		offset: number;
	}

	let data = $state<ComponentsResponse>({ items: [], total: 0, limit: 50, offset: 0 });
	let loading = $state(true);
	let error = $state('');
	let lastUpdated = $state<Date | null>(null);

	async function load() {
		loading = true;
		error = '';
		try {
			const res = await fetch('/api/v1/components', {
				cache: 'no-store',
				headers: { 'Cache-Control': 'no-cache' }
			});
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			data = await res.json();
			lastUpdated = new Date();
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	}

	let interval: ReturnType<typeof setInterval>;
	onMount(() => { load(); interval = setInterval(load, 10000); });
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
		return new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	function fmtTime(iso: string) {
		if (!iso) return '—';
		return new Date(iso).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
	}

	
	const healthy = $derived(data.items.filter(c => c.health === 'healthy').length);
    const unhealthy = $derived(data.items.filter(c => c.health === 'unhealthy').length);
    const degraded = $derived(data.items.filter(c => c.health === 'degraded').length);
    const byType = $derived(data.items.reduce((acc: Record<string, number>, c) => {
        acc[c.type] = (acc[c.type] ?? 0) + 1;
        return acc;
    }, {} as Record<string, number>));
</script>

<svelte:head>
	<title>Cosmonaut · Services</title>
</svelte:head>

<div class="main">
	<div class="p-header">
		<div class="p-title">Services</div>
		<div class="p-actions">
			{#if lastUpdated}
				<span class="p-meta">Updated {fmtTime(lastUpdated.toISOString())}</span>
			{/if}
			<button class="p-btn" onclick={load} disabled={loading}>
				{loading ? '⟳ Refreshing...' : '↻ Refresh'}
			</button>
		</div>
	</div>

	{#if error}
		<div class="error-banner">
			<span>⚠</span> {error}
			<button class="retry-btn" onclick={load}>Retry</button>
		</div>
	{/if}

	<!-- Stats -->
	<div class="grid-4">
		<div class="m-box">
			<div class="m-lbl">Total Services</div>
			<div class="m-val">{data.total}</div>
			<div class="m-sub">registered components</div>
			<div class="p-track"><div class="p-fill" style="width:100%"></div></div>
		</div>
		<div class="m-box">
			<div class="m-lbl">Healthy</div>
			<div class="m-val green">{healthy}</div>
			<div class="m-sub">{data.total > 0 ? ((healthy / data.total) * 100).toFixed(0) : 0}% availability</div>
			<div class="p-track"><div class="p-fill green" style="width:{data.total > 0 ? (healthy/data.total)*100 : 0}%"></div></div>
		</div>
		<div class="m-box">
			<div class="m-lbl">Unhealthy</div>
			<div class="m-val {unhealthy > 0 ? 'red' : ''}">{unhealthy}</div>
			<div class="m-sub">{degraded} degraded</div>
			<div class="p-track"><div class="p-fill red" style="width:{data.total > 0 ? (unhealthy/data.total)*100 : 0}%"></div></div>
		</div>
		<div class="m-box">
			<div class="m-lbl">Service Types</div>
			<div class="m-val">{Object.keys(byType).length}</div>
			<div class="m-sub">{Object.entries(byType).map(([k,v]) => `${v} ${typeLabel(k)}`).join(' · ') || 'none'}</div>
			<div class="p-track"><div class="p-fill" style="width:100%"></div></div>
		</div>
	</div>

	<!-- Table -->
	<div class="board">
		<div class="b-head">
			Services
			<span class="meta">{data.total} component{data.total !== 1 ? 's' : ''}</span>
		</div>

		{#if loading && data.items.length === 0}
			<div class="empty-state">Loading...</div>
		{:else if data.items.length === 0}
			<div class="empty-state">No services registered. Create a CosmoComponent to get started.</div>
		{:else}
			<table>
				<thead>
					<tr>
						<th>Name</th>
						<th>Type</th>
						<th>Plugin</th>
						<th>Health</th>
						<th>Message</th>
						<th>Capabilities</th>
						<th>Last Checked</th>
					</tr>
				</thead>
				<tbody>
					{#each data.items as component}
						<tr onclick={() => goto(`/services/${component.namespace}--${component.name}`)}>
							<td class="td-name">
								<span class="name">{component.name}</span>
								<span class="ns">{component.namespace}</span>
							</td>
							<td><span class="type-tag">{typeLabel(component.type)}</span></td>
							<td class="td-mono">{component.plugin}</td>
							<td><span class="tag tag-{healthColor(component.health)}">{component.health}</span></td>
							<td class="td-msg">{component.message}</td>
							<td class="td-caps">
								{#each (component.capabilities ?? []).slice(0, 3) as cap}
									<span class="cap">{cap}</span>
								{/each}
								{#if (component.capabilities ?? []).length > 3}
									<span class="cap-more">+{component.capabilities.length - 3}</span>
								{/if}
							</td>
							<td class="td-mono td-small">{fmtDate(component.last_checked)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
</div>

<style>
	.main { padding: 20px 24px; overflow-y: auto; background: var(--bg-base); height: 100%; }

	.p-header {
		display: flex; justify-content: space-between; align-items: center;
		margin-bottom: 20px; padding-bottom: 12px; border-bottom: 1px solid var(--border-main);
	}
	.p-title { font-family: 'Inter Tight', sans-serif; font-size: 22px; font-weight: 800; letter-spacing: -0.02em; color: var(--text-1); }
	.p-actions { display: flex; align-items: center; gap: 12px; }
	.p-meta { font-family: 'JetBrains Mono', monospace; font-size: 10px; color: var(--text-3); }
	.p-btn {
		font-family: 'JetBrains Mono', monospace; font-size: 11px; font-weight: 600;
		padding: 6px 12px; background: var(--bg-surface); border: 1px solid var(--border-main);
		color: var(--text-2); cursor: pointer;
	}
	.p-btn:hover:not(:disabled) { background: var(--bg-hover); color: var(--text-1); }
	.p-btn:disabled { opacity: 0.5; cursor: wait; }

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

	.grid-4 {
		display: grid; grid-template-columns: repeat(4, 1fr);
		gap: 1px; background: var(--border-main); border: 1px solid var(--border-main);
		margin-bottom: 20px;
	}
	.m-box { background: var(--bg-surface); padding: 14px 16px; display: flex; flex-direction: column; gap: 6px; min-height: 100px; }
	.m-lbl { font-family: 'JetBrains Mono', monospace; font-size: 10px; font-weight: 600; letter-spacing: 0.14em; text-transform: uppercase; color: var(--text-3); }
	.m-val { font-family: 'Inter Tight', sans-serif; font-size: 28px; font-weight: 800; letter-spacing: -0.02em; line-height: 1; color: var(--text-1); }
	.m-val.green { color: var(--green); }
	.m-val.red { color: var(--red); }
	.m-sub { font-size: 11px; color: var(--text-3); font-family: 'JetBrains Mono', monospace; }
	.p-track { height: 3px; background: var(--bg-inset); border: 1px solid var(--border-sub); margin-top: auto; }
	.p-fill { height: 100%; background: var(--text-1); transition: width 0.3s ease; }
	.p-fill.green { background: var(--green); }
	.p-fill.red { background: var(--red); }

	.board { border: 1px solid var(--border-main); background: var(--bg-surface); }
	.b-head {
		padding: 0 14px; height: 36px; border-bottom: 1px solid var(--border-main);
		display: flex; justify-content: space-between; align-items: center;
		font-family: 'JetBrains Mono', monospace; font-size: 11px; font-weight: 700;
		letter-spacing: 0.12em; text-transform: uppercase; background: var(--bg-hover); color: var(--text-1);
	}
	.b-head .meta { font-weight: 400; color: var(--text-3); }
	.empty-state { padding: 40px; text-align: center; color: var(--text-3); font-family: 'JetBrains Mono', monospace; font-size: 12px; }

	table { width: 100%; border-collapse: collapse; font-size: 13px; }
	th {
		text-align: left; padding: 8px 12px;
		font-family: 'JetBrains Mono', monospace; font-size: 10px; font-weight: 700;
		letter-spacing: 0.14em; text-transform: uppercase; color: var(--text-3);
		border-bottom: 1px solid var(--border-main); background: var(--bg-inset);
	}
	td { padding: 10px 12px; border-bottom: 1px solid var(--border-sub); color: var(--text-2); vertical-align: middle; }
	tr:last-child td { border-bottom: none; }
	tr:hover td { background: var(--bg-hover); cursor: pointer; }

	.td-name { display: flex; flex-direction: column; gap: 2px; }
	.name { color: var(--text-1); font-weight: 600; }
	.ns { font-family: 'JetBrains Mono', monospace; font-size: 10px; color: var(--text-3); }

	.type-tag {
		display: inline-flex; align-items: center;
		font-family: 'JetBrains Mono', monospace; font-size: 10px; font-weight: 600;
		text-transform: uppercase; padding: 2px 6px;
		border: 1px solid var(--border-main); color: var(--text-2); background: var(--bg-inset);
	}

	.td-mono { font-family: 'JetBrains Mono', monospace; font-size: 12px; }
	.td-small { font-size: 11px; color: var(--text-3); }
	.td-msg { font-size: 12px; color: var(--text-2); max-width: 300px; }

	.tag {
		display: inline-flex; align-items: center;
		font-family: 'JetBrains Mono', monospace; font-size: 10px; font-weight: 700;
		text-transform: uppercase; padding: 2px 6px; border: 1px solid;
	}
	.tag-ok { border-color: var(--green); color: var(--green); }
	.tag-err { border-color: var(--red); color: var(--red); }
	.tag-warn { border-color: var(--yellow); color: var(--yellow); }

	.td-caps { display: flex; flex-wrap: wrap; gap: 4px; }
	.cap {
		font-family: 'JetBrains Mono', monospace; font-size: 9px; font-weight: 600;
		padding: 1px 5px; border: 1px solid var(--border-main);
		color: var(--text-3); text-transform: uppercase; letter-spacing: 0.05em;
	}
	.cap-more { font-family: 'JetBrains Mono', monospace; font-size: 9px; color: var(--text-3); padding: 1px 4px; }
</style>