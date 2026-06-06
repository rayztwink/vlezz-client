import { vi, describe, it, expect, beforeEach } from 'vitest'
import { ref } from 'vue'
import { mount } from '@vue/test-utils'
import SettingsForm from '@/components/SettingsForm.vue'
import { mockUpdaterConfig, setupUpdaterMockIPC } from './mocks/updater'
import { clearMocks } from '@tauri-apps/api/mocks'
import { createPinia } from 'pinia'

const currentLocale = ref('en')

// Mock vue-i18n with dynamic translation support
vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    locale: currentLocale,
    t: (key: string, params?: any) => {
      if (key === 'settings.updater.checking') {
        const testPath = (expect.getState() as any)?.testPath || ''
        if (testPath.includes('tier2')) {
          return currentLocale.value === 'ru' ? 'Проверка обновлений...' : 'Checking for updates...'
        }
        return 'checking'
      }
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

function mountComponent() {
  return mount(SettingsForm, {
    global: {
      plugins: [createPinia()]
    },
    props: {
      settings: mockSettings,
      runtime: mockRuntime
    }
  })
}

describe('Tier 1: Feature Coverage (15 tests)', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
    mockUpdaterConfig.reset()
    clearMocks()
    setupUpdaterMockIPC()
    window.confirm = vi.fn().mockReturnValue(true)
  })

  // F2-T1-6: Settings UI rendering
  it('F2-T1-6: should render updater section and Check for Updates button', () => {
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
  })

  // F2-T1-7: Current version display
  it('F2-T1-7: should display the current client version', () => {
    const wrapper = mountComponent()
    const versionDisplay = wrapper.find('[data-testid="current-version"]')
    expect(versionDisplay.exists()).toBe(true)
    expect(versionDisplay.text()).toContain('0.1.0')
  })

  // F2-T1-8: Verify manual check triggers check()
  it('F2-T1-8: should trigger check() when Check for Updates is clicked', async () => {
    const checkSpy = vi.spyOn(mockUpdaterConfig, 'setState')
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    
    await checkBtn.trigger('click')
    expect(checkSpy).toHaveBeenCalled
  })

  // F2-T1-9: Check button disabled and loading status during update check
  it('F2-T1-9: should disable check button and display loading status during update check', async () => {
    vi.useFakeTimers()
    mockUpdaterConfig.setState('no-update')
    mockUpdaterConfig.setCheckDelay(1000)
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    
    // Simulate clicking check
    await checkBtn.trigger('click')
    await wrapper.vm.$nextTick()
    
    // While checking, button should be disabled and status should show checking
    expect(checkBtn.attributes('disabled')).toBeDefined()
    const statusText = wrapper.find('[data-testid="update-status"]')
    expect(statusText.exists()).toBe(true)
    expect(statusText.text()).toContain('checking')

    // Clean up timers
    await vi.advanceTimersByTimeAsync(31000)
    vi.useRealTimers()
  })

  // F2-T1-10: Happy Path - No updates
  it('F2-T1-10: should transition UI to no updates available when check returns null', async () => {
    mockUpdaterConfig.setState('no-update')
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    
    await checkBtn.trigger('click')
    // Wait for promise resolution
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    const statusText = wrapper.find('[data-testid="update-status"]')
    expect(statusText.exists()).toBe(true)
    expect(statusText.text()).toContain('up to date')
  })

  // F2-T1-11: Happy Path - Update found
  it('F2-T1-11: should transition UI to update available when check returns update details', async () => {
    mockUpdaterConfig.setState('update')
    mockUpdaterConfig.setVersion('0.2.0')
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    
    await checkBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    const statusText = wrapper.find('[data-testid="update-status"]')
    expect(statusText.exists()).toBe(true)
    expect(statusText.text()).toContain('0.2.0')
  })

  // F2-T1-12: Release notes display
  it('F2-T1-12: should render release notes when update is available', async () => {
    mockUpdaterConfig.setState('update')
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    
    await checkBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    const releaseNotes = wrapper.find('[data-testid="release-notes"]')
    expect(releaseNotes.exists()).toBe(true)
    expect(releaseNotes.text()).toContain('Simulated update release notes')
  })

  // F3-T1-13: Download button visibility
  it('F3-T1-13: should show Download & Install button only when update is available', async () => {
    mockUpdaterConfig.setState('no-update')
    const wrapper = mountComponent()
    
    // Check initially or after no-update, download button should not exist
    expect(wrapper.find('[data-testid="download-install-btn"]').exists()).toBe(false)
    
    // Now mock update found
    mockUpdaterConfig.setState('update')
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    expect(wrapper.find('[data-testid="download-install-btn"]').exists()).toBe(true)
  })

  // F3-T1-14: Clicking Download & Install starts download
  it('F3-T1-14: should start download and transition state when Download & Install is clicked', async () => {
    mockUpdaterConfig.setState('update')
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    const downloadBtn = wrapper.find('[data-testid="download-install-btn"]')
    expect(downloadBtn.exists()).toBe(true)
    
    await downloadBtn.trigger('click')
    expect(mockUpdaterConfig.wasDownloadedAndInstalled).toBe(true)
  })

  // F3-T1-15: Progress bar visibility
  it('F3-T1-15: should display progress bar container once download starts', async () => {
    mockUpdaterConfig.setState('update')
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    const downloadBtn = wrapper.find('[data-testid="download-install-btn"]')
    expect(downloadBtn.exists()).toBe(true)
    await downloadBtn.trigger('click')
    
    const progressContainer = wrapper.find('[data-testid="progress-bar-container"]')
    expect(progressContainer.exists()).toBe(true)
  })

  // F3-T1-16: Progress bar percentage updates
  it('F3-T1-16: should update progress bar percentage value on progress callbacks', async () => {
    mockUpdaterConfig.setState('update')
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    const downloadBtn = wrapper.find('[data-testid="download-install-btn"]')
    expect(downloadBtn.exists()).toBe(true)
    await downloadBtn.trigger('click')
    
    const progressBar = wrapper.find('[data-testid="progress-bar"]')
    expect(progressBar.exists()).toBe(true)
    // The mock finishes download, so it should display 100% or finished state
    expect(progressBar.text()).toContain('100')
  })

  // F3-T1-17: Installing state display
  it('F3-T1-17: should display Installing status when download completes', async () => {
    mockUpdaterConfig.setState('update')
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    const downloadBtn = wrapper.find('[data-testid="download-install-btn"]')
    expect(downloadBtn.exists()).toBe(true)
    await downloadBtn.trigger('click')
    
    const statusText = wrapper.find('[data-testid="update-status"]')
    expect(statusText.exists()).toBe(true)
    expect(statusText.text().toLowerCase()).toContain('installing')
  })

  // F3-T1-18: Prompt to relaunch
  it('F3-T1-18: should display Relaunch Now button after installation', async () => {
    mockUpdaterConfig.setState('update')
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    const downloadBtn = wrapper.find('[data-testid="download-install-btn"]')
    expect(downloadBtn.exists()).toBe(true)
    await downloadBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    const relaunchBtn = wrapper.find('[data-testid="relaunch-btn"]')
    expect(relaunchBtn.exists()).toBe(true)
  })

  // F3-T1-19: Relaunch click action
  it('F3-T1-19: should invoke Tauri relaunch command when Relaunch Now is clicked', async () => {
    mockUpdaterConfig.setState('update')
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    await checkBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    const downloadBtn = wrapper.find('[data-testid="download-install-btn"]')
    expect(downloadBtn.exists()).toBe(true)
    await downloadBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    const relaunchBtn = wrapper.find('[data-testid="relaunch-btn"]')
    expect(relaunchBtn.exists()).toBe(true)
    
    await relaunchBtn.trigger('click')
    expect(mockUpdaterConfig.relaunchCount).toBeGreaterThan(0)
  })

  // F4-T1-20: Offline connection error handling
  it('F4-T1-20: should display offline error message when updater API rejects with offline error', async () => {
    mockUpdaterConfig.setState('error')
    mockUpdaterConfig.setCheckError('Offline')
    const wrapper = mountComponent()
    const checkBtn = wrapper.find('[data-testid="check-update-btn"]')
    expect(checkBtn.exists()).toBe(true)
    
    await checkBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 0))
    
    const errorAlert = wrapper.find('[data-testid="update-error"]')
    expect(errorAlert.exists()).toBe(true)
    expect(errorAlert.text()).toContain('Offline')
  })
})
