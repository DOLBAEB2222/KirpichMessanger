<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  let email = '';
  let password = '';
  let mode: 'login' | 'register' = 'login';

  const submit = () => {
    dispatch('submit', { email, password, mode });
  };
</script>

<section class="auth-form">
  <header>
    <h2>{mode === 'login' ? 'Login' : 'Register'}</h2>
    <button on:click={() => (mode = mode === 'login' ? 'register' : 'login')}>
      {mode === 'login' ? 'Need an account?' : 'Already have an account?'}
    </button>
  </header>
  <div class="fields">
    <label>
      Email
      <input type="email" bind:value={email} placeholder="name@kirpich.app" />
    </label>
    <label>
      Password
      <input type="password" bind:value={password} />
    </label>
    <button on:click={submit} class="primary">
      {mode === 'login' ? 'Sign in' : 'Create account'}
    </button>
  </div>
</section>

<style>
  .auth-form {
    display: grid;
    gap: 1rem;
    background: var(--surface-2);
    padding: 1.5rem;
    border-radius: 1rem;
  }
  .fields {
    display: grid;
    gap: 0.75rem;
  }
  .primary {
    background: var(--primary);
    color: white;
    border: none;
    padding: 0.75rem 1rem;
    border-radius: 0.75rem;
  }
</style>
