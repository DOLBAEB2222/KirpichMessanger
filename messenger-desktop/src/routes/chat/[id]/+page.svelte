<script>
  import { page } from '$app/stores';
  $: chatId = $page.params.id;
</script>

<h1>Chat {chatId}</h1>
