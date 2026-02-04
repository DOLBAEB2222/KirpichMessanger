import { writable } from 'svelte/store';

export type UserProfile = {
  id: string;
  name: string;
  email: string;
};

export const authToken = writable<string | null>(null);
export const userProfile = writable<UserProfile | null>(null);
