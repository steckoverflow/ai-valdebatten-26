import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  build: {
    // Build straight into the Go embed directory so `go:embed all:dist` picks
    // up the production bundle. emptyOutDir clears stale assets each build.
    outDir: '../internal/web/dist',
    emptyOutDir: true,
  },
  server: {
    // During development, forward the WebSocket (and health check) to the Go
    // backend so the browser sees a single origin and we avoid CORS entirely.
    proxy: {
      '/ws': {
        target: 'http://localhost:8080',
        ws: true,
        changeOrigin: true,
      },
      '/healthz': 'http://localhost:8080',
    },
  },
})
