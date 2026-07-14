import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';

const apiProxy = {
	'^/today\\.v1\\.': 'http://localhost:8081'
};

export default defineConfig({
	plugins: [svelte()],
	server: {
		proxy: apiProxy
	},
	preview: {
		proxy: apiProxy
	}
});
