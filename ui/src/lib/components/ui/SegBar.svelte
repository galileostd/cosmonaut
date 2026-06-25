<script lang="ts">
  let { segments, height = 'h-2', showLabels = true }: {
    segments: Array<{ label?: string; pct: number; color: 'ok' | 'warn' | 'err' | 'mute' | 'info' }>;
    height?: string;
    showLabels?: boolean;
  } = $props();
</script>

<div class="space-y-1">
  <div class="flex w-full overflow-hidden rounded-full bg-[var(--bg-sunk)] {height}">
    {#each segments as seg}
      {@const colorClass =
        seg.color === 'ok' ? 'bg-[var(--ok)]' :
        seg.color === 'warn' ? 'bg-[var(--warn)]' :
        seg.color === 'err' ? 'bg-[var(--red)]' :
        seg.color === 'info' ? 'bg-[var(--info)]' :
        'bg-[var(--fg-mute)]'}
      <div
        class="{colorClass} transition-all"
        style="width: {Math.min(seg.pct, 100)}%;"
      ></div>
    {/each}
  </div>
  {#if showLabels}
    <div class="flex justify-between text-xs text-[var(--fg-mute)]">
      {#each segments as seg}
        <span>{seg.label ?? ''}</span>
      {/each}
    </div>
  {/if}
</div>