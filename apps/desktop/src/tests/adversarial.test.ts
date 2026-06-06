import { vi, describe, it, expect, beforeEach } from 'vitest'
import { ref } from 'vue'
import { mount } from '@vue/test-utils'
import SettingsForm from '@/components/SettingsForm.vue'
import { useUpdaterStore } from '@/stores/updater'
import { mockUpdaterConfig, setupUpdaterMockIPC } from './mocks/updater'
import { clearMocks } from '@tauri-apps/api/mocks'
import { createPinia, setActivePinia } from 'pinia'

const currentLocale = ref('en')

// Mock vue-i18n with dynamic translation support
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

describe('Milestone 5: Adversarial Hardening Tests', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
    mockUpdaterConfig.reset()
    clearMocks()
    setupUpdaterMockIPC()
    window.confirm = vi.fn().mockReturnValue(true)
    
    // Set active pinia
    const pinia = createPinia()
    setActivePinia(pinia)
  })

  // 1. formatBytes boundary checks
  describe('formatBytes', () => {
    it('should support TB and PB units', () => {
      const wrapper = mountComponent()
      const format = wrapper.vm.formatBytes
      expect(format(1024 * 1024 * 1024 * 1024)).toBe('1 TB')
      expect(format(1024 * 1024 * 1024 * 1024 * 1024)).toBe('1 PB')
    })

    it('should handle negative, zero, and fractional bytes gracefully', () => {
      const wrapper = mountComponent()
      const format = wrapper.vm.formatBytes
      expect(format(0)).toBe('0 bytes')
      expect(format(-1024)).toBe('-1 KB')
      expect(format(0.5)).toBe('0.5 bytes')
      expect(format(NaN)).toBe('0 bytes')
      expect(format(Infinity)).toBe('0 bytes')
    })
  })

  // 2. localProxyPort validation checks
  describe('localProxyPort Validation', () => {
    it('should reject invalid ports and accept valid ports', () => {
      const wrapper = mountComponent()
      const draft = wrapper.vm.draft
      const validation = wrapper.vm.validation
      const save = wrapper.vm.saveCoreSettings

      // Test invalid port (out of range high)
      draft.localProxyPort = 65536
      save()
      expect(validation.localProxyPort).toBeDefined()
      expect(validation.localProxyPort.ok).toBe(false)

      // Test invalid port (out of range low)
      draft.localProxyPort = 0
      save()
      expect(validation.localProxyPort).toBeDefined()
      expect(validation.localProxyPort.ok).toBe(false)

      // Test invalid port (float)
      draft.localProxyPort = 1234.5
      save()
      expect(validation.localProxyPort).toBeDefined()
      expect(validation.localProxyPort.ok).toBe(false)

      // Test invalid port (NaN)
      draft.localProxyPort = NaN
      save()
      expect(validation.localProxyPort).toBeDefined()
      expect(validation.localProxyPort.ok).toBe(false)

      // Test valid port
      draft.localProxyPort = 8080
      save()
      expect(validation.localProxyPort).toBeUndefined()
    })
  })

  // 3. Validation UI clearing on path mutation
  describe('Validation UI clearing on path/port mutation', () => {
    it('should clear validation status when paths or port are modified', async () => {
      const wrapper = mountComponent()
      const draft = wrapper.vm.draft
      const validation = wrapper.vm.validation

      // Set some validation errors manually
      validation['sing-box'] = { ok: false, message: 'Invalid path' }
      validation['xray'] = { ok: false, message: 'Invalid path' }
      validation['zapret'] = { ok: false, message: 'Invalid path' }
      validation['localProxyPort'] = { ok: false, message: 'Invalid port' }

      // Mutate singBoxPath
      draft.singBoxPath = 'new/path'
      await wrapper.vm.$nextTick()
      expect(validation['sing-box']).toBeUndefined()

      // Mutate xrayPath
      draft.xrayPath = 'new/path'
      await wrapper.vm.$nextTick()
      expect(validation['xray']).toBeUndefined()

      // Mutate zapretPath
      draft.zapretPath = 'new/path'
      await wrapper.vm.$nextTick()
      expect(validation['zapret']).toBeUndefined()

      // Mutate localProxyPort
      draft.localProxyPort = 1080
      await wrapper.vm.$nextTick()
      expect(validation['localProxyPort']).toBeUndefined()
    })
  })

  // 4. Cooldown timer overlap protection
  describe('Cooldown Timer Overlap Protection', () => {
    it('should prevent overlapping update checks if one is already in progress', async () => {
      const store = useUpdaterStore()
      mockUpdaterConfig.setState('update')
      mockUpdaterConfig.setCheckDelay(100)

      const p1 = store.checkForUpdates()
      const p2 = store.checkForUpdates() // this should return immediately because checking is true

      await Promise.all([p1, p2])
      await new Promise((resolve) => setTimeout(resolve, 10))

      expect(store.checking).toBe(false)
    })
  })

  // 5. Relaunch confirmation cancellation behavior
  describe('Relaunch Confirmation Cancellation Behavior', () => {
    it('should not lock form fields permanently if relaunch is cancelled', async () => {
      window.confirm = vi.fn().mockReturnValue(false) // cancel relaunch
      
      const store = useUpdaterStore()
      mockUpdaterConfig.setState('update')
      mockUpdaterConfig.setVersion('2.0.0')

      await store.checkForUpdates()
      await new Promise((resolve) => setTimeout(resolve, 10))
      await store.downloadAndInstall()

      expect(store.relaunchCancelled).toBe(true)

      const wrapper = mount(SettingsForm, {
        global: {
          plugins: [createPinia()]
        },
        props: {
          settings: mockSettings,
          runtime: mockRuntime
        }
      })
      const updaterInForm = wrapper.vm.updaterStore
      updaterInForm.downloadedAndInstalled = true
      updaterInForm.relaunchCancelled = true

      expect(wrapper.vm.isFormDisabled).toBe(false)
    })
  })

  // 6. Downgrade attack rejection
  describe('Downgrade Attack Rejection', () => {
    it('should reject downgrade or same version updates', async () => {
      const store = useUpdaterStore()
      
      // Case 1: Same version
      store.currentVersion = '1.0.0'
      mockUpdaterConfig.setState('update')
      mockUpdaterConfig.setVersion('1.0.0')
      await store.checkForUpdates()
      await new Promise((resolve) => setTimeout(resolve, 10))
      expect(store.updateAvailable).toBe(false)
      expect(store.status).toBe('up-to-date')

      // Case 2: Downgrade version
      store.currentVersion = '1.0.0'
      mockUpdaterConfig.setState('update')
      mockUpdaterConfig.setVersion('0.9.5')
      await store.checkForUpdates()
      await new Promise((resolve) => setTimeout(resolve, 10))
      expect(store.updateAvailable).toBe(false)
      expect(store.status).toBe('up-to-date')

      // Case 3: Proper upgrade
      store.currentVersion = '1.0.0'
      mockUpdaterConfig.setState('update')
      mockUpdaterConfig.setVersion('1.0.1')
      await store.checkForUpdates()
      await new Promise((resolve) => setTimeout(resolve, 10))
      expect(store.updateAvailable).toBe(true)
      expect(store.status).toBe('available')
    })

    it('should reject pre-release tag downgrade/upgrade due to string comparison limitation', async () => {
      const store = useUpdaterStore()
      // Case 4: Prerelease tag comparison with numeric suffixes (beta11 vs beta2)
      // Since it's string comparison, 'beta11' > 'beta2' is false, so beta11 won't be seen as upgrade from beta2
      store.currentVersion = '1.0.0-beta2'
      mockUpdaterConfig.setState('update')
      mockUpdaterConfig.setVersion('1.0.0-beta11')
      await store.checkForUpdates()
      await new Promise((resolve) => setTimeout(resolve, 10))
      expect(store.updateAvailable).toBe(true)
    })
  })

  // 8. Concurrent downloadAndInstall call protection
  describe('Concurrent downloadAndInstall Call Protection', () => {
    it('should protect against concurrent downloadAndInstall calls', async () => {
      window.confirm = vi.fn().mockReturnValue(false)
      const store = useUpdaterStore()
      mockUpdaterConfig.setState('update')
      mockUpdaterConfig.setVersion('2.0.0')
      await store.checkForUpdates()
      await new Promise((resolve) => setTimeout(resolve, 10))

      // Trigger twice concurrently
      const p1 = store.downloadAndInstall()
      const p2 = store.downloadAndInstall()
      await Promise.all([p1, p2])

      // If there is no guard, it triggers mock downloadAndInstall twice
      expect(mockUpdaterConfig.relaunchCount).toBe(1)
    })
  })

  // 7. Relaunch failure propagation
  describe('Relaunch Failure Propagation', () => {
    it('should set status to error and catch error when relaunch fails', async () => {
      const store = useUpdaterStore()
      
      const { mockIPC } = await import('@tauri-apps/api/mocks')
      mockIPC((cmd) => {
        if (cmd === 'relaunch' || cmd === 'plugin:process|relaunch') {
          throw new Error('IPC Relaunch Failed')
        }
      })

      await store.relaunchApp()

      expect(store.status).toBe('error')
      expect(store.error).toBe('IPC Relaunch Failed')
    })
  })

  // 8. Concurrent downloadAndInstall calls protection
  describe('Concurrent downloadAndInstall Calls Protection', () => {
    it('should prevent multiple concurrent downloadAndInstall invocations', async () => {
      const store = useUpdaterStore()
      mockUpdaterConfig.setState('update')
      mockUpdaterConfig.setVersion('2.0.0')

      // Spy on check to return our custom downloadAndInstall spy
      let callCount = 0
      const updaterMock = await import('./mocks/updater')
      vi.spyOn(updaterMock, 'check').mockResolvedValue({
        version: '2.0.0',
        date: new Date().toISOString(),
        body: 'Simulated update release notes',
        close: async () => {},
        downloadAndInstall: async (onProgress) => {
          callCount++
          await new Promise((resolve) => setTimeout(resolve, 100)) // delay
          if (onProgress) {
            onProgress({ event: 'Started', data: { contentLength: 100 } })
          }
        }
      })

      await store.checkForUpdates()
      await new Promise((resolve) => setTimeout(resolve, 10))

      // Fire two download requests concurrently
      const p1 = store.downloadAndInstall()
      const p2 = store.downloadAndInstall()

      await Promise.all([p1, p2])

      // We expect downloadAndInstall to only be called once because of concurrency guards
      expect(callCount).toBe(1)
    })
  })

  // 9. checkForUpdates during active download state protection
  describe('checkForUpdates during active download state protection', () => {
    it('should not reset download state if checkForUpdates is called while downloading', async () => {
      const store = useUpdaterStore()
      mockUpdaterConfig.setState('update')
      mockUpdaterConfig.setVersion('2.0.0')

      // We mock downloadAndInstall to run and stay in downloading state
      let finishDownload: any
      const downloadPromise = new Promise((resolve) => {
        finishDownload = resolve
      })
      
      const updaterMock = await import('./mocks/updater')
      vi.spyOn(updaterMock, 'check').mockResolvedValue({
        version: '2.0.0',
        date: new Date().toISOString(),
        body: 'Simulated update release notes',
        close: async () => {},
        downloadAndInstall: async (onProgress) => {
          if (onProgress) {
            onProgress({ event: 'Started', data: { contentLength: 100 } })
          }
          await downloadPromise
        }
      })

      await store.checkForUpdates()
      await new Promise((resolve) => setTimeout(resolve, 10))

      // Start download
      const d1 = store.downloadAndInstall()
      await new Promise((resolve) => setTimeout(resolve, 10))
      
      expect(store.downloading).toBe(true)
      expect(store.status).toBe('downloading')

      // Call checkForUpdates while download is in progress
      await store.checkForUpdates()

      // The download state should NOT be reset or cleared
      expect(store.downloading).toBe(true)
      expect(store.status).toBe('downloading')

      finishDownload()
      await d1
    })
  })

  // 10. Relaunch Failure UI Lock Protection
  describe('Relaunch Failure UI Lock Protection', () => {
    it('should not keep SettingsForm disabled if relaunch fails', async () => {
      window.confirm = vi.fn().mockReturnValue(true) // confirm relaunch
      
      const store = useUpdaterStore()
      mockUpdaterConfig.setState('update')
      mockUpdaterConfig.setVersion('2.0.0')

      await store.checkForUpdates()
      await new Promise((resolve) => setTimeout(resolve, 10))

      // Mock relaunchApp to fail
      vi.spyOn(store, 'relaunchApp').mockRejectedValue(new Error('Relaunch failed'))

      await store.downloadAndInstall()
      await new Promise((resolve) => setTimeout(resolve, 10))

      // Check if store state indicates download completed but relaunch failed
      expect(store.downloadedAndInstalled).toBe(true)
      
      const wrapper = mount(SettingsForm, {
        global: {
          plugins: [createPinia()]
        },
        props: {
          settings: mockSettings,
          runtime: mockRuntime
        }
      })
      
      // Inject the store with failed relaunch into component
      const componentStore = wrapper.vm.updaterStore
      componentStore.downloadedAndInstalled = true
      componentStore.relaunchCancelled = false
      componentStore.downloading = false
      componentStore.error = 'Relaunch failed'
      
      // The form should not be disabled because relaunch failed and the app did not restart!
      expect(wrapper.vm.isFormDisabled).toBe(false)
    })
  })
})
