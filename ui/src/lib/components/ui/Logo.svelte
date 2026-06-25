<script lang="ts">
	interface Props {
		size?: number;
	}
	let { size = 40 }: Props = $props();
</script>

<svg width={size} height={size} viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
	<ellipse cx="20" cy="20" rx="18" ry="7" transform="rotate(-22 20 20)" stroke="#C8102E" stroke-width="2"/>
	<circle cx="20" cy="20" r="8" fill="#F4F1E8"/>
	<circle cx="20" cy="20" r="8" stroke="#0B0B0A" stroke-width="1.5"/>
	<circle cx="36" cy="14" r="2.2" fill="#C8102E"/>
	<circle cx="6"  cy="26" r="1.8" fill="#B8862A"/>
	<circle cx="14" cy="6"  r="1.4" fill="#B8862A"/>
</svg>
