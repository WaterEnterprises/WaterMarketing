<script lang="ts">
  import '../app.css';
  import { page } from '$app/stores';
  import Sidebar from '$lib/components/Sidebar.svelte';
</script>

<div class="flex h-screen overflow-hidden">
  <Sidebar />
  <main class="flex-1 overflow-y-auto p-6 bg-surface-100 dark:bg-surface-900">
    <slot />
  </main>
</div>
