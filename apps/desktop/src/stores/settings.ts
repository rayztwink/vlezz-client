import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { setLocale } from '@/i18n'
import { rayflowApi } from '@/services/api'
import type { ActiveMode, AppSettings } from '@/types/api'

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<AppSettings | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const activeMode = computed<ActiveMode>(() => settings.value?.activeMode ?? 'direct')

  async function load() {
    loading.value = true
    error.value = null
    try {
      settings.value = await rayflowApi.settings()
      applyTheme(settings.value.theme)
      setLocale(settings.value.language)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load settings'
    } finally {
      loading.value = false
    }
  }

  async function setMode(mode: ActiveMode) {
    settings.value = await rayflowApi.patchSettings({ activeMode: mode })
  }

  async function patch(payload: Partial<AppSettings>) {
    error.value = null
    try {
      settings.value = await rayflowApi.patchSettings(payload)
      if (settings.value?.theme) {
        applyTheme(settings.value.theme)
      }
      if (settings.value?.language) {
        setLocale(settings.value.language)
      }
    } catch (err: any) {
      console.error('Failed to patch settings:', err)
      let errMsg = err instanceof Error ? err.message : 'Failed to update settings'
      if (err?.response?.status === 401) {
        errMsg = 'Unauthorized (401): Please restart the rayflowd backend process to sync tokens.'
      }
      error.value = errMsg
      throw err
    }
  }

  function applyTheme(theme: AppSettings['theme']) {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    const shouldDark = theme === 'dark' || (theme === 'system' && prefersDark)
    document.documentElement.classList.toggle('dark', shouldDark)
  }

  return { settings, loading, error, activeMode, load, setMode, patch }
})
