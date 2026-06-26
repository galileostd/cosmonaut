<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { theme } from '$lib/stores/theme.svelte';

	interface Node {
		name: string;
		role: string;
		cpuUsed: number;
		cpuAlloc: number;
		memoryUsed: number;
		memoryAlloc: number;
		status: string;
		statusMessage: string;
	}

	interface Metrics {
		activeServices: number;
		interactiveJobs: number;
		failedJobs: number;
		clusterResources: number;
	}

	interface ClusterData {
		metrics: Metrics;
		nodes: Node[];
	}

	let clusterData = $state<ClusterData>({ 
		metrics: { activeServices: 0, interactiveJobs: 0, failedJobs: 0, clusterResources: 0 }, 
		nodes: [] 
	});
	let loading = $state(true);
	let error = $state('');
	let lastUpdated = $state<Date | null>(null);

	async function loadClusterData() {
		loading = true;
		error = '';
		try {
			const res = await fetch('/api/v1/cluster', {
				cache: 'no-store',
				headers: { 'Cache-Control': 'no-cache' }
			});
			if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch cluster data`);
			clusterData = await res.json();
			lastUpdated = new Date();
		} catch (e: unknown) {
			if (e instanceof Error) {
				error = e.message;
			} else {
				error = String(e);
			}
		} finally {
			loading = false;
		}
	}

	let interval: ReturnType<typeof setInterval>;

	onMount(() => {
		document.documentElement.setAttribute('data-theme', theme.current);
		loadClusterData();
		// Auto-refresh a cada 5 segundos
		interval = setInterval(loadClusterData, 5000);
	});

	onDestroy(() => {
		if (interval) clearInterval(interval);
	});
</script>

<svelte:head>
	<title>Cosmonaut · Platform</title>
</svelte:head>

<div class="main">
	<div class="p-header">
		<div class="p-title">Cluster Overview</div>
		<div class="p-actions">
			{#if lastUpdated}
				<span class="p-meta">Updated {lastUpdated.toLocaleTimeString()}</span>
			{/if}
			<button class="p-btn" onclick={loadClusterData} disabled={loading}>
				{loading ? '⟳ Refreshing...' : '↻ Refresh'}
			</button>
		</div>
	</div>

	{#if error}
		<div class="error-banner">
			<span class="error-icon">⚠</span>
			{error}
			<button class="retry-btn" onclick={loadClusterData}>Retry</button>
		</div>
	{/if}

	{#if loading && clusterData.nodes.length === 0}
		<div class="loading">Loading cluster data...</div>
	{:else}
		<div class="grid-4">
			<div class="m-box">
				<div class="m-lbl">Active Services</div>
				<div class="m-val">{clusterData.metrics.activeServices}</div>
				<div class="m-sub">healthy components</div>
				<div class="p-track">
					<div class="p-fill" style="width:{Math.min((clusterData.metrics.activeServices / 20) * 100, 100)}%"></div>
				</div>
			</div>
			<div class="m-box">
				<div class="m-lbl">Interactive Jobs</div>
				<div class="m-val">{clusterData.metrics.interactiveJobs}</div>
				<div class="m-sub">running now</div>
				<div class="p-track">
					<div class="p-fill" style="width:{Math.min((clusterData.metrics.interactiveJobs / 10) * 100, 100)}%"></div>
				</div>
			</div>
			<div class="m-box">
				<div class="m-lbl">Failed Jobs (24h)</div>
				<div class="m-val red">{clusterData.metrics.failedJobs}</div>
				<div class="m-sub">
					{clusterData.metrics.failedJobs > 0 
						? `${((clusterData.metrics.failedJobs / (clusterData.metrics.interactiveJobs + clusterData.metrics.failedJobs || 1)) * 100).toFixed(1)}% failure rate` 
						: '0% failure rate'}
				</div>
				<div class="p-track">
					<div class="p-fill red" style="width:{Math.min((clusterData.metrics.failedJobs / 20) * 100, 100)}%"></div>
				</div>
			</div>
			<div class="m-box">
				<div class="m-lbl">Cluster Resources</div>
				<div class="m-val">{Math.round(clusterData.metrics.clusterResources)}%</div>
				<div class="m-sub">CPU · Memory</div>
				<div class="p-track">
					<div class="p-fill" style="width:{Math.min(clusterData.metrics.clusterResources, 100)}%"></div>
				</div>
			</div>
		</div>

		<div class="board">
			<div class="b-head">
				Cluster Nodes <span class="meta">{clusterData.nodes.length} node{clusterData.nodes.length !== 1 ? 's' : ''}</span>
			</div>
			{#if clusterData.nodes.length === 0}
				<div class="empty-state">No nodes found</div>
			{:else}
				<table>
					<thead>
						<tr>
							<th>Name</th>
							<th>Role</th>
							<th>CPU (used/alloc)</th>
							<th>Memory (used/alloc)</th>
							<th>Status</th>
							<th>Usage</th>
						</tr>
					</thead>
					<tbody>
						{#each clusterData.nodes as node}
							{@const cpuPct = node.cpuAlloc > 0 ? (node.cpuUsed / node.cpuAlloc) * 100 : 0}
							{@const memPct = node.memoryAlloc > 0 ? (node.memoryUsed / node.memoryAlloc) * 100 : 0}
							{@const avgPct = (cpuPct + memPct) / 2}
							<tr>
								<td class="td-t">{node.name}</td>
								<td class="td-mono"><span class="role-tag role-{node.role}">{node.role}</span></td>
								<td class="td-mono">
									{node.cpuUsed.toFixed(2)} / {node.cpuAlloc.toFixed(1)} cores
									<span class="pct">({cpuPct.toFixed(1)}%)</span>
								</td>
								<td class="td-mono">
									{node.memoryUsed.toFixed(2)} / {node.memoryAlloc.toFixed(2)} GB
									<span class="pct">({memPct.toFixed(1)}%)</span>
								</td>
								<td><span class="tag tag-{node.status}">{node.statusMessage}</span></td>
								<td>
									<div class="mini-track">
										<div class="mini-fill" style="width:{Math.min(avgPct, 100)}%"></div>
									</div>
									<span class="mini-pct">{avgPct.toFixed(0)}%</span>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>
	{/if}
</div>

<style>
	:global([data-theme="dark"]) {
		--bg-base: #0B0B0A;
		--bg-surface: #141413;
		--bg-inset: #0E0E0D;
		--bg-hover: #1E1E1C;
		--border-main: #2A2A28;
		--border-sub: #1E1E1C;
		--text-1: #ECE8DC;
		--text-2: #A8A49A;
		--text-3: #6A655A;
		--red: #C8102E;
		--red-deep: #8B0A1F;
		--green: #2F6B3A;
		--yellow: #B8762B;
		--header-bg: #050504;
		--accent: #ECE8DC;
	}

	:global([data-theme="light"]) {
		--bg-base: #F4F1E8;
		--bg-surface: #ECE8DC;
		--bg-inset: #FFFFFE;
		--bg-hover: #E0DDD0;
		--border-main: #1414141A;
		--border-sub: #14141410;
		--text-1: #141414;
		--text-2: #5A5A56;
		--text-3: #8A857A;
		--red: #C8102E;
		--red-deep: #8B0A1F;
		--green: #2F6B3A;
		--yellow: #B8762B;
		--header-bg: #E0DDD0;
		--accent: #141414;
	}

	.main {
		padding: 20px 24px;
		overflow-y: auto;
		background: var(--bg-base);
		height: 100%;
	}

	.p-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 20px;
		padding-bottom: 12px;
		border-bottom: 1px solid var(--border-main);
	}

	.p-title {
		font-family: 'Inter Tight', sans-serif;
		font-size: 22px;
		font-weight: 800;
		letter-spacing: -0.02em;
		color: var(--text-1);
	}

	.p-actions {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.p-meta {
		font-family: 'JetBrains Mono', monospace;
		font-size: 10px;
		color: var(--text-3);
	}

	.p-btn {
		font-family: 'JetBrains Mono', monospace;
		font-size: 11px;
		font-weight: 600;
		padding: 6px 12px;
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-2);
		cursor: pointer;
		transition: all 0.15s;
	}

	.p-btn:hover:not(:disabled) {
		background: var(--bg-hover);
		color: var(--text-1);
	}

	.p-btn:disabled {
		opacity: 0.5;
		cursor: wait;
	}

	.loading, .empty-state {
		padding: 40px;
		text-align: center;
		color: var(--text-3);
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
	}

	.error-banner {
		padding: 12px 16px;
		margin-bottom: 20px;
		background: var(--red-deep);
		border: 1px solid var(--red);
		color: var(--text-1);
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.error-icon {
		font-size: 14px;
	}

	.retry-btn {
		margin-left: auto;
		font-family: 'JetBrains Mono', monospace;
		font-size: 10px;
		padding: 4px 10px;
		background: var(--red);
		border: none;
		color: #fff;
		cursor: pointer;
		text-transform: uppercase;
		font-weight: 700;
	}

	.retry-btn:hover {
		background: var(--text-1);
		color: var(--red);
	}

	.grid-4 {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 1px;
		background: var(--border-main);
		border: 1px solid var(--border-main);
		margin-bottom: 20px;
	}

	.m-box {
		background: var(--bg-surface);
		padding: 14px 16px;
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-height: 100px;
	}

	.m-lbl {
		font-family: 'JetBrains Mono', monospace;
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--text-3);
	}

	.m-val {
		font-family: 'Inter Tight', sans-serif;
		font-size: 28px;
		font-weight: 800;
		letter-spacing: -0.02em;
		line-height: 1;
		font-variant-numeric: tabular-nums;
		color: var(--text-1);
	}

	.m-val.red {
		color: var(--red);
	}

	.m-sub {
		font-size: 11px;
		color: var(--text-3);
		font-family: 'JetBrains Mono', monospace;
	}

	.p-track {
		height: 3px;
		background: var(--bg-inset);
		border: 1px solid var(--border-sub);
		margin-top: auto;
	}

	.p-fill {
		height: 100%;
		background: var(--text-1);
		transition: width 0.3s ease;
	}

	.p-fill.red {
		background: var(--red);
	}

	.board {
		border: 1px solid var(--border-main);
		margin-bottom: 20px;
		background: var(--bg-surface);
	}

	.b-head {
		padding: 0 14px;
		height: 36px;
		border-bottom: 1px solid var(--border-main);
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-family: 'JetBrains Mono', monospace;
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		background: var(--bg-hover);
		color: var(--text-1);
	}

	.b-head .meta {
		font-weight: 400;
		color: var(--text-3);
		font-size: 11px;
	}

	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 13px;
	}

	th {
		text-align: left;
		padding: 8px 12px;
		font-family: 'JetBrains Mono', monospace;
		font-size: 10px;
		font-weight: 700;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--text-3);
		border-bottom: 1px solid var(--border-main);
		background: var(--bg-inset);
	}

	td {
		padding: 9px 12px;
		border-bottom: 1px solid var(--border-sub);
		color: var(--text-2);
		vertical-align: middle;
	}

	tr:last-child td {
		border-bottom: none;
	}

	tr:hover td {
		background: var(--bg-hover);
		cursor: pointer;
	}

	.td-t {
		color: var(--text-1);
		font-weight: 600;
	}

	.td-mono {
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
	}

	.pct {
		color: var(--text-3);
		font-size: 10px;
		margin-left: 4px;
	}

	.role-tag {
		display: inline-flex;
		align-items: center;
		font-family: 'JetBrains Mono', monospace;
		font-size: 10px;
		font-weight: 700;
		text-transform: uppercase;
		padding: 1px 6px;
		border: 1px solid var(--border-main);
	}

	.role-master {
		background: var(--yellow);
		color: var(--bg-base);
		border-color: var(--yellow);
	}

	.role-worker {
		background: transparent;
		color: var(--text-2);
	}

	.tag {
		display: inline-flex;
		align-items: center;
		font-family: 'JetBrains Mono', monospace;
		font-size: 10.5px;
		font-weight: 700;
		text-transform: uppercase;
		padding: 2px 6px;
		border: 1px solid;
	}

	.tag-ok {
		border-color: var(--green);
		color: var(--green);
	}

	.tag-err {
		border-color: var(--red);
		color: var(--red);
	}

	.tag-warn {
		border-color: var(--yellow);
		color: var(--yellow);
	}

	.mini-track {
		width: 60px;
		height: 4px;
		background: var(--bg-inset);
		border: 1px solid var(--border-sub);
		display: inline-block;
		vertical-align: middle;
		margin-right: 6px;
	}

	.mini-fill {
		height: 100%;
		background: var(--accent);
		transition: width 0.3s ease;
	}

	.mini-pct {
		font-family: 'JetBrains Mono', monospace;
		font-size: 10px;
		color: var(--text-3);
	}

	@media (max-width: 1024px) {
		.grid-4 {
			grid-template-columns: repeat(2, 1fr);
		}
		.p-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 10px;
		}
	}
</style>