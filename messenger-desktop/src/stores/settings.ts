import { writable } from 'svelte/store';

export type SettingsState = {
  notificationsEnabled: boolean;
  autoStart: boolean;
  compactMode: boolean;
};

export const settings = writable<SettingsState>({
  notificationsEnabled: true,
  autoStart: false,
  compactMode: false
});
