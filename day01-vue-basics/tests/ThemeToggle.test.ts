import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it } from 'vitest'
import ThemeToggle from '../src/components/ThemeToggle.vue'

const html = document.documentElement

describe('ThemeToggle', () => {
  beforeEach(() => {
    localStorage.clear()
    html.removeAttribute('data-theme')
  })

  it('เริ่มต้นเป็น system: ไม่มี data-theme', () => {
    mount(ThemeToggle)
    expect(html.hasAttribute('data-theme')).toBe(false)
  })

  it('เลือกมืด -> ตั้ง data-theme และจำใน localStorage', async () => {
    const w = mount(ThemeToggle)
    await w.findAll('button')[2].trigger('click')
    expect(html.getAttribute('data-theme')).toBe('dark')
    expect(localStorage.getItem('claim-list-theme')).toBe('dark')
  })

  it('กลับเป็น system -> เอา data-theme และค่าที่จำออก', async () => {
    const w = mount(ThemeToggle)
    await w.findAll('button')[1].trigger('click')
    await w.findAll('button')[0].trigger('click')
    expect(html.hasAttribute('data-theme')).toBe(false)
    expect(localStorage.getItem('claim-list-theme')).toBeNull()
  })

  it('อ่านค่าที่จำไว้ตอน mount', () => {
    localStorage.setItem('claim-list-theme', 'light')
    mount(ThemeToggle)
    expect(html.getAttribute('data-theme')).toBe('light')
  })
})
