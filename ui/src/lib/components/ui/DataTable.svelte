<script lang="ts">
  import Badge from './Badge.svelte';

  let { columns = [], rows = [], onRowClick = () => {} }: {
    columns: Array<{ key: string; label: string; sortable?: boolean }>;
    rows: Array<Record<string, any>>;
    onRowClick?: (row: any) => void;
  } = $props();
</script>

<div class="overflow-x-auto">
  <table class="w-full text-sm">
    <thead>
      <tr class="border-b border-[var(--line)] text-left text-xs font-semibold uppercase text-[var(--fg-mute)]">
        {#each columns as col}
          <th class="py-2 px-3">
            {col.label}
            {#if col.sortable}<span class="ml-1 text-[var(--fg-faint)]">↕</span>{/if}
          </th>
        {/each}
      </tr>
    </thead>
    <tbody>
      {#each rows as row, i}
        <tr
          class="border-b border-[var(--line-2)] hover:bg-[var(--bg-soft)] cursor-pointer transition-colors"
          onclick={() => onRowClick(row)}
        >
          {#each columns as col}
            <td class="py-2 px-3 text-[var(--fg-2)]">
              {#if col.key === 'status'}
                <Badge status={row[col.key]} label={row[col.key]} />
              {:else}
                {row[col.key] ?? '—'}
              {/if}
            </td>
          {/each}
        </tr>
      {:else}
        <tr><td colspan={columns.length} class="py-4 text-center text-[var(--fg-mute)]">No data</td></tr>
      {/each}
    </tbody>
  </table>
</div>
