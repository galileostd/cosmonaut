<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';

	interface LiveJob {
		job_id: string;
		job_name?: string | null;
		job_group?: string | null;
		state: number | string;  // ← API returns numbers!
		message?: string | null;
		details?: Record<string, string>;
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
	let jobNameDebounce: ReturnType<typeof setTimeout>;

	async function load() {
		loading = true;
		error = '';
		try {
			const params = new URLSearchParams();
			if (stateFilter) params.set('state', stateFilter);
			if (jobNameFilter) params.set('job_name', jobNameFilter);

			const qs = params.toString();
			const url = '/api/v1/jobs' + (qs ? '?' + qs : '');

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

	function onJobNameInput() {
		clearTimeout(jobNameDebounce);
		jobNameDebounce = setTimeout(load, 400);
	}

	async function cancelJob(jobId: string) {
		if (!confirm(`Cancel job ${jobId}? This will delete the workload immediately.`)) return;
		try {
			const res = await fetch(`/api/v1/jobs/${jobId}`, { method: 'DELETE' });
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			await load();
		} catch (e: unknown) {
			alert('Failed to cancel: ' + (e instanceof Error ? e.message : String(e)));
		}
	}

	let interval: ReturnType<typeof setInterval>;
	onMount(() => { load(); interval = setInterval(load, 5000); });
	onDestroy(() => {
		clearInterval(interval);
		clearTimeout(jobNameDebounce);
	});

	// proto enum values — API returns NUMBERS
	const STATE_UNSPECIFIED = 0;
	const STATE_PENDING     = 1;
	const STATE_RUNNING     = 2;
	const STATE_SUCCEEDED   = 3;
	const STATE_FAILED      = 4;
	const STATE_CANCELED    = 5;
	const STATE_UNKNOWN     = 6;

	function stateColor(s: number | string) {
		const n = typeof s === 'string' ? parseStateNumber(s) : s;
		switch (n) {
			case STATE_RUNNING:   return 'running';
			case STATE_PENDING:   return 'warn';
			case STATE_SUCCEEDED: return 'ok';
			case STATE_FAILED:    return 'err';
			case STATE_CANCELED:  return 'canceled';
			default:              return 'warn';
		}
	}

	function stateLabel(s: number | string) {
		const n = typeof s === 'string' ? parseStateNumber(s) : s;
		switch (n) {
			case STATE_PENDING:   return 'pending';
			case STATE_RUNNING:   return 'running';
			case STATE_SUCCEEDED: return 'succeeded';
			case STATE_FAILED:    return 'failed';
			case STATE_CANCELED:  return 'canceled';
			case STATE_UNKNOWN:   return 'unknown';
			default:              return String(n);
		}
	}

	function parseStateNumber(s: string): number {
		const n = parseInt(s, 10);
		return isNaN(n) ? 6 : n; // default to unknown
	}

	function isActive(s: number | string) {
		const n = typeof s === 'string' ? parseStateNumber(s) : s;
		return n === STATE_RUNNING || n === STATE_PENDING;
	}

	function componentOf(j: LiveJob) {
		return j.details?._component ?? j.details?.component ?? '—';
	}
	function pluginOf(j: LiveJob) {
		return j.details?._plugin ?? j.details?.plugin ?? '—';
	}

	// stats
	const running   = $derived(data.jobs.filter(j => {
		const n = typeof j.state === 'string' ? parseStateNumber(j.state) : j.state;
		return n === STATE_RUNNING;
	}).length);
	const failed    = $derived(data.jobs.filter(j => {
		const n = typeof j.state === 'string' ? parseStateNumber(j.state) : j.state;
		return n === STATE_FAILED;
	}).length);
	const succeeded = $derived(data.jobs.filter(j => {
		const n = typeof j.state === 'string' ? parseStateNumber(j.state) : j.state;
		return n === STATE_SUCCEEDED;
	}).length);
	const pending   = $derived(data.jobs.filter(j => {
		const n = typeof j.state === 'string' ? parseStateNumber(j.state) : j.state;
		return n === STATE_PENDING;
	}).length);

	// group by job_name (fallback to job_id), sort groups alphabetically
	// within each group sort: running first, then pending, then rest by insertion order
	const grouped = $derived(
		Object.entries(
			data.jobs.reduce((acc: Record<string, LiveJob[]>, j) => {
				const key = j.job_name || j.job_id || 'unnamed';
				if (!acc[key]) acc[key] = [];
				acc[key].push(j);
				return acc;
			}, {})
		)
		.sort(([a], [b]) => a.localeCompare(b))
		.map(([name, jobs]) => ({
			name,
			jobs: [...jobs].sort((a, b) => {
				const rank = (s: number | string) => {
					const n = typeof s === 'string' ? parseStateNumber(s) : s;
					return n === STATE_RUNNING ? 0 : n === STATE_PENDING ? 1 : 2;
				};
				return rank(a.state) - rank(b.state);
			}),
			runningCount:  jobs.filter(j => {
				const n = typeof j.state === 'string' ? parseStateNumber(j.state) : j.state;
				return n === STATE_RUNNING;
			}).length,
			failedCount:   jobs.filter(j => {
				const n = typeof j.state === 'string' ? parseStateNumber(j.state) : j.state;
				return n === STATE_FAILED;
			}).length,
			pendingCount:  jobs.filter(j => {
				const n = typeof j.state === 'string' ? parseStateNumber(j.state) : j.state;
				return n === STATE_PENDING;
			}).length,
			// use the first running job for engine display, else first job
			rep: jobs.find(j => {
				const n = typeof j.state === 'string' ? parseStateNumber(j.state) : j.state;
				return n === STATE_RUNNING;
			}) ?? jobs[0],
		}))
	);

	// expand/collapse state per group
	let collapsed = $state<Record<string, boolean>>({});
	function toggleGroup(name: string) {
		collapsed[name] = !collapsed[name];
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
				<span class="p-meta">Updated {lastUpdated.toLocaleTimeString()}</span>
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

	<!-- stats -->
	<div class="grid-4">
		<div class="m-box">
			<div class="m-lbl">Total</div>
			<div class="m-val">{data.total}</div>
			<div class="m-sub">across all engines</div>
			<div class="p-track"><div class="p-fill" style="width:100%"></div></div>
		</div>
		<div class="m-box">
			<div class="m-lbl">Running</div>
			<div class="m-val green">{running}</div>
			<div class="m-sub">{pending} pending</div>
			<div class="p-track">
				<div class="p-fill green"
					style="width:{data.total > 0 ? (running / data.total) * 100 : 0}%">
				</div>
			</div>
		</div>
		<div class="m-box">
			<div class="m-lbl">Failed</div>
			<div class="m-val {failed > 0 ? 'red' : ''}">{failed}</div>
			<div class="m-sub">{succeeded} succeeded</div>
			<div class="p-track">
				<div class="p-fill red"
					style="width:{data.total > 0 ? (failed / data.total) * 100 : 0}%">
				</div>
			</div>
		</div>
		<div class="m-box">
			<div class="m-lbl">Pipelines</div>
			<div class="m-val">{grouped.length}</div>
			<div class="m-sub">unique job names</div>
			<div class="p-track"><div class="p-fill" style="width:100%"></div></div>
		</div>
	</div>

	<!-- filters -->
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
			oninput={onJobNameInput}
		/>
	</div>

	<!-- list -->
	<div class="board">
		<div class="b-head">
			Live Jobs
			<span class="meta">{data.total} execution{data.total !== 1 ? 's' : ''}</span>
		</div>

		{#if loading && data.jobs.length === 0}
			<div class="empty-state">Loading...</div>
		{:else if data.jobs.length === 0}
			<div class="empty-state">
				No jobs found. Submit a Spark job or check your workload plugins.
			</div>
		{:else}
			{#each grouped as group}
				<div class="job-group">
					<button class="group-header" onclick={() => toggleGroup(group.name)}>
						<span class="group-arrow" class:rotated={collapsed[group.name]}>▾</span>
						<span class="group-name">{group.name}</span>
						<span class="group-meta">
							<span class="engine-tag">{pluginOf(group.rep)}</span>
							<span class="component-tag">{componentOf(group.rep)}</span>
							<span class="count-tag">{group.jobs.length} run{group.jobs.length !== 1 ? 's' : ''}</span>
							{#if group.runningCount > 0}
								<span class="tag tag-running">{group.runningCount} running</span>
							{/if}
							{#if group.failedCount > 0}
								<span class="tag tag-err">{group.failedCount} failed</span>
							{/if}
							{#if group.pendingCount > 0}
								<span class="tag tag-warn">{group.pendingCount} pending</span>
							{/if}
						</span>
					</button>

					{#if !collapsed[group.name]}
						<div class="group-body">
							<table>
								<thead>
									<tr>
										<th>Execution ID</th>
										<th>Group</th>
										<th>Component</th>
										<th>Status</th>
										<th>Message</th>
										<th>Actions</th>
									</tr>
								</thead>
								<tbody>
									{#each group.jobs as job (job.job_id)}
										<tr class:row-active={isActive(job.state)}>
											<td class="td-mono" title={job.job_id}>
												{job.job_id.length > 20
													? job.job_id.slice(0, 20) + '…'
													: job.job_id}
											</td>
											<td class="td-mono td-muted">
												{job.job_group || '—'}
											</td>
											<td class="td-mono">{componentOf(job)}</td>
											<td>
												<span class="tag tag-{stateColor(job.state)}">
													{stateLabel(job.state)}
												</span>
											</td>
											<td class="td-msg" title={job.message || ''}>
												{job.message || '—'}
											</td>
											<td>
												<div class="actions">
													{#if componentOf(job) !== '—'}
														<button
															class="action-btn"
															onclick={() => goto(`/services/cosmonaut--${componentOf(job)}`)}
															title="View component"
														>
															<svg width="12" height="12" viewBox="0 0 16 16" fill="none">
																<path d="M2 4h7l3 3v5H2z" stroke="currentColor" stroke-width="1.5"/>
																<path d="M9 4v3h3" stroke="currentColor" stroke-width="1.5"/>
															</svg>
														</button>
													{/if}
													{#if isActive(job.state)}
														<button
															class="action-btn danger"
															onclick={() => cancelJob(job.job_id)}
															title="Cancel job"
														>
															<svg width="12" height="12" viewBox="0 0 16 16" fill="none">
																<rect x="3" y="3" width="10" height="10" stroke="currentColor" stroke-width="1.5"/>
															</svg>
														</button>
													{/if}
												</div>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				</div>
			{/each}
		{/if}
	</div>
</div>

<style>
	.main { padding: 20px 24px; overflow-y: auto; background: var(--bg-base); height: 100%; }

	/* ── header ── */
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

	/* ── error ── */
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

	/* ── stats ── */
	.grid-4 {
		display: grid; grid-template-columns: repeat(4, 1fr);
		gap: 1px; background: var(--border-main); border: 1px solid var(--border-main);
		margin-bottom: 20px;
	}
	.m-box { background: var(--bg-surface); padding: 14px 16px; display: flex; flex-direction: column; gap: 6px; min-height: 100px; }
	.m-lbl { font-family: 'JetBrains Mono', monospace; font-size: 10px; font-weight: 600; letter-spacing: 0.14em; text-transform: uppercase; color: var(--text-3); }
	.m-val { font-family: 'Inter Tight', sans-serif; font-size: 28px; font-weight: 800; letter-spacing: -0.02em; line-height: 1; color: var(--text-1); font-variant-numeric: tabular-nums; }
	.m-val.green { color: var(--green); }
	.m-val.red { color: var(--red); }
	.m-sub { font-size: 11px; color: var(--text-3); font-family: 'JetBrains Mono', monospace; }
	.p-track { height: 3px; background: var(--bg-inset); border: 1px solid var(--border-sub); margin-top: auto; }
	.p-fill { height: 100%; background: var(--text-1); transition: width 0.3s ease; }
	.p-fill.green { background: var(--green); }
	.p-fill.red { background: var(--red); }

	/* ── filters ── */
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
		outline: none; flex: 1; max-width: 320px;
	}
	.filter-input::placeholder { color: var(--text-3); }
	.filter-input:focus { border-color: var(--text-3); }

	/* ── board ── */
	.board { border: 1px solid var(--border-main); background: var(--bg-surface); }
	.b-head {
		padding: 0 14px; height: 36px; border-bottom: 1px solid var(--border-main);
		display: flex; justify-content: space-between; align-items: center;
		font-family: 'JetBrains Mono', monospace; font-size: 11px; font-weight: 700;
		letter-spacing: 0.12em; text-transform: uppercase; background: var(--bg-hover); color: var(--text-1);
	}
	.b-head .meta { font-weight: 400; color: var(--text-3); }
	.empty-state { padding: 40px; text-align: center; color: var(--text-3); font-family: 'JetBrains Mono', monospace; font-size: 12px; }

	/* ── job groups ── */
	.job-group { border-bottom: 1px solid var(--border-main); }
	.job-group:last-child { border-bottom: none; }

	.group-header {
		width: 100%; display: flex; align-items: center; gap: 10px;
		padding: 9px 14px; background: var(--bg-inset);
		border: none; cursor: pointer; text-align: left;
		color: var(--text-1); transition: background 0.1s;
	}
	.group-header:hover { background: var(--bg-hover); }

	.group-arrow {
		font-size: 11px; color: var(--text-3);
		transition: transform 0.15s; flex-shrink: 0;
	}
	.group-arrow.rotated { transform: rotate(-90deg); }

	.group-name {
		font-family: 'Inter Tight', sans-serif; font-size: 13px; font-weight: 700;
		letter-spacing: -0.01em; flex-shrink: 0;
	}
	.group-meta { display: flex; align-items: center; gap: 8px; flex: 1; flex-wrap: wrap; }

	.engine-tag, .component-tag, .count-tag {
		font-family: 'JetBrains Mono', monospace; font-size: 10px; font-weight: 600;
		text-transform: uppercase; padding: 1px 5px;
		border: 1px solid var(--border-main); color: var(--text-3);
	}

	/* ── table ── */
	.group-body table { width: 100%; border-collapse: collapse; font-size: 13px; }
	.group-body th {
		text-align: left; padding: 5px 12px;
		font-family: 'JetBrains Mono', monospace; font-size: 9px; font-weight: 700;
		letter-spacing: 0.14em; text-transform: uppercase; color: var(--text-3);
		border-bottom: 1px solid var(--border-sub); background: var(--bg-base);
	}
	.group-body td {
		padding: 8px 12px; border-bottom: 1px solid var(--border-sub);
		color: var(--text-2); vertical-align: middle;
	}
	.group-body tr:last-child td { border-bottom: none; }
	.group-body tr:hover td { background: var(--bg-hover); }
	.row-active td { border-left: 2px solid var(--green); }
	.row-active:first-child td:first-child { padding-left: 10px; }

	.td-mono { font-family: 'JetBrains Mono', monospace; font-size: 12px; }
	.td-muted { color: var(--text-3); }
	.td-msg { font-size: 12px; color: var(--text-3); max-width: 240px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

	/* ── tags ── */
	.tag {
		display: inline-flex; align-items: center;
		font-family: 'JetBrains Mono', monospace; font-size: 10px; font-weight: 700;
		text-transform: uppercase; padding: 2px 6px; border: 1px solid;
	}
	.tag-ok       { border-color: var(--green);  color: var(--green); }
	.tag-running  { border-color: var(--green);  color: var(--green); }
	.tag-err      { border-color: var(--red);    color: var(--red); }
	.tag-warn     { border-color: var(--yellow); color: var(--yellow); }
	.tag-canceled { border-color: var(--text-3); color: var(--text-3); }

	/* ── actions ── */
	.actions { display: flex; gap: 4px; }
	.action-btn {
		display: inline-flex; align-items: center; justify-content: center;
		width: 24px; height: 24px;
		background: none; border: 1px solid var(--border-sub);
		color: var(--text-3); cursor: pointer;
	}
	.action-btn:hover { background: var(--bg-hover); border-color: var(--border-main); color: var(--text-1); }
	.action-btn.danger:hover { background: var(--red); border-color: var(--red); color: #fff; }

	@media (max-width: 1024px) {
		.grid-4 { grid-template-columns: repeat(2, 1fr); }
	}
</style>