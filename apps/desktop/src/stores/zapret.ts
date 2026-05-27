import { defineStore } from 'pinia'
import { ref } from 'vue'
import { rayflowApi } from '@/services/api'
import type { LogEntry, ZapretPreset } from '@/types/api'

export const useZapretStore = defineStore('zapret', () => {
  const presets = ref<ZapretPreset[]>([])
  const logs = ref<LogEntry[]>([])
  const loading = ref(false)

  async function load() {
    loading.value = true
    try {
      presets.value = await rayflowApi.presets()
      logs.value = await rayflowApi.zapretLogs()
    } finally {
      loading.value = false
    }
  }

  async function updatePresets() {
    await rayflowApi.updatePresets()
    await load()
  }

  async function startPreset(id: string) {
    await rayflowApi.startPreset(id)
    await load()
  }

  async function stop() {
    await rayflowApi.stopZapret()
    await load()
  }

  return { presets, logs, loading, load, updatePresets, startPreset, stop }
})

