<script lang="ts">
  import '../app.css';
  import BottomNav from '$lib/components/BottomNav.svelte';
</script>

<div class="flex flex-col h-screen bg-surface-100 dark:bg-surface-900">
  <main class="flex-1 overflow-y-auto p-4 pb-20">
    <slot />
  </main>
  <BottomNav />
</div>
