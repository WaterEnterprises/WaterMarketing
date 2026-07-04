<script lang="ts">
  import { onMount } from 'svelte';
  import type { Stats } from '$lib/api';
  import { getStats } from '$lib/api';
  import StatsGrid from '$lib/components/StatsGrid.svelte';

  let stats: Stats | null = null;
  let error: string | null = null;
  let loading = true;

  onMount(async () => {
    try {
      stats = await getStats();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load stats';
    } finally {
      loading = false;
    }
  });
</script>

<div class="max-w-6xl mx-auto">
  <div class="flex justify-between items-center mb-6">
    <h2 class="text-2xl font-bold">Dashboard</h2>
    {#if !loading && !error}
      <a href="/leads" class="btn variant-filled-primary">View All Leads</a>
    {/if}
  </div>

  {#if loading}
    <div class="text-center py-12"><div class="spinner size-12 mx-auto"></div><p class="mt-3 text-surface-500">Loading dashboard...</p></div>
  {:else if error}
    <div class="alert variant-filled-error"><p>{error}</p><p class="text-sm mt-1">Make sure <code class="inline-code">crm serve</code> is running on port 8080.</p></div>
  {:else if stats}
    <StatsGrid {stats} />
  {/if}
</div>
