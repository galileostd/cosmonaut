<script lang="ts">
  let { title, meta = '', children }: {
    title: string;
    meta?: string;
    children: () => any;
  } = $props();
</script>

<div class="bg-[var(--bg-elev)] border border-[var(--line)] rounded-lg overflow-hidden">
  <div class="flex items-center justify-between px-4 py-3 border-b border-[var(--line)]">
    <h2 class="text-sm font-semibold text-[var(--fg)] uppercase tracking-wide">{title}</h2>
    {#if meta}<span class="text-xs text-[var(--fg-mute)]">{meta}</span>{/if}
  </div>
  <div class="p-4">
    {@render children()}
  </div>
</div>