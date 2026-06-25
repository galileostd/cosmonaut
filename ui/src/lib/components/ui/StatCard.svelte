<script lang="ts">
  import SegBar from './SegBar.svelte';

  let { label, value, unit = '', color = 'red', trackPct = 0 }: {
    label: string;
    value: string | number;
    unit?: string;
    color?: 'red' | 'ok' | 'warn' | 'info' | 'mute';
    trackPct?: number;
  } = $props();
</script>

<div class="bg-[var(--bg-elev)] p-4 rounded-lg border border-[var(--line)]">
  <div class="text-sm font-medium text-[var(--fg-soft)] uppercase tracking-wide">
    {label}
  </div>
  <div class="text-3xl font-display font-bold text-[var(--fg)] mt-1">
    {value}
    {#if unit}<span class="text-base font-normal text-[var(--fg-mute)] ml-1">{unit}</span>{/if}
  </div>
  {#if trackPct !== undefined}
    <div class="mt-2">
      <SegBar
        segments={[
          { label: '', pct: trackPct, color: color === 'red' ? 'err' : color === 'ok' ? 'ok' : color === 'warn' ? 'warn' : color === 'info' ? 'info' : 'mute' }
        ]}
        height="h-1.5"
        showLabels={false}
      />
    </div>
  {/if}
</div>
