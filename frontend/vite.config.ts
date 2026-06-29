import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
	plugins: [svelte()],
	base: '/',
	build: {
		outDir: '../src/web/dist',
		emptyOutDir: true,
	},
	server: {
		proxy: {
			'/api': 'http://localhost:8765',
			'/video': 'http://localhost:8765',
		},
	},
});
