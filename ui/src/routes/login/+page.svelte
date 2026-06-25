<script lang="ts">
	import { goto } from '$app/navigation';
	import Logo from '$lib/components/ui/Logo.svelte';
	import * as m from '$lib/paraglide/messages';

	let username = $state('');
	let password = $state('');
	let loading = $state(false);
	let error = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!username || !password) {
			error = m.login_error_required();
			return;
		}
		loading = true;
		error = '';
		try {
			const res = await fetch('/api/v1/auth/login', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ username, password })
			});
			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				error = data.message ?? m.login_error_invalid();
				return;
			}
			await goto('/');
		} catch {
			error = m.login_error_server();
		} finally {
			loading = false;
		}
	}
</script>

<div class="login-shell" data-theme="dark">

	<!-- LEFT -->
	<div class="left">
		<div class="l-brand">
			<Logo size={40} />
			<span class="l-name">{m.app_name()}</span>
		</div>

		<div class="l-body">
			<div class="l-eyebrow">{m.app_tagline()}</div>
			<h1 class="l-headline">
				{m.login_headline_line1()}<br>
				{m.login_headline_line2()}<br>
				<em>{m.login_headline_line3()}</em> {m.login_headline_line3_suffix()}
			</h1>
			<p class="l-desc">{m.login_desc()}</p>

			<div class="l-stats">
				<div class="l-stat">
					<div class="l-stat-val">3</div>
					<div class="l-stat-lbl">{m.login_stat_services()}</div>
				</div>
				<div class="l-stat">
					<div class="l-stat-val">847</div>
					<div class="l-stat-lbl">{m.login_stat_tables()}</div>
				</div>
				<div class="l-stat">
					<div class="l-stat-val">12</div>
					<div class="l-stat-lbl">{m.login_stat_active_jobs()}</div>
				</div>
				<div class="l-stat">
					<div class="l-stat-val">100<span class="pct">%</span></div>
					<div class="l-stat-lbl">{m.login_stat_healthy()}</div>
				</div>
			</div>
		</div>

		<div class="l-foot">{m.app_license()}</div>
	</div>

	<!-- RIGHT -->
	<div class="right">
		<div class="r-form">
			<div class="r-title">{m.login_title()}</div>
			<div class="r-sub">cosmonaut-dev · on-prem</div>

			{#if error}
				<div class="r-error">
					<svg width="14" height="14" viewBox="0 0 16 16" fill="none">
						<circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="1.5"/>
						<path d="M8 5v3M8 10v.5" stroke="currentColor" stroke-width="2"/>
					</svg>
					{error}
				</div>
			{/if}

			<form onsubmit={handleSubmit}>
				<div class="field">
					<label for="username">{m.login_username()}</label>
					<input
						id="username"
						type="text"
						placeholder={m.login_username_placeholder()}
						autocomplete="username"
						bind:value={username}
						disabled={loading}
					/>
				</div>

				<div class="field">
					<div class="field-row">
						<label for="password">{m.login_password()}</label>
						<a href="/forgot-password" class="field-link">{m.login_forgot_password()}</a>
					</div>
					<input
						id="password"
						type="password"
						placeholder={m.login_password_placeholder()}
						autocomplete="current-password"
						bind:value={password}
						disabled={loading}
					/>
				</div>

				<button type="submit" class="btn-submit" disabled={loading}>
					{#if loading}
						<svg width="14" height="14" viewBox="0 0 16 16" fill="none" class="spin">
							<circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="1.5" stroke-dasharray="28" stroke-dashoffset="10"/>
						</svg>
						{m.login_signing_in()}
					{:else}
						<svg width="14" height="14" viewBox="0 0 16 16" fill="none">
							<path d="M6 3l5 5-5 5" stroke="currentColor" stroke-width="2"/>
						</svg>
						{m.login_submit()}
					{/if}
				</button>
			</form>

			<div class="divider">or</div>

			<button class="btn-ldap" onclick={() => goto('/login?method=ldap')}>
				<svg width="16" height="16" viewBox="0 0 16 16" fill="none">
					<rect x="1" y="1" width="6" height="6" stroke="currentColor" stroke-width="1.5"/>
					<rect x="9" y="1" width="6" height="6" stroke="currentColor" stroke-width="1.5"/>
					<rect x="1" y="9" width="6" height="6" stroke="currentColor" stroke-width="1.5"/>
					<rect x="9" y="9" width="6" height="6" stroke="currentColor" stroke-width="1.5"/>
				</svg>
				{m.login_ldap()}
			</button>

			<div class="r-foot">
				{m.app_version()} · <a href="/docs">{m.app_docs()}</a> · <a href="https://github.com/galileostd/cosmonaut" target="_blank">{m.app_github()}</a>
			</div>
		</div>
	</div>
</div>

<style>
	.login-shell {
		display: grid;
		grid-template-columns: 1fr 480px;
		height: 100vh;
		font-family: var(--font-sans);
		background: var(--bg-base);
		color: var(--fg);
	}

	.left {
		background: #050504;
		border-right: 1px solid #2A2A28;
		display: flex;
		flex-direction: column;
		padding: 48px;
		position: relative;
		overflow: hidden;
	}
	.left::before {
		content: '';
		position: absolute;
		inset: 0;
		background-image:
			linear-gradient(#C8102E08 1px, transparent 1px),
			linear-gradient(90deg, #C8102E08 1px, transparent 1px);
		background-size: 48px 48px;
		pointer-events: none;
	}
	.left::after {
		content: '';
		position: absolute;
		top: 0; right: 0;
		width: 6px; height: 100%;
		background: #C8102E;
	}

	.l-brand {
		display: flex;
		align-items: center;
		gap: 14px;
		position: relative;
		z-index: 1;
	}
	.l-name {
		font-family: var(--font-display);
		font-weight: 800;
		font-size: 22px;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: #F4F1E8;
	}

	.l-body {
		flex: 1;
		display: flex;
		flex-direction: column;
		justify-content: center;
		position: relative;
		z-index: 1;
		max-width: 480px;
	}
	.l-eyebrow {
		font-family: var(--font-mono);
		font-size: 10px;
		font-weight: 700;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: #C8102E;
		margin-bottom: 16px;
	}
	.l-headline {
		font-family: var(--font-display);
		font-weight: 800;
		font-size: 48px;
		letter-spacing: -0.03em;
		line-height: 1.0;
		color: #F4F1E8;
		margin: 0 0 24px;
	}
	.l-headline em {
		font-style: normal;
		color: #C8102E;
	}
	.l-desc {
		font-size: 14px;
		color: #6A655A;
		line-height: 1.6;
		max-width: 380px;
		margin: 0 0 48px;
	}

	.l-stats {
		display: flex;
		border: 1px solid #2A2A28;
		position: relative;
		z-index: 1;
	}
	.l-stat {
		flex: 1;
		padding: 14px 18px;
		border-right: 1px solid #2A2A28;
	}
	.l-stat:last-child { border-right: none; }
	.l-stat-val {
		font-family: var(--font-display);
		font-weight: 800;
		font-size: 24px;
		letter-spacing: -0.02em;
		color: #F4F1E8;
		font-variant-numeric: tabular-nums;
		line-height: 1;
	}
	.pct { font-size: 14px; color: #4A4640; margin-left: 2px; }
	.l-stat-lbl {
		font-family: var(--font-mono);
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: #4A4640;
		margin-top: 5px;
	}

	.l-foot {
		font-family: var(--font-mono);
		font-size: 11px;
		color: #2A2A28;
		position: relative;
		z-index: 1;
		margin-top: 24px;
	}

	.right {
		background: var(--bg-soft);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 48px 40px;
	}
	.r-form { width: 100%; max-width: 360px; }

	.r-title {
		font-family: var(--font-display);
		font-weight: 800;
		font-size: 26px;
		letter-spacing: -0.02em;
		color: var(--fg);
		margin-bottom: 6px;
	}
	.r-sub {
		font-size: 13px;
		color: var(--fg-mute);
		margin-bottom: 36px;
		font-family: var(--font-mono);
	}

	.r-error {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 14px;
		background: var(--red-soft);
		border: 1px solid var(--red);
		border-left: 3px solid var(--red);
		font-family: var(--font-mono);
		font-size: 12px;
		color: var(--red);
		margin-bottom: 20px;
	}

	.field { margin-bottom: 20px; }
	.field label {
		display: block;
		font-family: var(--font-mono);
		font-size: 10px;
		font-weight: 700;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: var(--fg-mute);
		margin-bottom: 7px;
	}
	.field-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 7px;
	}
	.field-row label { margin-bottom: 0; }
	.field-link {
		font-family: var(--font-mono);
		font-size: 10px;
		color: var(--red);
		text-decoration: none;
	}
	.field-link:hover { text-decoration: underline; }
	.field input {
		width: 100%;
		height: 40px;
		padding: 0 12px;
		background: var(--bg-elev);
		border: 1px solid var(--line-2);
		color: var(--fg);
		font-family: var(--font-mono);
		font-size: 13px;
		outline: none;
		transition: border-color 80ms;
	}
	.field input:focus { border-color: var(--red); }
	.field input::placeholder { color: var(--fg-mute); }
	.field input:disabled { opacity: 0.6; cursor: not-allowed; }

	.btn-submit {
		width: 100%;
		height: 42px;
		background: var(--red);
		border: none;
		color: #F4F1E8;
		font-family: var(--font-display);
		font-weight: 700;
		font-size: 14px;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		cursor: pointer;
		margin-top: 8px;
		transition: background 80ms;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 10px;
	}
	.btn-submit:hover:not(:disabled) { background: var(--red-deep); }
	.btn-submit:disabled { opacity: 0.7; cursor: not-allowed; }

	.divider {
		display: flex;
		align-items: center;
		gap: 12px;
		margin: 28px 0;
		color: var(--fg-mute);
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.12em;
		text-transform: uppercase;
	}
	.divider::before, .divider::after {
		content: '';
		flex: 1;
		height: 1px;
		background: var(--line-2);
	}

	.btn-ldap {
		width: 100%;
		height: 40px;
		background: transparent;
		border: 1px solid var(--line-2);
		color: var(--fg-soft);
		font-family: var(--font-sans);
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 10px;
		transition: background 80ms, border-color 80ms;
	}
	.btn-ldap:hover {
		background: var(--bg-sunk);
		border-color: var(--fg-mute);
		color: var(--fg);
	}

	.r-foot {
		margin-top: 36px;
		text-align: center;
		font-family: var(--font-mono);
		font-size: 10px;
		color: var(--fg-mute);
	}
	.r-foot a { color: var(--fg-mute); text-decoration: underline; }
	.r-foot a:hover { color: var(--fg-soft); }

	@keyframes spin { to { transform: rotate(360deg); } }
	.spin { animation: spin 0.8s linear infinite; }
</style>
