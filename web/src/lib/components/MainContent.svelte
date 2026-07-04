<script lang="ts">
  import { onMount } from 'svelte';
  import { activeTab } from '$lib/store';
  import { page } from '$app/stores';
  import { getStats, getLeads, getLead, getOutreach, deleteLead, updateLeadStatus, logOutreach, sendBulkEmail, exportLeadsCSV, tiers, statusColor } from '$lib/api';
  import StatsGrid from '$lib/components/StatsGrid.svelte';
  import type { Stats, Lead, OutreachEntry } from '$lib/api';

  $: tab = $activeTab;

  let stats: Stats | null = null;
  let statsError: string | null = null;
  let statsLoading = true;

  let paginatedLeads: { data: Lead[]; total: number; page: number; limit: number } = { data: [], total: 0, page: 1, limit: 50 };
  let leadsError: string | null = null;
  let leadsLoading = true;
  let search = '';
  let statusFilter = '';
  let tierFilter = '';
  let typeFilter = '';
  let currentPage = 1;
  const limit = 50;

  let selectedIds = new Set<string>();
  let selectAll = false;

  let selectedLead: Lead | null = null;
  let outreach: OutreachEntry[] = [];
  let outreachLoading = false;

  let newActivityType = 'email';
  let newNotes = '';
  let newOutcome = '';

  let showEmailModal = false;
  let emailSubject = '';
  let emailBody = '';
  let sendingEmail = false;
  let emailResult: string | null = null;

  let exporting = false;

  const statuses = ['cold', 'contacted', 'replied', 'meeting', 'negotiating', 'closed_won', 'closed_lost'];
  const activityTypes = ['email', 'call', 'meeting', 'note'];
  const filterStatuses = ['', 'cold', 'contacted', 'replied', 'meeting', 'negotiating', 'closed_won', 'closed_lost'];
  const tierOptions = ['', '1', '2', '3', '4', '5', '6'];
  const filterTypes = ['', 'Investor', 'Sponsor', 'Partner', 'Venue', 'Media', 'Grant', 'Other'];

  const path = typeof window !== 'undefined' ? window.location.pathname : '/';
  if (path === '/leads') {
    activeTab.set('leads');
  }

  onMount(async () => {
    try { stats = await getStats(); } catch (e) { statsError = 'Failed to load stats'; }
    finally { statsLoading = false; }
    await loadLeads();
  });

  async function loadLeads() {
    leadsLoading = true;
    leadsError = null;
    try {
      const filters: Record<string, string> = { page: String(currentPage), limit: String(limit) };
      if (search) filters.search = search;
      if (statusFilter) filters.status = statusFilter;
      if (tierFilter) filters.tier = tierFilter;
      if (typeFilter) filters.type = typeFilter;
      paginatedLeads = await getLeads(filters);
    } catch (e) { leadsError = 'Failed to load leads'; }
    finally {
      leadsLoading = false;
      selectedIds.clear();
      selectAll = false;
    }
  }

  function goPage(p: number) {
    currentPage = p;
    loadLeads();
  }

  function totalPages(): number {
    return Math.max(1, Math.ceil(paginatedLeads.total / limit));
  }

  function toggleSelectAll() {
    selectAll = !selectAll;
    if (selectAll) {
      paginatedLeads.data.forEach(l => selectedIds.add(l.id));
    } else {
      selectedIds.clear();
    }
  }

  function toggleSelect(id: string) {
    if (selectedIds.has(id)) {
      selectedIds.delete(id);
      selectAll = false;
    } else {
      selectedIds.add(id);
      if (selectedIds.size === paginatedLeads.data.length) selectAll = true;
    }
    selectedIds = selectedIds;
    selectAll = selectAll;
  }

  function openEmailModal() {
    emailSubject = '';
    emailBody = '';
    emailResult = null;
    showEmailModal = true;
  }

  async function sendEmail() {
    if (selectedIds.size === 0 || !emailSubject || !emailBody) return;
    sendingEmail = true;
    emailResult = null;
    try {
      const emails: string[] = [];
      for (const id of selectedIds) {
        const l = paginatedLeads.data.find(d => d.id === id);
        if (l && l.email) {
          l.email.split(',').map(e => e.trim()).filter(e => e).forEach(e => emails.push(e));
        }
      }
      if (emails.length === 0) { emailResult = 'No email addresses found for selected leads.'; return; }
      const res = await sendBulkEmail(emails, emailSubject, emailBody);
      emailResult = `Sent to ${res.sent} recipient(s)`;
    } catch (e: any) {
      emailResult = `Error: ${e.message}`;
    } finally {
      sendingEmail = false;
    }
  }

  async function doExport() {
    if (selectedIds.size === 0) return;
    exporting = true;
    try {
      await exportLeadsCSV(Array.from(selectedIds));
    } catch (e: any) {
      alert('Export failed: ' + e.message);
    } finally {
      exporting = false;
    }
  }

  async function viewLead(id: string) {
    try {
      const [l, o] = await Promise.all([getLead(id), getOutreach(id)]);
      selectedLead = l;
      outreach = o;
    } catch (e) { alert('Failed to load lead'); }
  }

  function backToList() {
    selectedLead = null;
    outreach = [];
  }

  async function changeStatus(status: string) {
    if (!selectedLead) return;
    try {
      await updateLeadStatus(selectedLead.id, status);
      selectedLead.status = status;
    } catch (e) { alert('Failed to update status'); }
  }

  async function submitActivity() {
    if (!selectedLead || !newNotes) return;
    try {
      await logOutreach(selectedLead.id, newActivityType, newNotes, newOutcome);
      outreach = await getOutreach(selectedLead.id);
      newNotes = '';
      newOutcome = '';
    } catch (e) { alert('Failed to log activity'); }
  }

  async function removeLead() {
    if (!selectedLead || !confirm(`Delete ${selectedLead.company}?`)) return;
    try {
      await deleteLead(selectedLead.id);
      selectedLead = null;
      outreach = [];
      loadLeads();
    } catch (e) { alert('Failed to delete'); }
  }
</script>

<div class="max-w-6xl mx-auto" style="display: {tab === 'dashboard' ? 'block' : 'none'}">
  <div class="flex justify-between items-center mb-6">
    <h2 class="text-2xl font-bold">Dashboard</h2>
  </div>
  {#if statsLoading}
    <div class="text-center py-12"><div class="spinner size-12 mx-auto"></div><p class="mt-3 text-surface-500">Loading dashboard...</p></div>
  {:else if statsError}
    <div class="alert variant-filled-error"><p>{statsError}</p></div>
  {:else if stats}
    <StatsGrid {stats} />
  {/if}
</div>

<div class="max-w-6xl mx-auto" style="display: {tab === 'leads' ? 'block' : 'none'}">
  {#if selectedLead}
    <button class="btn variant-ghost-surface mb-4" on:click={backToList}>&larr; Back to Leads</button>
    <div class="card p-6 mb-6">
      <div class="flex justify-between items-start">
        <div>
          <h2 class="text-2xl font-bold">{selectedLead.company}</h2>
          <p class="text-surface-500 mt-1">{selectedLead.contact_name || 'No contact'} &middot; {selectedLead.email || 'No email'}</p>
        </div>
        <button class="btn variant-ghost-error" on:click={removeLead}>Delete</button>
      </div>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
        <div><div class="text-xs text-surface-500">Tier</div><div class="font-medium">{tiers[selectedLead.tier] || selectedLead.tier}</div></div>
        <div><div class="text-xs text-surface-500">Type</div><div class="font-medium">{selectedLead.type || '-'}</div></div>
        <div><div class="text-xs text-surface-500">Vertical</div><div class="font-medium">{selectedLead.vertical || '-'}</div></div>
        <div><div class="text-xs text-surface-500">Check Size</div><div class="font-medium">{selectedLead.check_size || '-'}</div></div>
        <div><div class="text-xs text-surface-500">Phone</div><div class="font-medium">{selectedLead.phone || '-'}</div></div>
        <div><div class="text-xs text-surface-500">Website</div><div class="font-medium">{selectedLead.website || '-'}</div></div>
        <div><div class="text-xs text-surface-500">Source</div><div class="font-medium">{selectedLead.source || '-'}</div></div>
        <div><div class="text-xs text-surface-500">Pitch</div><div class="font-medium">{selectedLead.pitch_angle || '-'}</div></div>
      </div>
      {#if selectedLead.notes}
        <div class="mt-4"><div class="text-xs text-surface-500">Notes</div><div class="mt-1 p-3 bg-surface-200 dark:bg-surface-800 rounded">{selectedLead.notes}</div></div>
      {/if}
      {#if selectedLead.next_action}
        <div class="mt-4 flex items-center gap-2"><span class="badge variant-filled-warning">Next: {selectedLead.next_action}{#if selectedLead.next_action_date} (due {selectedLead.next_action_date}){/if}</span></div>
      {/if}
    </div>
    <div class="card p-6 mb-6">
      <h3 class="font-bold mb-3">Status</h3>
      <div class="flex flex-wrap gap-2">
        {#each statuses as s}
          <button class="badge {s === selectedLead.status ? statusColor(s) : 'variant-filled-surface'}" on:click={() => changeStatus(s)}>{s.replace('_', ' ')}</button>
        {/each}
      </div>
    </div>
    <div class="card p-6 mb-6">
      <h3 class="font-bold mb-3">Log Activity</h3>
      <div class="flex flex-wrap gap-2 mb-3">
        {#each activityTypes as t}
          <button class="btn {newActivityType === t ? 'variant-filled-primary' : 'variant-ghost-surface'} btn-sm" on:click={() => newActivityType = t}>{t}</button>
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
              {#if entry.notes}<p class="mt-1">{entry.notes}</p>{/if}
              {#if entry.outcome}<p class="mt-1 text-sm text-surface-500">Outcome: {entry.outcome}</p>{/if}
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {:else}
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-2xl font-bold">Leads ({paginatedLeads.total})</h2>
    </div>
    <div class="card p-4 mb-4">
      <div class="grid grid-cols-1 md:grid-cols-5 gap-3">
        <input class="input" type="text" placeholder="Search company, contact, email..." bind:value={search} />
        <select class="select" bind:value={statusFilter}>
          <option value="">All Statuses</option>
          {#each filterStatuses.slice(1) as s}<option value={s}>{s.replace('_', ' ')}</option>{/each}
        </select>
        <select class="select" bind:value={tierFilter}>
          <option value="">All Tiers</option>
          {#each tierOptions.slice(1) as t}<option value={t}>{tiers[t]}</option>{/each}
        </select>
        <select class="select" bind:value={typeFilter}>
          <option value="">All Types</option>
          {#each filterTypes.slice(1) as t}<option value={t}>{t}</option>{/each}
        </select>
        <button class="btn variant-filled-primary" on:click={loadLeads}>Filter</button>
      </div>
    </div>

    {#if selectedIds.size > 0}
      <div class="sticky top-0 z-10 card p-3 mb-4 flex items-center justify-between variant-filled-primary">
        <span class="font-bold">{selectedIds.size} selected</span>
        <div class="flex gap-2">
          <button class="btn variant-filled-secondary btn-sm" on:click={openEmailModal}>Send Email</button>
          <button class="btn variant-filled-secondary btn-sm" on:click={doExport} disabled={exporting}>{exporting ? 'Exporting...' : 'Export CSV'}</button>
          <button class="btn variant-ghost-surface btn-sm" on:click={() => { selectedIds.clear(); selectAll = false; }}>Clear</button>
        </div>
      </div>
    {/if}

    {#if leadsLoading}
      <div class="text-center py-12"><div class="spinner size-10 mx-auto"></div></div>
    {:else if leadsError}
      <div class="alert variant-filled-error">{leadsError}</div>
    {:else if paginatedLeads.data.length === 0}
      <div class="card p-8 text-center text-surface-500">No leads found.</div>
    {:else}
      <div class="card overflow-x-auto">
        <table class="table w-full">
          <thead>
            <tr>
              <th class="w-10"><input type="checkbox" checked={selectAll} on:change={toggleSelectAll} /></th>
              <th>Company</th><th>Contact</th><th>Tier</th><th>Type</th><th>Status</th><th>Next Action</th>
            </tr>
          </thead>
          <tbody>
            {#each paginatedLeads.data as lead}
              <tr class="cursor-pointer hover:bg-surface-200 dark:hover:bg-surface-800" on:click={() => viewLead(lead.id)}>
                <td class="w-10" on:click|stopPropagation><input type="checkbox" checked={selectedIds.has(lead.id)} on:change={() => toggleSelect(lead.id)} /></td>
                <td class="font-medium">{lead.company}</td>
                <td><div>{lead.contact_name || '-'}</div><div class="text-xs text-surface-500">{lead.email || ''}</div></td>
                <td>{tiers[lead.tier] || lead.tier}</td>
                <td>{lead.type || '-'}</td>
                <td><span class="badge {statusColor(lead.status)}">{lead.status.replace('_', ' ')}</span></td>
                <td class="text-sm">{#if lead.next_action}<div>{lead.next_action}</div>{#if lead.next_action_date}<div class="text-xs text-surface-500">{lead.next_action_date}</div>{/if}{:else}<span class="text-surface-400">-</span>{/if}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      <div class="flex justify-between items-center mt-4">
        <span class="text-sm text-surface-500">Page {paginatedLeads.page} of {totalPages()} ({paginatedLeads.total} total)</span>
        <div class="flex gap-1">
          <button class="btn variant-ghost-surface btn-sm" disabled={currentPage <= 1} on:click={() => goPage(currentPage - 1)}>Prev</button>
          {#each Array(Math.min(totalPages(), 10)) as _, i}
            {@const p = i + 1}
            <button class="btn btn-sm {p === currentPage ? 'variant-filled-primary' : 'variant-ghost-surface'}" on:click={() => goPage(p)}>{p}</button>
          {/each}
          <button class="btn variant-ghost-surface btn-sm" disabled={currentPage >= totalPages()} on:click={() => goPage(currentPage + 1)}>Next</button>
        </div>
      </div>
    {/if}
  {/if}
</div>

{#if showEmailModal}
  <dialog class="modal-overlay" open>
    <div class="modal-content card p-6 max-w-lg w-full">
      <h3 class="text-xl font-bold mb-4">Send Email to {selectedIds.size} lead(s)</h3>
      {#if emailResult}
        <div class="alert variant-filled-surface mb-4">{emailResult}</div>
      {:else}
        <div class="mb-4">
          <label class="label" for="email-subject">Subject</label>
          <input class="input w-full" id="email-subject" type="text" bind:value={emailSubject} placeholder="Email subject..." />
        </div>
        <div class="mb-4">
          <label class="label" for="email-body">Body</label>
          <textarea class="input w-full" id="email-body" rows="6" bind:value={emailBody} placeholder="Email body..."></textarea>
        </div>
      {/if}
      <div class="flex justify-end gap-2">
        <button class="btn variant-ghost-surface" on:click={() => showEmailModal = false} disabled={sendingEmail}>Close</button>
        {#if !emailResult}
          <button class="btn variant-filled-primary" on:click={sendEmail} disabled={sendingEmail || !emailSubject || !emailBody}>
            {sendingEmail ? 'Sending...' : 'Send'}
          </button>
        {/if}
      </div>
    </div>
  </dialog>
{/if}

<style>
  :global(dialog.modal-overlay) {
    position: fixed;
    inset: 0;
    z-index: 50;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0,0,0,0.6);
    border: none;
    width: 100%;
    height: 100%;
  }
</style>
