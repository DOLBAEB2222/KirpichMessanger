<script lang="ts">
  import type { SettingsState } from '$lib/stores/settings';

  export let state: SettingsState;
</script>

<section class="settings">
  <header>
    <h2>Settings</h2>
    <p>Customize notifications, theme, and performance.</p>
  </header>
  <div class="row">
    <label>
      <input type="checkbox" bind:checked={state.notificationsEnabled} />
      Notifications
    </label>
    <label>
      <input type="checkbox" bind:checked={state.autoStart} />
      Auto-start
    </label>
    <label>
      <input type="checkbox" bind:checked={state.compactMode} />
      Compact mode
    </label>
  </div>
</section>

<style>
  .settings {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    background: var(--surface-2);
    padding: 1rem;
    border-radius: 1rem;
  }
  .row {
    display: grid;
    gap: 0.5rem;
  }
</style>
