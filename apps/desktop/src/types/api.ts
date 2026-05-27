export type ActiveMode = 'direct' | 'proxy' | 'zapret' | 'hybrid' | 'smart'

export interface AppSettings {
  id: number
  theme: 'system' | 'light' | 'dark'
  language: 'system' | 'ru' | 'en'
  autostart: boolean
  activeMode: ActiveMode
  defaultCore: 'sing-box' | 'xray' | string
  localProxyPort: number
  singBoxPath: string
  xrayPath: string
  zapretPath: string
  enableSystemProxyOnConnect: boolean
  preferredNetworkMode: NetworkMode
  tunEnabled: boolean
  tunStack: string
  tunAutoRoute: boolean
  tunStrictRoute: boolean
  updatedAt: string
}

export type NetworkMode = 'local_proxy' | 'system_proxy' | 'tun'
export type IPCheckRoute = 'direct' | 'rayflow_proxy' | 'tun'
export type ProxyProtocol = 'http' | 'socks5'

export interface Node {
  id: string
  name: string
  protocol: string
  address: string
  port: number
  uuid: string
  security: string
  transport: string
  rawLink?: string
  latencyMs?: number
  country?: string
  createdAt: string
}

export interface Subscription {
  id: string
  name: string
  url: string
  updateInterval: number
  lastUpdateAt?: string
  createdAt: string
}

export interface SubscriptionUpdateResult {
  status: string
  id: string
  total: number
  imported: number
  skipped: number
  failed: number
}

export interface ZapretPreset {
  id: string
  name: string
  source: string
  command: string
  description?: string
  isActive: boolean
  updatedAt: string
}

export interface RoutingRule {
  id: string
  domain: string
  mode: ActiveMode
  enabled: boolean
}

export interface DiagnosticCheck {
  id: string
  target: string
  mode: ActiveMode | string
  status: 'ok' | 'failed' | string
  latencyMs?: number
  error?: string
  checkedAt: string
}

export interface IPCheckResult {
  route: IPCheckRoute | string
  status: 'ok' | 'failed' | string
  ip?: string
  country?: string
  provider: string
  latencyMs: number
  checkedAt: string
  error?: string
}

export interface RuntimeCapabilities {
  platform: string
  isAdmin: boolean
  systemProxySupported: boolean
}

export interface LogEntry {
  id: string
  source: string
  level: string
  message: string
  createdAt: string
}

export interface CoreSnapshot {
  id: string
  name: string
  status: 'stopped' | 'running' | 'failed' | string
  pid?: number
  startedAt?: string
  error?: string
}

export interface ValidateCoreResponse {
  ok: boolean
  core: string
  path: string
  version?: string
  error?: string
}

export interface SystemProxyStatus {
  supported: boolean
  proxyEnable: boolean
  proxyServer: string
  proxyOverride: string
  enabledByRayflow: boolean
  currentProxyServer: string
  error?: string
}

export interface ConnectionStatus {
  id: number
  activeMode: ActiveMode | string
  selectedNodeId?: string
  selectedNodeName?: string
  selectedCore: string
  networkMode: NetworkMode | string
  localProxyAddress: string
  status: 'disconnected' | 'connecting' | 'connected' | 'failed' | string
  lastError?: string
  updatedAt: string
  processStatus: CoreSnapshot
  systemProxy: SystemProxyStatus
}

export interface DiagnosticReport {
  status: ConnectionStatus
  logs: LogEntry[]
  checks: Record<string, unknown>
}
