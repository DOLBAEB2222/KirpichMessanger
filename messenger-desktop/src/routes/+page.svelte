<script lang="ts">
  import ChatList from '$components/ChatList.svelte';
  import MessageView from '$components/MessageView.svelte';
  import UserProfile from '$components/UserProfile.svelte';
  import SettingsPanel from '$components/SettingsPanel.svelte';
  import AuthForm from '$components/AuthForm.svelte';
  import { chatList, activeChatId } from '$stores/chats';
  import { settings } from '$stores/settings';
  import { themeMode } from '$stores/theme';
  import { login } from '$lib/api';

  const demoMessages = [
    {
      id: '1',
      author: 'Kirpich Bot',
      body: 'Welcome to KirpichMessanger Desktop.',
      timestamp: '10:00'
    },
    {
      id: '2',
      author: 'You',
      body: 'Ready to build something great!',
      timestamp: '10:01',
      outgoing: true
    }
  ];

  $chatList = [
    {
      id: 'general',
      title: 'General',
      lastMessage: 'Welcome to KirpichMessanger',
      unreadCount: 0
    },
    {
      id: 'design',
      title: 'Design Sync',
      lastMessage: 'New mockups are ready',
      unreadCount: 2
    }
  ];

  const handleSelect = (id: string) => {
    $activeChatId = id;
  };

  const handleAuth = async (event: CustomEvent) => {
    const { email, password } = event.detail;
    await login({ email, password });
  };
</script>

<svelte:head>
  <title>KirpichMessanger Desktop</title>
</svelte:head>

<div class:dark={$themeMode === 'dark'} class="app">
  <aside>
    <UserProfile name="Guest" email="guest@kirpich.app" />
    <ChatList chats={$chatList} activeId={$activeChatId} onSelect={handleSelect} />
    <SettingsPanel state={$settings} />
  </aside>

  <main>
    <MessageView messages={demoMessages} />
  </main>

  <section class="right-panel">
    <AuthForm on:submit={handleAuth} />
    <div class="shortcut">
      <h3>Shortcuts</h3>
      <ul>
        <li><strong>Ctrl + K</strong> Search chats</li>
        <li><strong>Ctrl + Shift + M</strong> Mute</li>
        <li><strong>Ctrl + Shift + D</strong> Toggle theme</li>
      </ul>
    </div>
  </section>
</div>

<style>
  :global(body) {
    margin: 0;
    font-family: 'Inter', system-ui, sans-serif;
    background: var(--surface-1);
    color: var(--text-primary);
  }
  :global(:root) {
    --primary: #6366f1;
    --surface-1: #f5f5f8;
    --surface-2: #ffffff;
    --text-primary: #111827;
    --text-muted: #6b7280;
    --border: #e5e7eb;
  }
  :global(.dark) {
    --surface-1: #0f172a;
    --surface-2: #1e293b;
    --text-primary: #f8fafc;
    --text-muted: #94a3b8;
    --border: #334155;
  }
  .app {
    display: grid;
    grid-template-columns: 320px 1fr 320px;
    gap: 1.5rem;
    height: 100vh;
    padding: 1.5rem;
    box-sizing: border-box;
  }
  aside,
  .right-panel {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }
  main {
    background: var(--surface-2);
    border-radius: 1.5rem;
    padding: 1.5rem;
  }
  .shortcut {
    background: var(--surface-2);
    padding: 1rem;
    border-radius: 1rem;
  }
  ul {
    padding-left: 1.25rem;
  }
</style>
