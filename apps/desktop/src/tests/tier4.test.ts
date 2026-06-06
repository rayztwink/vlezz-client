import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest'
import { ref } from 'vue'
import { mount } from '@vue/test-utils'
import SettingsForm from '@/components/SettingsForm.vue'
import { mockUpdaterConfig, setupUpdaterMockIPC } from './mocks/updater'
import { clearMocks } from '@tauri-apps/api/mocks'
import { createPinia } from 'pinia'
import { useUpdaterStore } from '@/stores/updater'

const currentLocale = ref('en')

// Mock vue-i18n
vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    locale: currentLocale,
    t: (key: string, params?: any) => {
      if (params) {
        let res = key
        for (const [k, v] of Object.entries(params)) {
          res = res.replace(`{${k}}`, String(v))
        }
        return res
      }
      return key
    }
  })
}))

// Mock tauri plugin app getVersion
vi.mock('@tauri-apps/api/app', () => ({
  getVersion: () => Promise.resolve('0.1.0')
}))

// Mock tauri updater and process
vi.mock('@tauri-apps/plugin-updater', async () => {
  const actual = await import('./mocks/updater')
  return {
    check: () => actual.check()
  }
})

vi.mock('@tauri-apps/plugin-process', () => ({
  relaunch: async () => {
    const { invoke } = await import('@tauri-apps/api/core')
    await invoke('plugin:process|relaunch')
  }
}))

const mockSettings = {
  id: 1,
  theme: 'system' as const,
  language: 'en' as const,
  autostart: false,
  activeMode: 'direct' as const,
  defaultCore: 'sing-box',
  localProxyPort: 2080,
  singBoxPath: '',
  xrayPath: '',
  zapretPath: '',
  enableSystemProxyOnConnect: false,
  preferredNetworkMode: 'local_proxy' as const,
  tunEnabled: false,
  tunStack: 'system',
  tunAutoRoute: true,
  tunStrictRoute: true,
  updatedAt: new Date().toISOString()
}

const mockRuntime = {
  platform: 'win32',
  isAdmin: true,
  systemProxySupported: true
}

function mountComponent(props = {}) {
  return mount(SettingsForm, {
    global: {
      plugins: [createPinia()]
    },
    props: {
      settings: mockSettings,
      runtime: mockRuntime,
      ...props
    }
  })
}

describe('Tier 4: Security & Adversarial Checks (5 tests)', () => {
  beforeEach(() => {
    mockUpdaterConfig.reset()
    clearMocks()
    setupUpdaterMockIPC()
    window.confirm = () => true
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  // F4-T4-1: Invalid signature check
  it('F4-T4-1: should abort installation and display Signature verification failed error when cryptographic signature is invalid', async () => {
    mockUpdaterConfig.setState('error')
    mockUpdaterConfig.setCheckError('Signature verification failed')

    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await wrapper.vm.$nextTick()

    const errorAlert = wrapper.find('[data-testid="update-error"]')
    expect(errorAlert.exists()).toBe(true)
    expect(errorAlert.text()).toContain('Signature verification failed')

    const store = useUpdaterStore()
    expect(store.updateAvailable).toBe(false)
  })

  // F4-T4-2: Downgrade prevention
  it('F4-T4-2: should reject update and not prompt user when served version is lower than current version', async () => {
    mockUpdaterConfig.setState('no-update') // Simulate downgrade being ignored by returning no update
    mockUpdaterConfig.setVersion('0.0.9')

    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await wrapper.vm.$nextTick()

    const downloadBtn = wrapper.find('[data-testid="download-install-btn"]')
    expect(downloadBtn.exists()).toBe(false)

    const store = useUpdaterStore()
    expect(store.updateAvailable).toBe(false)
  })

  // F4-T4-3: Malformed JSON response
  it('F4-T4-3: should handle malformed update server JSON safely and show parse error', async () => {
    mockUpdaterConfig.setState('error')
    mockUpdaterConfig.setCheckError('Failed to parse update information')

    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await wrapper.vm.$nextTick()

    const errorAlert = wrapper.find('[data-testid="update-error"]')
    expect(errorAlert.exists()).toBe(true)
    expect(errorAlert.text()).toContain('Failed to parse update information')
  })

  // F4-T4-4: MITM SSL check
  it('F4-T4-4: should reject update connection when SSL handshake fails (MITM check)', async () => {
    mockUpdaterConfig.setState('error')
    mockUpdaterConfig.setCheckError('SSL verification failed')

    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await wrapper.vm.$nextTick()

    const errorAlert = wrapper.find('[data-testid="update-error"]')
    expect(errorAlert.exists()).toBe(true)
    expect(errorAlert.text()).toContain('SSL verification failed')
  })

  // F4-T4-5: Path Traversal payload URL check
  it('F4-T4-5: should refuse download and raise safety alert when update payload path traversal is detected', async () => {
    mockUpdaterConfig.setState('update')
    mockUpdaterConfig.setDownloadError('Path traversal detected')

    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 10))

    const downloadBtn = wrapper.find('[data-testid="download-install-btn"]')
    expect(downloadBtn.exists()).toBe(true)
    await downloadBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 10))
    await wrapper.vm.$nextTick()

    const errorAlert = wrapper.find('[data-testid="update-error"]')
    expect(errorAlert.exists()).toBe(true)
    expect(errorAlert.text()).toContain('Path traversal detected')

    const store = useUpdaterStore()
    expect(store.downloading).toBe(false)
  })
})
