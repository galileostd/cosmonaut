<script lang="ts">
	import { page } from '$app/state';
	import Logo from '$lib/components/ui/Logo.svelte';
	import * as m from '$lib/paraglide/messages';
	import { sidebar } from '$lib/stores/sidebar.svelte';

	const isActive = (path: string) =>
		page.url.pathname === path || page.url.pathname.startsWith(path + '/');

	const navGroups = [
		{
			label: 'Control plane',
			items: [
				{ href: '/cluster',         label: () => m.nav_cluster(),         icon: 'grid' },
				{ href: '/services',        label: () => m.nav_services(),        icon: 'box',     badge: '4' },
				{ href: '/schedules',       label: () => m.nav_schedules(),       icon: 'clock' },
				{ href: '/jobs',            label: () => m.nav_jobs(),            icon: 'layers',  badge: '12' },
				{ href: '/requests',        label: () => m.nav_requests(),        icon: 'list' },
				{ href: '/alerts',          label: () => m.nav_alerts(),          icon: 'bell',    badge: '2' },
				{ href: '/events',          label: () => m.nav_events(),          icon: 'feed' },
			]
		},
		{
			label: 'Data',
			items: [
				{ href: '/tables',  label: () => m.nav_tables(),   icon: 'database' },
				{ href: '/sql',     label: () => m.nav_sql_lab(),  icon: 'sql' },
				{ href: '/lineage', label: () => m.nav_lineage(),  icon: 'lineage' },
			]
		},
		{
			label: 'Platform',
			items: [
				{ href: '/plugin-registry', label: () => m.nav_plugin_registry(), icon: 'plugin' },
				{ href: '/pods',            label: () => m.nav_pods(),            icon: 'pod' },
				{ href: '/storage',         label: () => m.nav_storage(),         icon: 'storage' },
				{ href: '/catalogs',        label: () => m.nav_catalogs(),        icon: 'catalog' },
				{ href: '/presets',         label: () => m.nav_presets(),         icon: 'preset' },
			]
		},
		{
			label: 'Observability',
			items: [
				{ href: '/grafana', label: () => m.nav_grafana(), icon: 'chart' },
				{ href: '/mlflow',  label: () => m.nav_mlflow(),  icon: 'ml' },
				{ href: '/modules', label: () => m.nav_modules(), icon: 'module' },
			]
		},
		{
			label: 'Security',
			items: [
				{ href: '/users',     label: () => m.nav_users(),     icon: 'user' },
				{ href: '/groups',    label: () => m.nav_groups(),    icon: 'group' },
				{ href: '/roles',     label: () => m.nav_roles(),     icon: 'shield' },
				{ href: '/event-log', label: () => m.nav_event_log(), icon: 'log' },
			]
		}
	];

	const icons: Record<string, string> = {
		grid:     '<rect x="1" y="1" width="6" height="6" stroke="currentColor" stroke-width="1.5"/><rect x="9" y="1" width="6" height="6" stroke="currentColor" stroke-width="1.5"/><rect x="1" y="9" width="6" height="6" stroke="currentColor" stroke-width="1.5"/><rect x="9" y="9" width="6" height="6" stroke="currentColor" stroke-width="1.5"/>',
		box:      '<path d="M2 4l6-3 6 3v8l-6 3-6-3z" stroke="currentColor" stroke-width="1.5"/>',
		clock:    '<circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="1.5"/><path d="M8 4v4l3 2" stroke="currentColor" stroke-width="1.5"/>',
		layers:   '<path d="M8 1l6 3-6 3-6-3z" stroke="currentColor" stroke-width="1.5"/><path d="M2 7l6 3 6-3M2 11l6 3 6-3" stroke="currentColor" stroke-width="1.5"/>',
		list:     '<path d="M2 3h12M2 8h12M2 13h8" stroke="currentColor" stroke-width="1.5"/>',
		bell:     '<path d="M3 6a5 5 0 0 1 10 0v3l1 2H2l1-2zM6 13a2 2 0 0 0 4 0" stroke="currentColor" stroke-width="1.5"/>',
		feed:     '<circle cx="3" cy="13" r="1.5" fill="currentColor"/><path d="M2 8a6 6 0 0 1 6 6M2 4a10 10 0 0 1 10 10" stroke="currentColor" stroke-width="1.5"/>',
		database: '<ellipse cx="8" cy="3.5" rx="6" ry="2" stroke="currentColor" stroke-width="1.5"/><path d="M2 3.5v9c0 1.1 2.7 2 6 2s6-.9 6-2v-9" stroke="currentColor" stroke-width="1.5"/><path d="M2 8.5c0 1.1 2.7 2 6 2s6-.9 6-2" stroke="currentColor" stroke-width="1.5"/>',
		sql:      '<path d="M2 4h12M2 8h12M2 12h8" stroke="currentColor" stroke-width="1.5"/>',
		lineage:  '<circle cx="4" cy="8" r="2" stroke="currentColor" stroke-width="1.5"/><circle cx="12" cy="4" r="2" stroke="currentColor" stroke-width="1.5"/><circle cx="12" cy="12" r="2" stroke="currentColor" stroke-width="1.5"/><path d="M6 8l4-4M6 8l4 4" stroke="currentColor" stroke-width="1.5"/>',
		plugin:   '<path d="M3 8l5-6 5 6-5 6z" stroke="currentColor" stroke-width="1.5"/>',
		pod:      '<path d="M2 2h12v12H2z" stroke="currentColor" stroke-width="1.5"/><path d="M5 2v12M2 8h12" stroke="currentColor" stroke-width="1.5"/>',
		storage:  '<rect x="2" y="3" width="12" height="4" stroke="currentColor" stroke-width="1.5"/><rect x="2" y="9" width="12" height="4" stroke="currentColor" stroke-width="1.5"/><circle cx="12" cy="5" r="1" fill="currentColor"/><circle cx="12" cy="11" r="1" fill="currentColor"/>',
		catalog:  '<path d="M2 3h12M2 3v10h12V3" stroke="currentColor" stroke-width="1.5"/><path d="M5 7h6M5 10h4" stroke="currentColor" stroke-width="1.5"/>',
		preset:   '<circle cx="8" cy="8" r="2.5" stroke="currentColor" stroke-width="1.5"/><path d="M8 1v2M8 13v2M1 8h2M13 8h2" stroke="currentColor" stroke-width="1.5"/>',
		chart:    '<path d="M2 8h3l2-5 2 10 2-5h3" stroke="currentColor" stroke-width="1.5"/>',
		ml:       '<path d="M2 12L6 5l3 4 2-3 3 6" stroke="currentColor" stroke-width="1.5"/>',
		module:   '<circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="1.5"/><path d="M8 3v10M3 8h10" stroke="currentColor" stroke-width="1.5"/>',
		user:     '<circle cx="8" cy="5" r="3" stroke="currentColor" stroke-width="1.5"/><path d="M2 14c0-3 2.7-5 6-5s6 2 6 5" stroke="currentColor" stroke-width="1.5"/>',
		group:    '<circle cx="6" cy="5" r="2.5" stroke="currentColor" stroke-width="1.5"/><path d="M1 14c0-2.5 2-4.5 5-4.5s5 2 5 4.5" stroke="currentColor" stroke-width="1.5"/><circle cx="12" cy="6" r="2" stroke="currentColor" stroke-width="1.5"/>',
		shield:   '<path d="M8 1l6 3v5c0 3-2.5 5.5-6 6.5C4.5 14.5 2 12 2 9V4z" stroke="currentColor" stroke-width="1.5"/>',
		log:      '<path d="M3 2h7l3 3v9H3z" stroke="currentColor" stroke-width="1.5"/><path d="M10 2v3h3M5 8h6M5 11h4" stroke="currentColor" stroke-width="1.5"/>',
	};
</script>

<aside class="sidebar" class:collapsed={sidebar.collapsed}>

	<div class="sb-brand">
		<Logo size={24} />
		{#if !sidebar.collapsed}
			<span class="sb-name">{m.app_name()}</span>
		{/if}
	</div>

	<nav class="sb-nav">
		{#each navGroups as group}
			{#if !sidebar.collapsed}
				<div class="sb-section-title">{group.label}</div>
			{:else}
				<div class="sb-section-divider"></div>
			{/if}
			{#each group.items as item}
				<a href={item.href} class="sb-link" class:active={isActive(item.href)} title={sidebar.collapsed ? item.label() : undefined}>
					<span class="sb-icon">
						<svg width="14" height="14" viewBox="0 0 16 16" fill="none">
							{@html icons[item.icon] ?? ''}
						</svg>
					</span>
					{#if !sidebar.collapsed}
						<span class="sb-link-text">{item.label()}</span>
						{#if item.badge}
							<span class="sb-badge">{item.badge}</span>
						{/if}
					{/if}
				</a>
			{/each}
		{/each}
	</nav>

	<div class="sb-foot">
		{#if !sidebar.collapsed}
			<span class="sb-foot-dot"></span>
			<span class="sb-foot-text">cosmonaut-dev</span>
		{/if}
		<button
			class="sb-collapse-btn"
			onclick={() => { sidebar.collapsed = !sidebar.collapsed; }}
			title={sidebar.collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
		>
			<svg width="14" height="14" viewBox="0 0 16 16" fill="none" style={sidebar.collapsed ? 'transform:rotate(180deg)' : ''}>
				<path d="M10 3l-5 5 5 5" stroke="currentColor" stroke-width="1.5"/>
			</svg>
		</button>
	</div>
</aside>

<style>
	.sidebar {
		background: #050504;
		border-right: 1px solid #2A2A28;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		grid-row: 1 / -1;
		width: 220px;
		transition: width 180ms ease;
	}
	.sidebar.collapsed { width: 52px; }

	.sb-brand {
		height: 48px;
		min-height: 48px;
		display: flex;
		align-items: center;
		padding: 0 14px;
		gap: 10px;
		border-bottom: 1px solid #2A2A28;
		flex-shrink: 0;
		overflow: hidden;
	}
	.sb-name {
		font-family: var(--font-display);
		font-weight: 800;
		font-size: 14px;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: #F4F1E8;
		white-space: nowrap;
	}

	.sb-nav {
		flex: 1;
		overflow-y: auto;
		overflow-x: hidden;
		padding-bottom: 8px;
	}
	.sb-nav::-webkit-scrollbar { width: 4px; }
	.sb-nav::-webkit-scrollbar-track { background: transparent; }
	.sb-nav::-webkit-scrollbar-thumb { background: #2A2A28; }

	.sb-section-title {
		padding: 12px 14px 4px;
		font-family: var(--font-mono);
		font-size: 9px;
		font-weight: 700;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: #4A4640;
		border-top: 1px solid #1A1A18;
		margin-top: 6px;
		white-space: nowrap;
	}
	.sb-section-divider {
		height: 1px;
		background: #1A1A18;
		margin: 6px 0;
	}

	.sb-link {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 7px 14px;
		color: #8A857A;
		text-decoration: none;
		font-size: 12.5px;
		font-weight: 500;
		border-left: 2px solid transparent;
		white-space: nowrap;
		overflow: hidden;
	}
	.sb-link:hover { background: #1A1A18; color: #F4F1E8; }
	.sb-link.active {
		background: var(--red);
		color: #F4F1E8;
		border-left-color: #F4F1E8;
		font-weight: 600;
	}
	.sb-icon { flex-shrink: 0; opacity: 0.85; display: flex; }
	.sb-link.active .sb-icon { opacity: 1; }
	.sb-link-text { flex: 1; overflow: hidden; text-overflow: ellipsis; }

	.sb-badge {
		margin-left: auto;
		font-family: var(--font-mono);
		font-size: 10px;
		font-weight: 700;
		padding: 1px 5px;
		background: #2A2A28;
		color: #8A857A;
		flex-shrink: 0;
	}
	.sb-link.active .sb-badge { background: #8B0A1F; color: #F4F1E8; }

	.sb-foot {
		border-top: 1px solid #2A2A28;
		padding: 10px 14px;
		display: flex;
		align-items: center;
		gap: 8px;
		flex-shrink: 0;
		min-height: 44px;
	}
	.sb-foot-dot {
		width: 6px; height: 6px;
		background: var(--ok);
		border-radius: 50%;
		flex-shrink: 0;
	}
	.sb-foot-text {
		font-family: var(--font-mono);
		font-size: 11px;
		color: #6A655A;
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.sb-collapse-btn {
		margin-left: auto;
		width: 24px; height: 24px;
		background: transparent;
		border: 1px solid #2A2A28;
		color: #6A655A;
		cursor: pointer;
		display: grid;
		place-items: center;
		flex-shrink: 0;
	}
	.sb-collapse-btn:hover { background: #1A1A18; color: #F4F1E8; }
</style>