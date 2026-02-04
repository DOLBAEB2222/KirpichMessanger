import { writable } from 'svelte/store';

export const messagesStore = writable({
  chats: [],
  currentChatId: null,
  messages: {}
});
