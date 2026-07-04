<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import type { Lead, OutreachEntry } from '$lib/api';
  import { getLead, getOutreach, updateLeadStatus, logOutreach, deleteLead, tiers, statusColor } from '$lib/api';

  let lead: Lead | null = null;
  let outreach: OutreachEntry[] = [];
  let loading = true;
  let error: string | null = null;

  let newActivityType = 'email';
  let newNotes = '';
  let newOutcome = '';

  const statuses = ['cold', 'contacted', 'replied', 'meeting', 'negotiating', 'closed_won', 'closed_lost'];
  const activityTypes = ['email', 'call', 'meeting', 'note'];

  onMount(async () => {
    const id = $page.params.id;
    try {
      const [l, o] = await Promise.all([getLead(id), getOutreach(id)]);
      lead = l;
      outreach = o;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load lead';
    } finally {
      loading = false;
    }
  });

  async function changeStatus(status: string) {
    if (!lead) return;
    try {
      await updateLeadStatus(lead.id, status);
      lead.status = status;
    } catch (e) {
      alert('Failed to update status');
    }
  }

  async function submitActivity() {
    if (!lead || !newNotes) return;
    try {
      await logOutreach(lead.id, newActivityType, newNotes, newOutcome);
      outreach = await getOutreach(lead.id);
      newNotes = '';
      newOutcome = '';
    } catch (e) {
      alert('Failed to log activity');
    }
  }

  async function removeLead() {
    if (!lead || !confirm(`Delete ${lead.company}?`)) return;
    try {
      await deleteLead(lead.id);
      goto('/leads');
    } catch (e) {
      alert('Failed to delete');
    }
  }
</script>

<div class="max-w-4xl mx-auto">
  <button class="btn variant-ghost-surface mb-4" on:click={() => goto('/leads')}>&larr; Back to Leads</button>

  {#if loading}
    <div class="text-center py-12"><div class="spinner size-10 mx-auto"></div></div>
  {:else if error}
    <div class="alert variant-filled-error">{error}</div>
  {:else if lead}
    <div class="card p-6 mb-6">
      <div class="flex justify-between items-start">
        <div>
          <h2 class="text-2xl font-bold">{lead.company}</h2>
          <p class="text-surface-500 mt-1">
            {lead.contact_name || 'No contact'} &middot; {lead.email || 'No email'}
          </p>
        </div>
        <button class="btn variant-ghost-error" on:click={removeLead}>Delete</button>
      </div>

      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
        <div>
          <div class="text-xs text-surface-500">Tier</div>
          <div class="font-medium">{tiers[lead.tier] || lead.tier}</div>
        </div>
        <div>
          <div class="text-xs text-surface-500">Type</div>
          <div class="font-medium">{lead.type || '-'}</div>
        </div>
        <div>
          <div class="text-xs text-surface-500">Vertical</div>
          <div class="font-medium">{lead.vertical || '-'}</div>
        </div>
        <div>
          <div class="text-xs text-surface-500">Check Size</div>
          <div class="font-medium">{lead.check_size || '-'}</div>
        </div>
        <div>
          <div class="text-xs text-surface-500">Phone</div>
          <div class="font-medium">{lead.phone || '-'}</div>
        </div>
        <div>
          <div class="text-xs text-surface-500">Website</div>
          <div class="font-medium">{lead.website || '-'}</div>
        </div>
        <div>
          <div class="text-xs text-surface-500">Source</div>
          <div class="font-medium">{lead.source || '-'}</div>
        </div>
        <div>
          <div class="text-xs text-surface-500">Pitch Angle</div>
          <div class="font-medium">{lead.pitch_angle || '-'}</div>
        </div>
      </div>

      {#if lead.notes}
        <div class="mt-4">
          <div class="text-xs text-surface-500">Notes</div>
          <div class="mt-1 p-3 bg-surface-200 dark:bg-surface-800 rounded">{lead.notes}</div>
        </div>
      {/if}

      {#if lead.next_action}
        <div class="mt-4 flex items-center gap-2">
          <span class="badge variant-filled-warning">Next: {lead.next_action}</span>
          {#if lead.next_action_date}
            <span class="text-sm text-surface-500">due {lead.next_action_date}</span>
          {/if}
        </div>
      {/if}
    </div>

    <div class="card p-6 mb-6">
      <h3 class="font-bold mb-3">Status</h3>
      <div class="flex flex-wrap gap-2">
        {#each statuses as s}
          <button
            class="badge {s === lead.status ? statusColor(s) : 'variant-filled-surface'}"
            on:click={() => changeStatus(s)}
          >{s.replace('_', ' ')}</button>
        {/each}
      </div>
    </div>

    <div class="card p-6 mb-6">
      <h3 class="font-bold mb-3">Log Activity</h3>
      <div class="flex flex-wrap gap-2 mb-3">
        {#each activityTypes as t}
          <button
            class="btn {newActivityType === t ? 'variant-filled-primary' : 'variant-ghost-surface'} btn-sm"
            on:click={() => newActivityType = t}
          >{t}</button>
        {/each}
      </div>
      <textarea class="input w-full mb-2" rows="2" placeholder="Notes..." bind:value={newNotes}></textarea>
      <input class="input w-full mb-2" type="text" placeholder="Outcome (optional)" bind:value={newOutcome} />
      <button class="btn variant-filled-primary" on:click={submitActivity} disabled={!newNotes}>Log Activity</button>
    </div>

    <div class="card p-6">
      <h3 class="font-bold mb-3">Activity Log ({outreach.length})</h3>
      {#if outreach.length === 0}
        <p class="text-surface-500">No activity logged yet.</p>
      {:else}
        <div class="space-y-3">
          {#each outreach as entry}
            <div class="p-3 bg-surface-200 dark:bg-surface-800 rounded">
              <div class="flex justify-between text-sm">
                <span class="badge variant-filled-surface">{entry.activity_type}</span>
                <span class="text-xs text-surface-500">{entry.created_at}</span>
              </div>
              {#if entry.notes}
                <p class="mt-1">{entry.notes}</p>
              {/if}
              {#if entry.outcome}
                <p class="mt-1 text-sm text-surface-500">Outcome: {entry.outcome}</p>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>
