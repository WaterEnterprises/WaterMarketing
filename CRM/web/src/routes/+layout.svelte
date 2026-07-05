<script lang="ts">
  import '../app.css';
  import BottomNav from '$lib/components/BottomNav.svelte';
</script>

<div class="flex flex-col h-screen bg-surface-100 dark:bg-surface-900">
  <header class="flex items-center gap-3 px-4 py-2 bg-surface-800 dark:bg-surface-950 border-b border-surface-700 shrink-0">
    <img src="/logo.png" alt="WaterParty" class="h-8 w-auto" />
    <span class="font-bold text-lg">WaterParty CRM</span>
  </header>
  <main class="flex-1 overflow-y-auto p-4 pb-20">
    <slot />
  </main>
  <BottomNav />
</div>
