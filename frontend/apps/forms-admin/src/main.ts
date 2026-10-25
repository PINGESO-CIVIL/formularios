import { createPinia } from 'pinia'
import { createApp } from 'vue'

import '@ms-formularios/forms-renderer/theme.css'
import App from './App.vue'
import { router } from './router'

async function enableMocks(): Promise<void> {
  if (!import.meta.env.DEV || import.meta.env.VITE_USE_MOCKS !== 'true') return
  const { worker } = await import('./mocks/browser')
  await worker.start({ onUnhandledFrame: 'bypass' })
}

await enableMocks()

createApp(App).use(createPinia()).use(router).mount('#app')
