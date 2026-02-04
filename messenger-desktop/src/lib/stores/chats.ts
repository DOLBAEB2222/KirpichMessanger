import { writable } from 'svelte/store';

export type ChatItem = {
  id: string;
  title: string;
  lastMessage?: string;
  unreadCount: number;
};

export const chatList = writable<ChatItem[]>([]);
export const activeChatId = writable<string | null>(null);
