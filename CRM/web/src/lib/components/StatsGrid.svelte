<script lang="ts">
  import type { Stats } from '$lib/api';
  import { tiers } from '$lib/api';

  export let stats: Stats;
</script>

<div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
  <div class="card p-4">
    <div class="text-3xl font-bold">{stats.total}</div>
    <div class="text-sm text-surface-500">Total Leads</div>
  </div>
  <div class="card p-4">
    <div class="text-3xl font-bold">{stats.followups_due}</div>
    <div class="text-sm text-surface-500">Follow-ups Due</div>
  </div>
  <div class="card p-4">
    <div class="text-3xl font-bold">{stats.by_status.length}</div>
    <div class="text-sm text-surface-500">Status Buckets</div>
  </div>
  <div class="card p-4">
    <div class="text-3xl font-bold">{stats.by_tier.length}</div>
    <div class="text-sm text-surface-500">Tiers Active</div>
  </div>
</div>

<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
  <div class="card p-4">
    <h3 class="font-bold mb-3">By Tier</h3>
    <div class="space-y-2">
      {#each stats.by_tier as t}
        <div class="flex justify-between items-center">
          <span>{tiers[t.tier] || `Tier ${t.tier}`}</span>
          <span class="badge variant-filled-primary">{t.count}</span>
        </div>
      {/each}
    </div>
  </div>
  <div class="card p-4">
    <h3 class="font-bold mb-3">By Status</h3>
    <div class="space-y-2">
      {#each stats.by_status as s}
        <div class="flex justify-between items-center">
          <span class="capitalize">{s.status.replace('_', ' ')}</span>
          <span class="badge variant-filled-surface">{s.count}</span>
        </div>
      {/each}
    </div>
  </div>
</div>

{#if stats.recent.length}
  <div class="card p-4 mt-6">
    <h3 class="font-bold mb-3">Recent Leads</h3>
    <table class="table w-full">
      <thead>
        <tr>
          <th>Company</th>
          <th>Tier</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        {#each stats.recent as lead}
          <tr>
            <td><a href="/leads/{lead.id}" class="anchor">{lead.company}</a></td>
            <td>{tiers[lead.tier] || lead.tier}</td>
            <td><span class="badge variant-filled-surface">{lead.status}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}
