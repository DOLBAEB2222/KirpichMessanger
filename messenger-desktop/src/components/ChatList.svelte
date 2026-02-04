<script lang="ts">
  import type { ChatItem } from '$stores/chats';

  export let chats: ChatItem[] = [];
  export let activeId: string | null = null;
  export let onSelect: (id: string) => void;
</script>

<section class="chat-list">
  <header>
    <h2>Chats</h2>
    <button class="new-chat">New</button>
  </header>
  <ul>
    {#each chats as chat}
      <li class:active={chat.id === activeId}>
        <button on:click={() => onSelect(chat.id)}>
          <div class="title">{chat.title}</div>
          {#if chat.lastMessage}
            <div class="preview">{chat.lastMessage}</div>
          {/if}
        </button>
        {#if chat.unreadCount > 0}
          <span class="badge">{chat.unreadCount}</span>
        {/if}
      </li>
    {/each}
  </ul>
</section>

<style>
  .chat-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  ul {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    border-radius: 0.75rem;
    background: var(--surface-2);
  }
  li.active {
    outline: 2px solid var(--primary);
  }
  button {
    background: transparent;
    border: none;
    padding: 0.75rem 1rem;
    text-align: left;
    width: 100%;
    color: inherit;
  }
  .badge {
    background: var(--primary);
    color: white;
    padding: 0.25rem 0.5rem;
    border-radius: 1rem;
    margin-right: 0.75rem;
  }
  .preview {
    font-size: 0.85rem;
    color: var(--text-muted);
  }
</style>
