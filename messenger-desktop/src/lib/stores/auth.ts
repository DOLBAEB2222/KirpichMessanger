import { writable } from 'svelte/store';

export const authStore = writable({
  user: null,
  token: null,
  isAuthenticated: false
});

export const login = async (username: string, password: string) => {
  // TODO: Implement login
};

export const logout = () => {
  authStore.set({
    user: null,
    token: null,
    isAuthenticated: false
  });
};
