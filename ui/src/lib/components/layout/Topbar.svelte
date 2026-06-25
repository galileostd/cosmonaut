<script lang="ts">
	import { theme } from '$lib/stores/theme.svelte';
	import * as m from '$lib/paraglide/messages';
</script>

<header class="topbar">
	<div class="t-search">
		<svg class="t-search-ico" width="14" height="14" viewBox="0 0 16 16" fill="none">
			<circle cx="7" cy="7" r="5" stroke="currentColor" stroke-width="1.5"/>
			<path d="M11 11l3 3" stroke="currentColor" stroke-width="1.5"/>
		</svg>
		<input type="text" placeholder={m.topbar_search_placeholder()} />
		<span class="t-search-kbd">{m.topbar_command_hint()}</span>
	</div>
	<div class="t-tools">
		<button class="t-btn" aria-label="Notifications">
			<svg width="16" height="16" viewBox="0 0 16 16" fill="none">
				<path d="M3 6a5 5 0 0 1 10 0v3l1 2H2l1-2zM6 13a2 2 0 0 0 4 0" stroke="currentColor" stroke-width="1.5"/>
			</svg>
			<span class="t-btn-dot"></span>
		</button>
		<button class="t-btn" onclick={() => theme.toggle()} aria-label="Toggle theme">
			<svg width="14" height="14" viewBox="0 0 16 16" fill="none">
				<path d="M13 8.5A5 5 0 1 1 7.5 3a4 4 0 0 0 5.5 5.5z" stroke="currentColor" stroke-width="1.5"/>
			</svg>
		</button>
		<span class="t-ver">{m.app_version()}</span>
		<div class="t-user">D</div>
	</div>
</header>

<style>
	.topbar {
		grid-column: 2;
		grid-row: 1;
		background: var(--bg-soft);
		border-bottom: 1px solid var(--line);
		display: flex;
		align-items: center;
		padding: 0 16px;
		gap: 16px;
		min-width: 0;
	}
	.t-search {
		flex: 1;
		max-width: 520px;
		position: relative;
	}
	.t-search input {
		width: 100%;
		height: 32px;
		padding: 0 80px 0 32px;
		background: var(--bg-inset);
		border: 1px solid var(--line-2);
		color: var(--fg);
		font-family: var(--font-sans);
		font-size: 13px;
		outline: none;
	}
	.t-search input:focus { border-color: var(--line-bold); }
	.t-search input::placeholder { color: var(--fg-mute); }
	.t-search-ico {
		position: absolute; left: 10px; top: 50%;
		transform: translateY(-50%);
		color: var(--fg-mute); pointer-events: none;
	}
	.t-search-kbd {
		position: absolute; right: 8px; top: 50%;
		transform: translateY(-50%);
		font-family: var(--font-mono); font-size: 10px;
		color: var(--fg-mute); background: var(--bg-base);
		border: 1px solid var(--line-2); padding: 2px 6px;
	}
	.t-tools { margin-left: auto; display: flex; align-items: center; gap: 6px; }
	.t-btn {
		height: 32px; min-width: 32px; padding: 0 8px;
		display: inline-flex; align-items: center; justify-content: center;
		background: transparent; border: 1px solid transparent;
		color: var(--fg-soft); cursor: pointer; position: relative;
	}
	.t-btn:hover { background: var(--bg-sunk); border-color: var(--line-2); }
	.t-btn-dot {
		position: absolute; top: 6px; right: 6px;
		width: 6px; height: 6px; background: var(--red); border-radius: 50%;
	}
	.t-ver { font-family: var(--font-mono); font-size: 11px; color: var(--fg-mute); padding: 0 4px; }
	.t-user {
		width: 28px; height: 28px; background: var(--red); color: #F4F1E8;
		display: inline-flex; align-items: center; justify-content: center;
		font-weight: 700; font-size: 12px;
	}
</style>