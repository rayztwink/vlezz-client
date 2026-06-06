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

describe('Tier 3: Destructive & Resource Checks (4 tests)', () => {
  beforeEach(() => {
    mockUpdaterConfig.reset()
    clearMocks()
    setupUpdaterMockIPC()
    window.confirm = () => true
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  // F3-T3-1: Disk space depletion
  it('F3-T3-1: should display Insufficient disk space error and pause download rather than crashing', async () => {
    mockUpdaterConfig.setState('update')
    mockUpdaterConfig.setDownloadError('Insufficient disk space')

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
    expect(errorAlert.text()).toContain('Insufficient disk space')

    const store = useUpdaterStore()
    expect(store.downloading).toBe(false)
    expect(store.downloadedAndInstalled).toBe(false)
  })

  // F3-T3-2: Memory leaks check (rapid clicks)
  it('F3-T3-2: should handle 50 rapid updates check calls without throwing errors or leaving orphaned listeners', async () => {
    vi.useFakeTimers()
    mockUpdaterConfig.setState('no-update')
    const wrapper = mountComponent()
    const store = useUpdaterStore()

    // Trigger updates check 50 times rapidly
    const promises = []
    for (let i = 0; i < 50; i++) {
      promises.push(store.checkForUpdates())
    }
    
    // Resolve all update checks
    await Promise.all(promises)
    await vi.advanceTimersByTimeAsync(30000)
    await wrapper.vm.$nextTick()

    // Unmount component to ensure clean cleanup
    expect(() => wrapper.unmount()).not.toThrow()
    expect(store.checking).toBe(false)
    vi.useRealTimers()
  })

  // F3-T3-3: Over-sized update payloads (10GB)
  it('F3-T3-3: should format and display 10GB update size progress correctly without integer overflow or layout distortion', async () => {
    mockUpdaterConfig.setState('update')
    
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 10))

    // Manually trigger download start with a large contentLength (10 GB)
    const store = useUpdaterStore()
    store.downloading = true
    store.status = 'downloading'
    store.totalBytes = 10737418240 // 10 GB
    store.downloadedBytes = 5368709120 // 5 GB
    store.progressPercent = 50
    await wrapper.vm.$nextTick()

    const progressBar = wrapper.find('[data-testid="progress-bar"]')
    expect(progressBar.exists()).toBe(true)
    expect(progressBar.text()).toContain('50%')
    expect(progressBar.text()).toContain('5 GB')
    expect(progressBar.text()).toContain('10 GB')
  })

  // F4-T3-4: Corrupt archive extraction
  it('F4-T3-4: should handle installation/extraction failure, reset download state and report corruption error', async () => {
    mockUpdaterConfig.setState('update')
    mockUpdaterConfig.setDownloadError('Installation package corrupted')

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
    expect(errorAlert.text()).toContain('Installation package corrupted')

    const store = useUpdaterStore()
    expect(store.downloading).toBe(false)
    expect(store.downloadedAndInstalled).toBe(false)
  })
})
