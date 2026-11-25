import { es } from '@ms-formularios/forms-renderer'
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import App from '../App.vue'
import { router } from '../router'
import HomeView from './HomeView.vue'

describe('HomeView', () => {
  it('muestra el título de la app', () => {
    const wrapper = mount(HomeView)
    expect(wrapper.get('h1').text()).toBe('Formularios')
  })

  it('renderiza App con el router principal y diccionario base', async () => {
    router.push('/')
    await router.isReady()
    const wrapper = mount(App, {
      global: {
        plugins: [router],
      },
    })
    expect(wrapper.find('main.app').exists()).toBe(true)
    expect(es).toBeDefined()
  })
})
