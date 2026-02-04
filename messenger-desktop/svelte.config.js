import adapter from '@sveltejs/adapter-static';

const config = {
  kit: {
    adapter: adapter({
      fallback: 'index.html'
    }),
    alias: {
      $components: 'src/components',
      $lib: 'src/lib',
      $stores: 'src/stores'
    }
  }
};

export default config;
