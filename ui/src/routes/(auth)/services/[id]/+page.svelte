<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { theme } from '$lib/stores/theme.svelte';

	let activeTab = $state('jobs');

	onMount(() => {
		document.documentElement.setAttribute('data-theme', theme.current);
	});

	const tabs = [
		{ id: 'jobs', label: 'Jobs', badge: '0' },
		{ id: 'requests', label: 'Requests', badge: '0' },
		{ id: 'parameters', label: 'Parameters', badge: '1' },
		{ id: 'tags', label: 'Tags', badge: '1' },
		{ id: 'memory', label: 'Memory' },
		{ id: 'execute', label: 'Execute Job' },
		{ id: 'audit', label: 'User audit' }
	];

	function goBack() {
		window.history.back();
	}
</script>

<svelte:head>
	<title>Cosmonaut · Services</title>
</svelte:head>

<div class="main">
	<!-- Breadcrumb -->
	<div class="breadcrumb">
		<span class="back-link" onclick={goBack}>
			<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M10 3l-5 5 5 5" stroke="currentColor" stroke-width="1.5"/></svg>
			Back to List
		</span>
		<span class="sep">/</span>
		<span class="crumb">
			<span class="crumb-tag">cluster: default</span>
			<svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M6 3l5 5-5 5" stroke="currentColor" stroke-width="1.5"/></svg>
			<span class="crumb-tag">service: JOB Based Service - Interactive Spark Job</span>
		</span>
	</div>

	<!-- Service Header -->
	<div class="service-header">
		<div class="service-icon">
			<svg width="20" height="20" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg">
				<path d="M834.867 272.111l-322.867 186.781-322.867-186.781 299.444-171.104c5.195-2.969 10.761-4.87 16.466-5.752 10.111-1.577 20.64 0.139 30.334 5.752zM490.015 1020.395c6.54 3.525 14.054 5.519 21.985 5.519 8.024 0 15.584-2.040 22.171-5.613 16.512-2.691 32.514-8.349 47.171-16.698l324.721-185.575c28.664-16.558 49.165-41.605 60.25-70.222 6.169-15.863 9.37-32.792 9.509-50.279v-371.057c0-22.171-5.241-43.182-14.564-61.873-1.021-3.293-2.412-6.493-4.221-9.601-2.134-3.665-4.639-6.911-7.514-9.74-1.438-1.995-2.969-3.989-4.499-5.891-10.668-13.266-23.747-24.49-38.544-33.164l-325.093-185.761c-28.294-16.327-60.297-21.568-90.631-16.883-16.837 2.644-33.164 8.302-48.098 16.837l-324.721 185.575c-17.764 10.25-32.421 23.794-43.645 39.425-2.783 2.783-5.288 6.030-7.375 9.601-1.809 3.154-3.247 6.401-4.267 9.74-1.856 3.71-3.525 7.561-5.009 11.41-6.122 15.909-9.323 32.839-9.462 50v371.382c0.047 32.931 11.502 63.219 30.798 87.106 10.668 13.266 23.747 24.49 38.544 33.164l325.093 185.761c15.028 8.674 31.076 14.24 47.402 16.79zM465.617 539.226v370.64l-301.252-172.124c-5.148-3.014-9.601-6.865-13.219-11.363-6.447-7.978-10.204-17.903-10.204-28.85v-346.104zM558.383 909.867v-370.64l324.676-187.801v346.104c-0.047 5.613-1.16 11.41-3.247 16.79-3.71 9.554-10.436 17.811-19.759 23.191z"/>
			</svg>
		</div>
		<div class="service-main">
			<div class="name">
				JOB Based Service - Interactive Spark Job
				<span class="badge-paused">Paused</span>
			</div>
			<div class="id">20260624-2120-seqs3</div>
			<div class="service-attrs">
				<div class="attr">
					<span class="lbl">Type</span>
					<span class="val">Job <svg class="copy" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></span>
				</div>
				<div class="attr">
					<span class="lbl">Language</span>
					<span class="val">Python <svg class="copy" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></span>
				</div>
				<div class="attr">
					<span class="lbl">Created</span>
					<span class="val">2026/06/24 23:20:08 <svg class="copy" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></span>
				</div>
				<div class="attr">
					<span class="lbl">Description</span>
					<span class="val" style="max-width:240px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">
						Read me Please! JOB-Based Interactive Service Example...
						<svg class="copy" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg>
					</span>
				</div>
			</div>
		</div>
		<div class="service-actions">
			<button class="btn-action"><svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M3 10l5-6 5 6" stroke="currentColor" stroke-width="1.5"/></svg> Service Actions</button>
		</div>
	</div>

	<!-- Tabs -->
	<div class="tabs">
		{#each tabs as tab}
			<button
				class="tab {activeTab === tab.id ? 'active' : ''}"
				onclick={() => activeTab = tab.id}
			>
				{tab.label}
				{#if tab.badge}<span class="tab-badge">{tab.badge}</span>{/if}
			</button>
		{/each}
	</div>

	<!-- TAB: JOBS -->
	{#if activeTab === 'jobs'}
		<div class="tab-content active">
			<div class="toolbar">
				<div class="search-wrapper">
					<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="7" cy="7" r="5" stroke="currentColor" stroke-width="1.5"/><path d="M11 11l3 3" stroke="currentColor" stroke-width="1.5"/></svg>
					<input type="text" placeholder="Search jobs" />
				</div>
				<button class="btn-add"><svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M8 3v10M3 8h10" stroke="currentColor" stroke-width="2"/></svg> New Job</button>
			</div>
			<div class="board">
				<div class="b-head">Jobs <span class="meta">0 total</span></div>
				<table>
					<thead>
						<tr>
							<th style="width:20px;"><input type="checkbox" /></th>
							<th>Name</th><th>Id</th><th>Type</th><th>Requests</th>
							<th>Status</th><th>Submit Time</th><th>Start Time</th>
							<th>End Time</th><th>Duration</th>
						</tr>
					</thead>
					<tbody>
						<tr>
							<td colspan="10" class="empty-state">
								<div>
									<svg width="40" height="40" viewBox="0 0 40 40" fill="none"><circle cx="20" cy="20" r="16" stroke="var(--text-3)" stroke-width="1.5"/><path d="M20 10v10l5 5" stroke="var(--text-3)" stroke-width="1.5"/></svg>
									<div class="empty-title">No Data</div>
									<div class="empty-sub">There is no items for selected list</div>
								</div>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
			<div class="pagination">
				<button class="btn-action" disabled>‹</button>
				<span class="page-info">1 / 1</span>
				<button class="btn-action" disabled>›</button>
				<span class="page-size">25 / page</span>
			</div>
		</div>
	{/if}

	<!-- TAB: REQUESTS -->
	{#if activeTab === 'requests'}
		<div class="tab-content active">
			<div class="toolbar">
				<div class="search-wrapper">
					<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="7" cy="7" r="5" stroke="currentColor" stroke-width="1.5"/><path d="M11 11l3 3" stroke="currentColor" stroke-width="1.5"/></svg>
					<input type="text" placeholder="Search requests" />
				</div>
			</div>
			<div class="board">
				<div class="b-head">Requests <span class="meta">0 total</span></div>
				<table>
					<thead>
						<tr>
							<th style="width:20px;"><input type="checkbox" /></th>
							<th>Id</th><th>Class</th><th>Start Time</th>
							<th>End Time</th><th>Duration</th><th>Result</th>
						</tr>
					</thead>
					<tbody>
						<tr>
							<td colspan="7" class="empty-state">
								<div>
									<svg width="40" height="40" viewBox="0 0 40 40" fill="none"><circle cx="20" cy="20" r="16" stroke="var(--text-3)" stroke-width="1.5"/><path d="M20 10v10l5 5" stroke="var(--text-3)" stroke-width="1.5"/></svg>
									<div class="empty-title">No Data</div>
									<div class="empty-sub">There is no items for selected list</div>
								</div>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
			<div class="pagination">
				<button class="btn-action" disabled>‹</button>
				<span class="page-info">1 / 1</span>
				<button class="btn-action" disabled>›</button>
				<span class="page-size">25 / page</span>
			</div>
		</div>
	{/if}

	<!-- TAB: PARAMETERS -->
	{#if activeTab === 'parameters'}
		<div class="tab-content active">
			<div class="board">
				<div class="b-head">Parameters <span class="meta">1</span></div>
				<table class="kv-table">
					<thead><tr><th>Key</th><th>Value</th></tr></thead>
					<tbody>
						<tr>
							<td>spark.submit.pyFiles <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td>
							<td>s3a://ilum-files/ilum/default/job_service_example.py <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
	{/if}

	<!-- TAB: TAGS -->
	{#if activeTab === 'tags'}
		<div class="tab-content active">
			<div class="board">
				<div class="b-head">Tags <span class="meta">1</span></div>
				<table class="kv-table">
					<thead><tr><th>Key</th><th>Value</th></tr></thead>
					<tbody>
						<tr>
							<td>ExampleServiceTag <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td>
							<td>IlumJobServiceExample <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
	{/if}

	<!-- TAB: MEMORY -->
	{#if activeTab === 'memory'}
		<div class="tab-content active">
			<div class="board">
				<div class="b-head">Memory Configuration</div>
				<table class="kv-table">
					<thead><tr><th>Key</th><th>Value</th></tr></thead>
					<tbody>
						<tr><td>executors <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td><td>1 <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td></tr>
						<tr><td>driverCores <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td><td>- <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td></tr>
						<tr><td>executorCores <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td><td>- <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td></tr>
						<tr><td>driverMemory <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td><td>- <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td></tr>
						<tr><td>executorMemory <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td><td>- <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td></tr>
						<tr><td>dynamicAllocation <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td><td>false <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td></tr>
						<tr><td>initialExecutors <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td><td>- <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td></tr>
						<tr><td>minExecutors <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td><td>- <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td></tr>
						<tr><td>maxExecutors <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td><td>- <svg class="copy-btn" width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg></td></tr>
					</tbody>
				</table>
			</div>
		</div>
	{/if}

	<!-- TAB: EXECUTE JOB -->
	{#if activeTab === 'execute'}
		<div class="tab-content active">
			<div class="execute-layout">
				<div class="execute-panel">
					<div class="endpoint-bar">
						<span class="method">post</span>
						<span class="url">http://ilum-core:9888/api/v1/group/20260624-2120-seqs3/job/execute</span>
						<svg class="copy" width="14" height="14" viewBox="0 0 16 16" fill="none"><rect x="5.5" y="5.5" width="9" height="9" stroke="currentColor" stroke-width="1.5"/><path d="M2 10.5V2h8.5" stroke="currentColor" stroke-width="1.5"/></svg>
					</div>
					<button class="btn-execute">Execute</button>
					<div class="field" style="margin-top:14px;">
						<span class="lbl">Class</span>
						<input type="text" placeholder="File and class name (e.g., load_data_to_bronze.RawDataToBronze)" />
					</div>
					<div class="field">
						<span class="lbl">Parameters</span>
						<textarea>{`{}`}</textarea>
					</div>
				</div>
				<div class="execute-panel">
					<div class="field">
						<span class="lbl">End Time</span>
						<input type="text" readonly value="" />
					</div>
					<div class="field">
						<span class="lbl">Duration</span>
						<input type="text" readonly value="" />
					</div>
					<div class="field">
						<span class="lbl">Result</span>
						<textarea readonly></textarea>
					</div>
				</div>
			</div>
		</div>
	{/if}

	<!-- TAB: USER AUDIT -->
	{#if activeTab === 'audit'}
		<div class="tab-content active">
			<div class="audit-card">
				<div class="audit-head">
					<h3>Audit &amp; ownership</h3>
					<button class="btn-action"><svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="10" cy="5" r="2.5" stroke="currentColor" stroke-width="1.5"/><path d="M5 14c0-2.5 2-4.5 5-4.5s5 2 5 4.5" stroke="currentColor" stroke-width="1.5"/><circle cx="4" cy="4" r="1.5" stroke="currentColor" stroke-width="1.5"/></svg> Transfer ownership</button>
				</div>
				<dl class="audit-grid">
					<dt>Owner</dt><dd>system</dd>
					<dt>Created by</dt><dd>system</dd>
					<dt>Created at</dt><dd>6/24/2026, 11:20:08 PM</dd>
					<dt>Last modified by</dt><dd>system</dd>
					<dt>Last modified at</dt><dd>6/24/2026, 11:20:08 PM</dd>
				</dl>
			</div>
		</div>
	{/if}
</div>

<style>
	/* TEMA */
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
		--orange: #E66F00;
		--orange-deep: #B85900;
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
		--orange: #E66F00;
		--orange-deep: #B85900;
	}

	.main {
		padding: 20px 24px;
		overflow-y: auto;
		background: var(--bg-base);
		height: 100%;
	}

	/* Breadcrumb */
	.breadcrumb {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 16px;
		font-size: 12px;
		color: var(--text-3);
	}
	.back-link {
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 4px;
		color: var(--text-2);
	}
	.back-link:hover { color: var(--text-1); }
	.sep { color: var(--text-3); }
	.crumb { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
	.crumb-tag {
		font-family: 'JetBrains Mono', monospace;
		font-size: 11px;
		background: var(--bg-hover);
		padding: 2px 8px;
		border: 1px solid var(--border-main);
	}

	/* Service Header */
	.service-header {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		padding: 16px 20px;
		margin-bottom: 20px;
		display: flex;
		flex-wrap: wrap;
		gap: 16px 24px;
		align-items: flex-start;
	}
	.service-icon {
		width: 32px;
		min-width: 32px;
		height: 32px;
		background: var(--orange);
		border-radius: 2px;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.service-icon svg { fill: white; }
	.service-main {
		flex: 1;
		min-width: 200px;
	}
	.service-main .name {
		font-family: 'Inter Tight', sans-serif;
		font-size: 16px;
		font-weight: 700;
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
	}
	.service-main .id {
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
		color: var(--text-3);
		margin-top: 2px;
	}
	.service-attrs {
		display: flex;
		flex-wrap: wrap;
		gap: 16px 24px;
		margin-top: 6px;
	}
	.service-attrs .attr {
		display: flex;
		flex-direction: column;
		gap: 1px;
	}
	.service-attrs .attr .lbl {
		font-family: 'JetBrains Mono', monospace;
		font-size: 9px;
		font-weight: 600;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--text-3);
	}
	.service-attrs .attr .val {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 13px;
	}
	.service-attrs .attr .val .copy {
		opacity: 0;
		width: 14px;
		height: 14px;
		cursor: pointer;
		color: var(--text-3);
		transition: opacity 0.15s;
	}
	.service-attrs .attr:hover .copy { opacity: 1; }
	.service-attrs .attr .val .copy:hover { color: var(--text-1); }
	.service-actions {
		margin-left: auto;
		display: flex;
		gap: 6px;
		align-items: center;
	}
	.badge-paused {
		font-family: 'JetBrains Mono', monospace;
		font-size: 10px;
		font-weight: 700;
		text-transform: uppercase;
		padding: 1px 10px;
		border: 1px solid var(--border-main);
		background: var(--bg-hover);
		color: var(--text-2);
	}
	.badge-running {
		font-family: 'JetBrains Mono', monospace;
		font-size: 10px;
		font-weight: 700;
		text-transform: uppercase;
		padding: 1px 10px;
		border: 1px solid var(--green);
		color: var(--green);
	}

	/* Tabs */
	.tabs {
		display: flex;
		gap: 4px;
		border-bottom: 1px solid var(--border-main);
		margin-bottom: 16px;
		padding: 0 2px;
	}
	.tab {
		padding: 8px 14px;
		font-family: 'JetBrains Mono', monospace;
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--text-3);
		cursor: pointer;
		border-bottom: 2px solid transparent;
		transition: color 0.15s, border-color 0.15s;
		display: flex;
		align-items: center;
		gap: 6px;
		background: transparent;
		border-top: none;
		border-left: none;
		border-right: none;
	}
	.tab:hover { color: var(--text-1); }
	.tab.active {
		color: var(--text-1);
		border-bottom-color: var(--red);
	}
	.tab-badge {
		font-family: 'JetBrains Mono', monospace;
		font-size: 9px;
		font-weight: 700;
		padding: 1px 4px;
		background: var(--bg-hover);
		color: var(--text-3);
		border-radius: 2px;
	}
	.tab.active .tab-badge { background: var(--border-main); color: var(--text-1); }

	/* Toolbar */
	.toolbar {
		display: flex;
		gap: 10px;
		align-items: center;
		margin-bottom: 12px;
	}
	.search-wrapper {
		position: relative;
		flex: 1;
		max-width: 220px;
	}
	.search-wrapper svg {
		position: absolute;
		left: 8px;
		top: 50%;
		transform: translateY(-50%);
		color: var(--text-3);
	}
	.search-wrapper input {
		width: 100%;
		background: var(--bg-inset);
		border: 1px solid var(--border-main);
		padding: 5px 8px 5px 28px;
		color: var(--text-1);
		font-family: 'Inter', sans-serif;
		font-size: 12px;
		outline: none;
	}
	.search-wrapper input:focus { border-color: var(--text-1); }

	/* Board */
	.board {
		border: 1px solid var(--border-main);
		margin-bottom: 12px;
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
	tr:last-child td { border-bottom: none; }
	tr:hover td { background: var(--bg-hover); cursor: pointer; }

	.empty-state {
		text-align: center;
		padding: 40px 0;
		color: var(--text-3);
		font-family: 'JetBrains Mono', monospace;
		font-size: 13px;
	}
	.empty-state div { display: flex; flex-direction: column; align-items: center; gap: 6px; }
	.empty-title { font-weight: 700; color: var(--text-2); }
	.empty-sub { font-size: 12px; }

	/* Pagination */
	.pagination {
		display: flex;
		align-items: center;
		gap: 12px;
		justify-content: flex-end;
		font-size: 12px;
		color: var(--text-3);
	}
	.page-info { font-family: 'JetBrains Mono', monospace; }
	.page-size { font-family: 'JetBrains Mono', monospace; font-size: 11px; }

	/* Buttons */
	.btn-action {
		background: transparent;
		border: 1px solid var(--border-main);
		color: var(--text-2);
		padding: 4px 12px;
		font-family: 'Inter', sans-serif;
		font-size: 11px;
		font-weight: 600;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		gap: 5px;
	}
	.btn-action:hover:not(:disabled) { background: var(--bg-hover); color: var(--text-1); border-color: var(--text-1); }
	.btn-action:disabled { opacity: 0.4; cursor: not-allowed; }

	.btn-add {
		background: transparent;
		border: 1px solid var(--border-main);
		color: var(--text-2);
		padding: 5px 14px;
		font-family: 'Inter', sans-serif;
		font-size: 12px;
		font-weight: 600;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		margin-bottom: 0;
	}
	.btn-add:hover { background: var(--bg-hover); color: var(--text-1); border-color: var(--text-1); }

	/* KV Table */
	.kv-table td:first-child {
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
		color: var(--text-1);
		font-weight: 500;
		width: 40%;
	}
	.kv-table td:last-child {
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
		color: var(--text-2);
	}
	.kv-table .copy-btn {
		opacity: 0;
		width: 14px;
		height: 14px;
		cursor: pointer;
		color: var(--text-3);
		transition: opacity 0.15s;
		margin-left: 8px;
	}
	.kv-table tr:hover .copy-btn { opacity: 1; }
	.kv-table .copy-btn:hover { color: var(--text-1); }

	/* Execute Job Layout */
	.execute-layout {
		display: flex;
		gap: 1px;
		background: var(--border-main);
		border: 1px solid var(--border-main);
		min-height: 400px;
	}
	.execute-panel {
		background: var(--bg-surface);
		padding: 16px 20px;
		flex: 1;
		min-width: 0;
	}
	.execute-panel:first-child { flex: 2; }
	.execute-panel:last-child { flex: 1; }
	.execute-panel .field { margin-bottom: 14px; }
	.execute-panel .field .lbl {
		font-family: 'JetBrains Mono', monospace;
		font-size: 10px;
		font-weight: 700;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--text-3);
		display: block;
		margin-bottom: 2px;
	}
	.execute-panel .field input,
	.execute-panel .field textarea {
		width: 100%;
		background: var(--bg-inset);
		border: 1px solid var(--border-main);
		padding: 6px 10px;
		color: var(--text-1);
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
		outline: none;
		resize: vertical;
	}
	.execute-panel .field input:focus,
	.execute-panel .field textarea:focus {
		border-color: var(--text-1);
	}
	.execute-panel .field textarea { min-height: 120px; }
	.execute-panel .field input[readonly] { color: var(--text-3); cursor: default; }
	.execute-panel .field textarea[readonly] { color: var(--text-3); cursor: default; }

	.endpoint-bar {
		display: flex;
		align-items: center;
		gap: 8px;
		background: var(--bg-inset);
		border: 1px solid var(--border-main);
		padding: 4px 12px;
		margin-bottom: 14px;
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
	}
	.endpoint-bar .method {
		color: var(--green);
		font-weight: 700;
		text-transform: uppercase;
		font-size: 10px;
	}
	.endpoint-bar .url {
		flex: 1;
		color: var(--text-2);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.endpoint-bar .copy {
		cursor: pointer;
		color: var(--text-3);
		flex-shrink: 0;
	}
	.endpoint-bar .copy:hover { color: var(--text-1); }

	.btn-execute {
		background: var(--orange);
		border: none;
		color: #F4F1E8;
		padding: 6px 20px;
		font-family: 'Inter Tight', sans-serif;
		font-size: 13px;
		font-weight: 700;
		cursor: pointer;
		text-transform: uppercase;
		letter-spacing: 0.08em;
	}
	.btn-execute:hover { background: var(--orange-deep); }

	/* Audit Card */
	.audit-card {
		border: 1px solid var(--border-main);
		background: var(--bg-surface);
		padding: 16px 20px;
	}
	.audit-head {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 12px;
	}
	.audit-head h3 {
		font-family: 'Inter Tight', sans-serif;
		font-size: 14px;
		font-weight: 700;
	}
	.audit-grid {
		display: grid;
		grid-template-columns: max-content 1fr;
		gap: 6px 24px;
		font-size: 13px;
	}
	.audit-grid dt {
		color: var(--text-3);
		font-family: 'JetBrains Mono', monospace;
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}
	.audit-grid dd {
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
		color: var(--text-2);
	}

	@media (max-width: 1024px) {
		.execute-layout { flex-direction: column; }
	}
</style>