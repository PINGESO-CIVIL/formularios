import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import HomeView from './HomeView.vue'

describe('HomeView', () => {
  it('muestra el título de la app', () => {
    const wrapper = mount(HomeView)
    expect(wrapper.get('h1').text()).toBe('Formularios')
  })
})
