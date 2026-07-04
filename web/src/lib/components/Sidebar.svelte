<script lang="ts">
  import { page } from '$app/stores';
  const links = [
    { href: '/', label: 'Dashboard', icon: '📊' },
    { href: '/leads', label: 'Leads', icon: '👥' },
  ];
</script>

<aside class="w-64 bg-surface-800 dark:bg-surface-950 text-white flex flex-col">
  <div class="p-4 border-b border-surface-700">
    <h1 class="text-lg font-bold">WaterParty CRM</h1>
    <p class="text-sm text-surface-400">Lead Management</p>
  </div>
  <nav class="flex-1 p-3 space-y-1">
    {#each links as link}
      <a
        href={link.href}
        class="flex items-center gap-3 px-3 py-2 rounded-lg transition-colors
               {$page.url.pathname === link.href
                 ? 'bg-primary-600 text-white'
                 : 'hover:bg-surface-700 text-surface-300'}"
      >
        <span>{link.icon}</span>
        <span>{link.label}</span>
      </a>
    {/each}
  </nav>
  <div class="p-4 border-t border-surface-700 text-xs text-surface-500">
    Water Enterprises
  </div>
</aside>
