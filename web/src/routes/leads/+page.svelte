<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import type { Lead } from '$lib/api';
  import { getLeads, tiers, statusColor } from '$lib/api';

  let leads: Lead[] = [];
  let loading = true;
  let error: string | null = null;

  let search = '';
  let statusFilter = '';
  let tierFilter = '';

  const statuses = ['', 'cold', 'contacted', 'replied', 'meeting', 'negotiating', 'closed_won', 'closed_lost'];
  const tierOptions = ['', '1', '2', '3', '4', '5', '6'];

  onMount(() => load());

  async function load() {
    loading = true;
    error = null;
    try {
      const filters: Record<string, string> = {};
      if (search) filters.search = search;
      if (statusFilter) filters.status = statusFilter;
      if (tierFilter) filters.tier = tierFilter;
      leads = await getLeads(filters);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load leads';
    } finally {
      loading = false;
    }
  }

  function viewLead(id: string) {
    goto(`/leads/${id}`);
  }
</script>

<div class="max-w-6xl mx-auto">
  <div class="flex justify-between items-center mb-4">
    <h2 class="text-2xl font-bold">Leads ({leads.length})</h2>
  </div>

  <div class="card p-4 mb-4">
    <div class="grid grid-cols-1 md:grid-cols-4 gap-3">
      <input class="input" type="text" placeholder="Search company, contact, email..." bind:value={search} />
      <select class="select" bind:value={statusFilter}>
        <option value="">All Statuses</option>
        {#each statuses.slice(1) as s}
          <option value={s}>{s.replace('_', ' ')}</option>
        {/each}
      </select>
      <select class="select" bind:value={tierFilter}>
        <option value="">All Tiers</option>
        {#each tierOptions.slice(1) as t}
          <option value={t}>{tiers[t]}</option>
        {/each}
      </select>
      <button class="btn variant-filled-primary" on:click={load}>Filter</button>
    </div>
  </div>

  {#if loading}
    <div class="text-center py-12"><div class="spinner size-10 mx-auto"></div></div>
  {:else if error}
    <div class="alert variant-filled-error">{error}</div>
  {:else if leads.length === 0}
    <div class="card p-8 text-center text-surface-500">No leads found.</div>
  {:else}
    <div class="card overflow-x-auto">
      <table class="table w-full">
        <thead>
          <tr>
            <th>Company</th>
            <th>Contact</th>
            <th>Tier</th>
            <th>Type</th>
            <th>Status</th>
            <th>Next Action</th>
          </tr>
        </thead>
        <tbody>
          {#each leads as lead}
            <tr class="cursor-pointer hover:bg-surface-200 dark:hover:bg-surface-800" on:click={() => viewLead(lead.id)}>
              <td class="font-medium">{lead.company}</td>
              <td>
                <div>{lead.contact_name || '-'}</div>
                <div class="text-xs text-surface-500">{lead.email || ''}</div>
              </td>
              <td>{tiers[lead.tier] || lead.tier}</td>
              <td>{lead.type || '-'}</td>
              <td><span class="badge {statusColor(lead.status)}">{lead.status.replace('_', ' ')}</span></td>
              <td class="text-sm">
                {#if lead.next_action}
                  <div>{lead.next_action}</div>
                  {#if lead.next_action_date}
                    <div class="text-xs text-surface-500">{lead.next_action_date}</div>
                  {/if}
                {:else}
                  <span class="text-surface-400">-</span>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>
