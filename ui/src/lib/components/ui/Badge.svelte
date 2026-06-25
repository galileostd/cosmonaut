<script lang="ts">
  let { status = 'mute', label = '' }: {
    status?: 'ok' | 'warn' | 'err' | 'mute' | 'running';
    label?: string;
  } = $props();
</script>

<span
  class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium
    {status === 'ok' ? 'bg-[var(--ok-soft)] text-[var(--ok)]' :
     status === 'warn' ? 'bg-[var(--warn-soft)] text-[var(--warn)]' :
     status === 'err' ? 'bg-[var(--red-soft)] text-[var(--red)]' :
     status === 'running' ? 'bg-[var(--info-soft)] text-[var(--info)]' :
     'bg-[var(--bg-soft)] text-[var(--fg-mute)]'}"
>
  <span class="w-1.5 h-1.5 rounded-full
    {status === 'ok' ? 'bg-[var(--ok)]' :
     status === 'warn' ? 'bg-[var(--warn)]' :
     status === 'err' ? 'bg-[var(--red)]' :
     status === 'running' ? 'bg-[var(--info)] animate-pulse' :
     'bg-[var(--fg-mute)]'}"
  ></span>
  {label}
</span>