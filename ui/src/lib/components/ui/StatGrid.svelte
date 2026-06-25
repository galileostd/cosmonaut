<script lang="ts">
  import StatCard from './StatCard.svelte';

  let { stats = [] }: {
    stats: Array<{
      label: string;
      value: string | number;
      unit?: string;
      color?: 'red' | 'ok' | 'warn' | 'info' | 'mute';
      trackPct?: number;
    }>;
  } = $props();
</script>

<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
  {#each stats as stat}
    <StatCard {...stat} />
  {/each}
</div>
