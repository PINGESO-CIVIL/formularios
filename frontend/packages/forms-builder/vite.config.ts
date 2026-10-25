import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [vue()],
  build: {
    lib: { entry: 'src/index.ts', formats: ['es'], fileName: 'index' },
    rollupOptions: {
      external: ['vue', '@ms-formularios/forms-core', '@ms-formularios/forms-renderer'],
    },
  },
  test: { environment: 'happy-dom', passWithNoTests: true },
})
