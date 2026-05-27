import axios from 'axios'
import type {
  AppSettings,
  ConnectionStatus,
  CoreSnapshot,
  DiagnosticReport,
  DiagnosticCheck,
  IPCheckResult,
  LogEntry,
  Node,
  RuntimeCapabilities,
  RoutingRule,
  Subscription,
  SubscriptionUpdateResult,
  SystemProxyStatus,
  ValidateCoreResponse,
  ZapretPreset
} from '@/types/api'

import { invoke } from '@tauri-apps/api/core'

export const api = axios.create({
  baseURL: import.meta.env.VITE_RAYFLOW_API_URL ?? 'http://127.0.0.1:8787',
  timeout: 15000
})

let tokenPromise: Promise<string> | null = null

function getSessionToken(): Promise<string> {
  if (!tokenPromise) {
    const envToken = import.meta.env.VITE_RAYFLOW_AUTH_TOKEN
    if (envToken) {
      tokenPromise = Promise.resolve(envToken)
    } else {
      tokenPromise = invoke<string>('get_api_token').catch((err) => {
        console.warn('Failed to retrieve session token via Tauri IPC:', err)
        return ''
      })
    }
  }
  return tokenPromise
}

api.interceptors.request.use(async (config) => {
  try {
    const token = await getSessionToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
  } catch (err) {
    console.error('Error in session token interceptor:', err)
  }
  return config
})

export const rayflowApi = {
  health: () => api.get<{ status: string; service: string }>('/health').then((r) => r.data),

  nodes: () => api.get<Node[]>('/nodes').then((r) => r.data),
  importNode: (payload: { link: string; name?: string }) =>
    api.post<Node>('/nodes/import', payload).then((r) => r.data),
  deleteNode: (id: string) => api.delete(`/nodes/${id}`),
  checkNode: (id: string) => api.post<DiagnosticCheck>(`/nodes/${id}/check`).then((r) => r.data),
  connectNode: (id: string, core = 'sing-box') =>
    api.post<ConnectionStatus>(`/nodes/${id}/connect`, { core }).then((r) => r.data),
  connectNodeWithOptions: (payload: { nodeId: string; core?: string; networkMode?: string }) =>
    api
      .post<ConnectionStatus>(`/nodes/${payload.nodeId}/connect`, {
        core: payload.core,
        networkMode: payload.networkMode
      })
      .then((r) => r.data),
  disconnect: () => api.post<ConnectionStatus>('/connection/disconnect').then((r) => r.data),

  subscriptions: () => api.get<Subscription[]>('/subscriptions').then((r) => r.data),
  createSubscription: (payload: { name: string; url: string; updateInterval: number }) =>
    api.post<Subscription>('/subscriptions', payload).then((r) => r.data),
  updateSubscription: (id: string) =>
    api.post<SubscriptionUpdateResult>(`/subscriptions/${id}/update`).then((r) => r.data),
  deleteSubscription: (id: string) => api.delete(`/subscriptions/${id}`),

  presets: () => api.get<ZapretPreset[]>('/zapret/presets').then((r) => r.data),
  updatePresets: () => api.post('/zapret/presets/update').then((r) => r.data),
  startPreset: (id: string) => api.post(`/zapret/presets/${id}/start`).then((r) => r.data),
  stopZapret: () => api.post('/zapret/stop').then((r) => r.data),
  zapretLogs: () => api.get<LogEntry[]>('/zapret/logs').then((r) => r.data),

  routingRules: () => api.get<RoutingRule[]>('/routing/rules').then((r) => r.data),
  createRoutingRule: (payload: { domain: string; mode: string; enabled: boolean }) =>
    api.post<RoutingRule>('/routing/rules', payload).then((r) => r.data),
  deleteRoutingRule: (id: string) => api.delete(`/routing/rules/${id}`),

  diagnosticsCheck: (payload: { target: string; mode?: string; type?: string }) =>
    api.post<DiagnosticCheck>('/diagnostics/check', payload).then((r) => r.data),
  ipCheck: (payload: { route: string; proxyAddress?: string; proxyProtocol?: string }) =>
    api.post<IPCheckResult>('/diagnostics/ip-check', payload).then((r) => r.data),
  diagnosticsHistory: () => api.get<DiagnosticCheck[]>('/diagnostics/history').then((r) => r.data),
  runtimeCapabilities: () => api.get<RuntimeCapabilities>('/runtime/capabilities').then((r) => r.data),

  settings: () => api.get<AppSettings>('/settings').then((r) => r.data),
  patchSettings: (payload: Partial<AppSettings>) =>
    api.patch<AppSettings>('/settings', payload).then((r) => r.data),

  coreStatus: () =>
    api.get<Record<'singBox' | 'xray' | 'zapret', CoreSnapshot>>('/cores/status').then((r) => r.data),
  validateCore: (payload: { core: string; path: string }) =>
    api.post<ValidateCoreResponse>('/cores/validate', payload).then((r) => r.data),
  logs: (source = '', limit = 200) =>
    api.get<LogEntry[]>('/logs', { params: { source, limit } }).then((r) => r.data),

  connectionStatus: () => api.get<ConnectionStatus>('/connection/status').then((r) => r.data),
  connectionReport: () => api.get<DiagnosticReport>('/connection/report').then((r) => r.data),
  systemProxyStatus: () => api.get<SystemProxyStatus>('/system-proxy/status').then((r) => r.data),
  enableSystemProxy: (proxyServer?: string) =>
    api.post<SystemProxyStatus>('/system-proxy/enable', { proxyServer }).then((r) => r.data),
  disableSystemProxy: () => api.post<SystemProxyStatus>('/system-proxy/disable').then((r) => r.data)
}
