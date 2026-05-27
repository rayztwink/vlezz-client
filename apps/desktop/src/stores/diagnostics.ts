import { defineStore } from 'pinia'
import { ref } from 'vue'
import { rayflowApi } from '@/services/api'
import type { DiagnosticCheck, IPCheckResult, IPCheckRoute, ProxyProtocol, RuntimeCapabilities } from '@/types/api'

export const useDiagnosticsStore = defineStore('diagnostics', () => {
  const history = ref<DiagnosticCheck[]>([])
  const ipChecks = ref<Record<string, IPCheckResult | null>>({
    direct: null,
    rayflow_proxy: null,
    tun: null
  })
  const runtime = ref<RuntimeCapabilities | null>(null)
  const ipLoadingRoute = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load() {
    loading.value = true
    error.value = null
    try {
      history.value = await rayflowApi.diagnosticsHistory()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load diagnostics'
    } finally {
      loading.value = false
    }
  }

  async function loadRuntime() {
    runtime.value = await rayflowApi.runtimeCapabilities()
    return runtime.value
  }

  async function check(target: string, type = 'tcp', mode = 'direct') {
    const result = await rayflowApi.diagnosticsCheck({ target, type, mode })
    history.value = [result, ...history.value]
    return result
  }

  async function checkIP(route: IPCheckRoute, proxyAddress?: string, proxyProtocol?: ProxyProtocol) {
    ipLoadingRoute.value = route
    try {
      const result = await rayflowApi.ipCheck({ route, proxyAddress, proxyProtocol })
      ipChecks.value = { ...ipChecks.value, [route]: result }
      return result
    } finally {
      ipLoadingRoute.value = null
    }
  }

  return { history, ipChecks, runtime, ipLoadingRoute, loading, error, load, loadRuntime, check, checkIP }
})
