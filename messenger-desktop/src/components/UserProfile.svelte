<script lang="ts">
  export let name = 'Guest';
  export let email = 'guest@kirpich.app';
</script>

<section class="profile">
  <div class="avatar">{name[0]}</div>
  <div>
    <h3>{name}</h3>
    <p>{email}</p>
  </div>
  <button>Edit</button>
</section>

<style>
  .profile {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background: var(--surface-2);
    border-radius: 1rem;
  }
  .avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: var(--primary);
    color: white;
    display: grid;
    place-items: center;
    font-size: 1.25rem;
  }
  button {
    margin-left: auto;
  }
</style>
