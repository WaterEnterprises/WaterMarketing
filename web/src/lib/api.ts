const BASE = '/api';

export interface Lead {
  id: string;
  company: string;
  contact_name: string;
  email: string;
  phone: string;
  website: string;
  tier: string;
  type: string;
  vertical: string;
  check_size: string;
  pitch_angle: string;
  status: string;
  next_action: string;
  next_action_date: string;
  notes: string;
  source: string;
  created_at: string;
  updated_at: string;
}

export interface OutreachEntry {
  id: string;
  lead_id: string;
  activity_type: string;
  notes: string;
  outcome: string;
  created_at: string;
}

export interface Stats {
  total: number;
  by_tier: { tier: string; count: number }[];
  by_status: { status: string; count: number }[];
  followups_due: number;
  recent: Lead[];
}

export const tiers: Record<string, string> = {
  '1': 'VC', '2': 'Corporate', '3': 'Local',
  '4': 'Grant', '5': 'Venue', '6': 'Media',
};

async function fetchJson<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${url}`, {
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || res.statusText);
  }
  return res.json();
}

export function getStats(): Promise<Stats> {
  return fetchJson('/stats');
}

export function getLeads(filters?: Record<string, string>): Promise<Lead[]> {
  const params = new URLSearchParams(filters || {}).toString();
  return fetchJson(`/leads${params ? '?' + params : ''}`);
}

export function getLead(id: string): Promise<Lead> {
  return fetchJson(`/leads/${id}`);
}

export function createLead(data: Partial<Lead>): Promise<Lead> {
  return fetchJson('/leads', { method: 'POST', body: JSON.stringify(data) });
}

export function updateLead(id: string, data: Record<string, unknown>): Promise<Lead> {
  return fetchJson(`/leads/${id}`, { method: 'PUT', body: JSON.stringify(data) });
}

export function deleteLead(id: string): Promise<void> {
  return fetchJson(`/leads/${id}`, { method: 'DELETE' });
}

export function updateLeadStatus(id: string, status: string): Promise<void> {
  return fetchJson(`/leads/${id}/status`, { method: 'PUT', body: JSON.stringify({ status }) });
}

export function getOutreach(leadId: string): Promise<OutreachEntry[]> {
  return fetchJson(`/leads/${leadId}/outreach`);
}

export function logOutreach(leadId: string, activityType: string, notes: string, outcome: string): Promise<{ id: string }> {
  return fetchJson(`/leads/${leadId}/outreach`, {
    method: 'POST',
    body: JSON.stringify({ activity_type: activityType, notes, outcome }),
  });
}

export function getFollowups(): Promise<Lead[]> {
  return fetchJson('/followups');
}

export function statusColor(status: string): string {
  const colors: Record<string, string> = {
    cold: 'variant-filled-surface',
    contacted: 'variant-filled-secondary',
    replied: 'variant-filled-primary',
    meeting: 'variant-filled-warning',
    negotiating: 'variant-filled-tertiary',
    closed_won: 'variant-filled-success',
    closed_lost: 'variant-filled-error',
  };
  return colors[status] || 'variant-filled-surface';
}
