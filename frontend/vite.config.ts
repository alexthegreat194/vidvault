import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

const backend = process.env.VIDVAULT_BACKEND ?? 'http://localhost:8765';

export default defineConfig({
	plugins: [svelte()],
	base: '/',
	build: {
		outDir: '../src/web/dist',
		emptyOutDir: true,
	},
	server: {
		proxy: {
			'/api': backend,
			'/video': backend,
			'/favicon.svg': backend,
		},
	},
});
