const THEME_KEY = 'cosmonaut-theme';

function createTheme() {
	let current = $state<'dark' | 'light'>('dark');

	function init() {
		if (typeof localStorage === 'undefined') return;
		const saved = localStorage.getItem(THEME_KEY) as 'dark' | 'light' | null;
		current = saved ?? 'dark';
		apply();
	}

	function apply() {
		document.documentElement.setAttribute('data-theme', current);
	}

	function toggle() {
		current = current === 'dark' ? 'light' : 'dark';
		localStorage.setItem(THEME_KEY, current);
		apply();
	}

	return { get current() { return current; }, init, toggle };
}

export const theme = createTheme();