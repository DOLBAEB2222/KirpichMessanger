<script lang="ts">
  export type MessageItem = {
    id: string;
    author: string;
    body: string;
    timestamp: string;
    outgoing?: boolean;
  };

  export let messages: MessageItem[] = [];
</script>

<section class="message-view">
  <header>
    <h2>Conversation</h2>
    <p class="sub">Real-time chat · Voice & Video ready</p>
  </header>
  <div class="messages">
    {#each messages as message}
      <article class:outgoing={message.outgoing}>
        <div class="meta">
          <strong>{message.author}</strong>
          <span>{message.timestamp}</span>
        </div>
        <p>{message.body}</p>
      </article>
    {/each}
  </div>
  <footer>
    <input placeholder="Type a message" />
    <button>Send</button>
  </footer>
</section>

<style>
  .message-view {
    display: grid;
    grid-template-rows: auto 1fr auto;
    height: 100%;
  }
  .messages {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    overflow-y: auto;
    padding: 1rem 0;
  }
  article {
    background: var(--surface-2);
    padding: 0.75rem 1rem;
    border-radius: 0.75rem;
  }
  article.outgoing {
    align-self: flex-end;
    background: var(--primary);
    color: white;
  }
  footer {
    display: flex;
    gap: 0.5rem;
  }
  input {
    flex: 1;
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    padding: 0.5rem 0.75rem;
  }
</style>
