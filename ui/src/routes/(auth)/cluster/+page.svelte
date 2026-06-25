<script lang="ts">
	import { onMount } from 'svelte';
	import { theme } from '$lib/stores/theme.svelte';

	onMount(() => {
		// Aplica o tema atual
		document.documentElement.setAttribute('data-theme', theme.current);
	});
</script>

<svelte:head>
	<title>Cosmonaut · Platform</title>
</svelte:head>

<div class="main">
	<div class="p-title">Cluster Overview</div>

	<div class="grid-4">
		<div class="m-box">
			<div class="m-lbl">Active Services</div>
			<div class="m-val">12</div>
			<div class="m-sub">4 busy · 8 ready</div>
			<div class="p-track"><div class="p-fill" style="width:33%"></div></div>
		</div>
		<div class="m-box">
			<div class="m-lbl">Interactive Jobs</div>
			<div class="m-val">7</div>
			<div class="m-sub">2 busy · 5 ready</div>
			<div class="p-track"><div class="p-fill" style="width:28%"></div></div>
		</div>
		<div class="m-box">
			<div class="m-lbl">Failed Jobs (24h)</div>
			<div class="m-val red">1</div>
			<div class="m-sub">0.8% failure rate</div>
			<div class="p-track"><div class="p-fill red" style="width:1%"></div></div>
		</div>
		<div class="m-box">
			<div class="m-lbl">Cluster Resources</div>
			<div class="m-val">17%</div>
			<div class="m-sub">CPU 1.2% · Mem 17%</div>
			<div class="p-track"><div class="p-fill" style="width:17%"></div></div>
		</div>
	</div>

	<div class="board">
		<div class="b-head">Cluster Nodes <span class="meta">3 nodes · 1 master · 2 workers</span></div>
		<table>
			<thead>
				<tr>
					<th>Name</th>
					<th>Role</th>
					<th>CPU (used/alloc)</th>
					<th>Memory (used/alloc)</th>
					<th>Status</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td class="td-t">node-01</td>
					<td class="td-mono">master</td>
					<td class="td-mono">0.8 / 4.0 cores</td>
					<td class="td-mono">3.2 / 16.0 GB</td>
					<td><span class="tag tag-ok">Ready</span></td>
				</tr>
				<tr>
					<td class="td-t">node-02</td>
					<td class="td-mono">worker</td>
					<td class="td-mono">1.2 / 8.0 cores</td>
					<td class="td-mono">7.8 / 32.0 GB</td>
					<td><span class="tag tag-ok">Ready</span></td>
				</tr>
				<tr>
					<td class="td-t">node-03</td>
					<td class="td-mono">worker</td>
					<td class="td-mono">0.4 / 8.0 cores</td>
					<td class="td-mono">4.1 / 32.0 GB</td>
					<td><span class="tag tag-warn">SchedulingDisabled</span></td>
				</tr>
			</tbody>
		</table>
	</div>
</div>

<style>
	/* ── TEMA ────────────────────────────────────────────────────── */
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
	}

	.main {
		padding: 20px 24px;
		overflow-y: auto;
		background: var(--bg-base);
		height: 100%;
	}

	.p-title {
		font-family: 'Inter Tight', sans-serif;
		font-size: 22px;
		font-weight: 800;
		letter-spacing: -0.02em;
		margin-bottom: 20px;
		padding-bottom: 12px;
		border-bottom: 1px solid var(--border-main);
		color: var(--text-1);
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

	@media (max-width: 1024px) {
		.grid-4 {
			grid-template-columns: repeat(2, 1fr);
		}
	}
</style>