<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { SvelteURLSearchParams } from 'svelte/reactivity';

	interface LiveJob {
		job_id: string;
		job_name: string;
		job_group: string;
		state: string;
		message: string;
		details: Record<string, string>;
	}

	interface JobsResponse {
		jobs: LiveJob[];
		total: number;
	}

	let data = $state<JobsResponse>({ jobs: [], total: 0 });
	let loading = $state(true);
	let error = $state('');
	let lastUpdated = $state<Date | null>(null);
	let stateFilter = $state('');
	let jobNameFilter = $state('');

	async function load() {
		loading = true;
		error = '';
		try {
			const params = new SvelteURLSearchParams();
			if (stateFilter) params.set('state', stateFilter);
			if (jobNameFilter) params.set('job_name', jobNameFilter);

			let url = '/api/v1/jobs';
			const qs = params.toString();
			if (qs) url += '?' + qs;

			const res = await fetch(url, {
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

	async function cancelJob(componentName: string, jobId: string) {
		try {
			const res = await fetch(`/api/v1/jobs/${jobId}`, { method: 'DELETE' });
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			load();
		} catch (e: unknown) {
			alert('Failed to cancel: ' + (e instanceof Error ? e.message : String(e)));
		}
	}

	let interval: ReturnType<typeof setInterval>;
	onMount(() => { load(); interval = setInterval(load, 5000); });
	onDestroy(() => { if (interval) clearInterval(interval); });

	function stateColor(s: string) {
		switch (s) {
			case 'RUNNING': return 'ok';
			case 'PENDING': return 'warn';
			case 'SUCCEEDED': return 'ok';
			case 'FAILED': return 'err';
			case 'CANCELED': return 'warn';
			default: return 'warn';
		}
	}

	function stateLabel(s: string) {
		switch (s) {
			case 'JOB_STATE_PENDING': return 'pending';
			case 'JOB_STATE_RUNNING': return 'running';
			case 'JOB_STATE_SUCCEEDED': return 'succeeded';
			case 'JOB_STATE_FAILED': return 'failed';
			case 'JOB_STATE_CANCELED': return 'canceled';
			default: return s.toLowerCase();
		}
	}

	function fmtTime(iso: string) {
		if (!iso) return '—';
		return new Date(iso).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
	}

	const running = $derived(data.jobs.filter(j => j.state === 'JOB_STATE_RUNNING').length);
	const failed = $derived(data.jobs.filter(j => j.state === 'JOB_STATE_FAILED').length);
	const succeeded = $derived(data.jobs.filter(j => j.state === 'JOB_STATE_SUCCEEDED').length);
	const pending = $derived(data.jobs.filter(j => j.state === 'JOB_STATE_PENDING').length);

	const grouped = $derived(Object.entries(
		data.jobs.reduce((acc: Record<string, LiveJob[]>, j) => {
			const key = j.job_name || j.job_id;
			if (!acc[key]) acc[key] = [];
			acc[key].push(j);
			return acc;
		}, {} as Record<string, LiveJob[]>)
	).sort(([a], [b]) => a.localeCompare(b)));

	function componentName(j: LiveJob) {
		return j.details?._component ?? '—';
	}

	function pluginName(j: LiveJob) {
		return j.details?._plugin ?? '—';
	}
</script>

<svelte:head>
	<title>Cosmonaut · Jobs</title>
</svelte:head>

<div class="main">
	<div class="p-header">
		<div class="p-title">Jobs</div>
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

	<div class="grid-4">
		<div class="m-box">
			<div class="m-lbl">Total Jobs</div>
			<div class="m-val">{data.total}</div>
			<div class="m-sub">across all engines</div>
			<div class="p-track"><div class="p-fill" style="width:100%"></div></div>
		</div>
		<div class="m-box">
			<div class="m-lbl">Running</div>
			<div class="m-val green">{running}</div>
			<div class="m-sub">{pending} pending</div>
			<div class="p-track"><div class="p-fill green" style="width:{data.total > 0 ? (running/data.total)*100 : 0}%"></div></div>
		</div>
		<div class="m-box">
			<div class="m-lbl">Failed</div>
			<div class="m-val {failed > 0 ? 'red' : ''}">{failed}</div>
			<div class="m-sub">{succeeded} succeeded</div>
			<div class="p-track"><div class="p-fill red" style="width:{data.total > 0 ? (failed/data.total)*100 : 0}%"></div></div>
		</div>
		<div class="m-box">
			<div class="m-lbl">Job Names</div>
			<div class="m-val">{grouped.length}</div>
			<div class="m-sub">unique pipelines</div>
			<div class="p-track"><div class="p-fill" style="width:100%"></div></div>
		</div>
	</div>

	<div class="filters">
		<select class="filter-select" bind:value={stateFilter} onchange={load}>
			<option value="">All states</option>
			<option value="running">Running</option>
			<option value="pending">Pending</option>
			<option value="succeeded">Succeeded</option>
			<option value="failed">Failed</option>
			<option value="canceled">Canceled</option>
		</select>
		<input
			class="filter-input"
			type="text"
			placeholder="Filter by job name..."
			bind:value={jobNameFilter}
			oninput={load}
		/>
	</div>

	<div class="board">
		<div class="b-head">
			Jobs
			<span class="meta">{data.total} execution{data.total !== 1 ? 's' : ''}</span>
		</div>

		{#if loading && data.jobs.length === 0}
			<div class="empty-state">Loading...</div>
		{:else if data.jobs.length === 0}
			<div class="empty-state">No jobs found. Submit a job or check your workload plugins.</div>
		{:else}
			{#each grouped as [jobName, executions]}
				{@const latestExec = executions[0]}
				{@const runningCount = executions.filter(e => e.state === 'JOB_STATE_RUNNING').length}
				{@const failedCount = executions.filter(e => e.state === 'JOB_STATE_FAILED').length}

				<div class="job-group">
					<button class="group-header" onclick={(e) => {
						const target = e.currentTarget;
						target.classList.toggle('collapsed');
					}}>
						<span class="group-name">{jobName}</span>
						<span class="group-meta">
							<span class="group-engine">{pluginName(latestExec)} · {componentName(latestExec)}</span>
							<span class="group-count">{executions.length} executions</span>
							{#if runningCount > 0}
								<span class="tag tag-ok">{runningCount} running</span>
							{/if}
							{#if failedCount > 0}
								<span class="tag tag-err">{failedCount} failed</span>
							{/if}
						</span>
						<span class="group-arrow">▾</span>
					</button>

					<div class="group-body">
						<table>
							<thead>
								<tr>
									<th>Execution ID</th>
									<th>Engine</th>
									<th>Component</th>
									<th>Status</th>
									<th>Message</th>
									<th>Actions</th>
								</tr>
							</thead>
							<tbody>
								{#each executions as job}
									<tr>
										<td class="td-mono" title={job.job_id}>
											{job.job_id.length > 16 ? job.job_id.slice(0, 16) + '...' : job.job_id}
										</td>
										<td><span class="engine-tag">{pluginName(job)}</span></td>
										<td class="td-mono">{componentName(job)}</td>
										<td><span class="tag tag-{stateColor(job.state)}">{stateLabel(job.state)}</span></td>
										<td class="td-msg">{job.message}</td>
										<td>
											<div class="actions">
												{#if job.details?._component}
													<button
														class="action-btn"
														onclick={() => goto(`/services/cosmonaut--${job.details._component}`)}
														title="View component"
													>
														📋
													</button>
												{/if}
												{#if job.state === 'JOB_STATE_RUNNING' || job.state === 'JOB_STATE_PENDING'}
													<button
														class="action-btn danger"
														onclick={() => cancelJob(componentName(job), job.job_id)}
														title="Cancel"
													>
														⏹
													</button>
												{/if}
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/each}
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

	.filters { display: flex; gap: 12px; margin-bottom: 16px; }
	.filter-select {
		font-family: 'JetBrains Mono', monospace; font-size: 11px;
		padding: 6px 10px; background: var(--bg-surface);
		border: 1px solid var(--border-main); color: var(--text-2); cursor: pointer;
	}
	.filter-input {
		font-family: 'JetBrains Mono', monospace; font-size: 11px;
		padding: 6px 10px; background: var(--bg-surface);
		border: 1px solid var(--border-main); color: var(--text-2);
		outline: none; flex: 1; max-width: 300px;
	}
	.filter-input:focus { border-color: var(--text-3); }

	.board { border: 1px solid var(--border-main); background: var(--bg-surface); }
	.b-head {
		padding: 0 14px; height: 36px; border-bottom: 1px solid var(--border-main);
		display: flex; justify-content: space-between; align-items: center;
		font-family: 'JetBrains Mono', monospace; font-size: 11px; font-weight: 700;
		letter-spacing: 0.12em; text-transform: uppercase; background: var(--bg-hover); color: var(--text-1);
	}
	.b-head .meta { font-weight: 400; color: var(--text-3); }
	.empty-state { padding: 40px; text-align: center; color: var(--text-3); font-family: 'JetBrains Mono', monospace; font-size: 12px; }

	.job-group { border-bottom: 1px solid var(--border-main); }
	.job-group:last-child { border-bottom: none; }

	.group-header {
		width: 100%; display: flex; align-items: center; gap: 12px;
		padding: 10px 14px; background: var(--bg-inset); border: none;
		cursor: pointer; text-align: left; font-family: inherit;
		color: var(--text-1); transition: background 0.1s;
	}
	.group-header:hover { background: var(--bg-hover); }
	.group-header.collapsed + .group-body { display: none; }
	.group-header.collapsed .group-arrow { transform: rotate(-90deg); }

	.group-name { font-family: 'Inter Tight', sans-serif; font-size: 14px; font-weight: 700; letter-spacing: -0.01em; }
	.group-meta { display: flex; align-items: center; gap: 8px; flex: 1; }
	.group-engine { font-family: 'JetBrains Mono', monospace; font-size: 10px; color: var(--text-3); }
	.group-count { font-family: 'JetBrains Mono', monospace; font-size: 10px; color: var(--text-3); }
	.group-arrow {
		font-size: 12px; color: var(--text-3); transition: transform 0.15s;
		margin-left: auto;
	}

	.group-body { }
	.group-body table { width: 100%; border-collapse: collapse; font-size: 13px; }
	.group-body th {
		text-align: left; padding: 6px 12px;
		font-family: 'JetBrains Mono', monospace; font-size: 9px; font-weight: 700;
		letter-spacing: 0.14em; text-transform: uppercase; color: var(--text-3);
		border-bottom: 1px solid var(--border-sub); background: var(--bg-base);
	}
	.group-body td { padding: 8px 12px; border-bottom: 1px solid var(--border-sub); color: var(--text-2); vertical-align: middle; }
	.group-body tr:last-child td { border-bottom: none; }
	.group-body tr:hover td { background: var(--bg-hover); cursor: default; }

	.td-mono { font-family: 'JetBrains Mono', monospace; font-size: 12px; }
	.td-msg { font-size: 12px; color: var(--text-3); max-width: 250px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

	.tag {
		display: inline-flex; align-items: center;
		font-family: 'JetBrains Mono', monospace; font-size: 10px; font-weight: 700;
		text-transform: uppercase; padding: 2px 6px; border: 1px solid;
	}
	.tag-ok { border-color: var(--green); color: var(--green); }
	.tag-err { border-color: var(--red); color: var(--red); }
	.tag-warn { border-color: var(--yellow); color: var(--yellow); }

	.engine-tag {
		display: inline-flex; align-items: center;
		font-family: 'JetBrains Mono', monospace; font-size: 10px; font-weight: 600;
		text-transform: uppercase; padding: 2px 6px;
		border: 1px solid var(--border-main); color: var(--text-2);
	}

	.actions { display: flex; gap: 4px; }
	.action-btn {
		background: none; border: 1px solid var(--border-sub);
		padding: 2px 6px; cursor: pointer; font-size: 12px; line-height: 1;
	}
	.action-btn:hover { background: var(--bg-hover); border-color: var(--border-main); }
	.action-btn.danger:hover { background: var(--red); border-color: var(--red); color: #fff; }
</style>