import { defineStore } from 'pinia'
import { ref } from 'vue'
import { rayflowApi } from '@/services/api'
import type { ConnectionStatus, DiagnosticReport, LogEntry, NetworkMode } from '@/types/api'

export const useConnectionStore = defineStore('connection', () => {
  const status = ref<ConnectionStatus | null>(null)
  const logs = ref<LogEntry[]>([])
  const report = ref<DiagnosticReport | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load() {
    try {
      const [connectionStatus, logEntries] = await Promise.all([
        rayflowApi.connectionStatus(),
        rayflowApi.logs('', 200)
      ])
      status.value = connectionStatus
      logs.value = logEntries
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load connection status'
    }
  }

  async function connect(nodeId: string, core: string, networkMode: NetworkMode) {
    loading.value = true
    error.value = null
    try {
      status.value = await rayflowApi.connectNodeWithOptions({ nodeId, core, networkMode })
      await load()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to connect'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function disconnect() {
    loading.value = true
    error.value = null
    try {
      status.value = await rayflowApi.disconnect()
      await load()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to disconnect'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function buildReport() {
    report.value = await rayflowApi.connectionReport()
    return report.value
  }

  return { status, logs, report, loading, error, load, connect, disconnect, buildReport }
})

