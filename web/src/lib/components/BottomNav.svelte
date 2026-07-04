<script lang="ts">
  import { activeTab } from '$lib/store';

  const links = [
    { id: 'dashboard', label: 'Dashboard', icon: '📊' },
    { id: 'leads', label: 'Leads', icon: '👥' },
  ];

  $: currentTab = $activeTab;
</script>

<nav class="fixed bottom-0 left-0 right-0 flex bg-surface-800 dark:bg-surface-950 border-t border-surface-700 z-50">
  {#each links as link}
    <button
      class="flex flex-1 flex-col items-center gap-1 py-3 text-xs transition-colors cursor-pointer border-0 bg-transparent
             {currentTab === link.id
               ? 'text-primary-400'
               : 'text-surface-400 hover:text-surface-200'}"
      on:click={() => activeTab.set(link.id)}
    >
      <span class="text-xl">{link.icon}</span>
      <span>{link.label}</span>
    </button>
  {/each}
</nav>
