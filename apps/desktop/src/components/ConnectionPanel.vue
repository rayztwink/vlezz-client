<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ClipboardDocumentIcon, PowerIcon, SignalIcon } from '@heroicons/vue/24/outline'
import type { ConnectionStatus, NetworkMode, Node } from '@/types/api'

const props = defineProps<{
  nodes: Node[]
  selectedNodeId: string
  status: ConnectionStatus | null
  loading?: boolean
  defaultCore?: string
  defaultNetworkMode?: NetworkMode
}>()

const emit = defineEmits<{
  connect: [nodeId: string, core: string, networkMode: NetworkMode]
  disconnect: []
  select: [nodeId: string]
}>()

const { t } = useI18n()
const selectedCore = ref('sing-box')
const networkMode = ref<NetworkMode>('local_proxy')

const selectedNodeIdModel = computed({
  get: () => props.selectedNodeId || props.nodes[0]?.id || '',
  set: (nodeId: string) => emit('select', nodeId)
})

const selectedNode = computed(() => props.nodes.find((node) => node.id === selectedNodeIdModel.value))
const isConnected = computed(() => props.status?.status === 'connected')
const isBusy = computed(() => props.loading || props.status?.status === 'connecting')
const canConnect = computed(() => Boolean(selectedNodeIdModel.value) && !isBusy.value)
const statusText = computed(() => props.status?.status ?? 'disconnected')
const localProxyAddress = computed(() => props.status?.localProxyAddress || '127.0.0.1:2080')
const endpoint = computed(() => (selectedNode.value ? `${selectedNode.value.address}:${selectedNode.value.port}` : t('client.noProfileSelected')))
const networkModeHint = computed(() => t(`client.modeHints.${networkMode.value}`))

const statusTone = computed(() => {
  if (isConnected.value) {
    return 'connected'
  }
  if (props.status?.status === 'failed') {
    return 'failed'
  }
  return 'idle'
})

const systemProxyLabel = computed(() => {
  const proxy = props.status?.systemProxy
  if (!proxy?.proxyEnable) {
    return t('common.off')
  }
  return proxy.enabledByRayflow ? t('ipCheck.systemProxyOwned') : t('common.external')
})
const systemProxyDetail = computed(() => {
  const proxy = props.status?.systemProxy
  if (!proxy?.proxyEnable) {
    return ''
  }
  return proxy.proxyServer || proxy.currentProxyServer
})
const externalProxyActive = computed(() => {
  const proxy = props.status?.systemProxy
  return Boolean(proxy?.proxyEnable && !proxy.enabledByRayflow)
})

const networkModeLabel = computed(() => {
  switch (networkMode.value) {
    case 'system_proxy':
      return t('modes.systemProxy')
    case 'tun':
      return t('modes.tun')
    default:
      return t('modes.localProxy')
  }
})

function isRuntimeActive(status?: string) {
  return status === 'connected' || status === 'connecting'
}

function applyDefaults(defaultCore = props.defaultCore, defaultNetworkMode = props.defaultNetworkMode) {
  if (defaultCore) {
    selectedCore.value = defaultCore === 'xray-core' ? 'xray' : defaultCore
  }
  if (defaultNetworkMode) {
    networkMode.value = defaultNetworkMode
  }
}

watch(
  () => props.status,
  (status) => {
    if (!status) {
      return
    }
    if (status.selectedCore && status.status !== 'disconnected') {
      selectedCore.value = status.selectedCore === 'xray-core' ? 'xray' : status.selectedCore
    }
    if (status.status !== 'disconnected' && (status.networkMode === 'local_proxy' || status.networkMode === 'system_proxy' || status.networkMode === 'tun')) {
      networkMode.value = status.networkMode
    }
  },
  { immediate: true }
)

watch(
  [() => props.defaultCore, () => props.defaultNetworkMode],
  ([defaultCore, defaultNetworkMode]) => {
    if (isRuntimeActive(props.status?.status)) {
      return
    }
    applyDefaults(defaultCore, defaultNetworkMode)
  },
  { immediate: true }
)

watch(
  () => props.status?.status,
  (status, previousStatus) => {
    if (status === 'disconnected' && previousStatus && previousStatus !== 'disconnected') {
      applyDefaults()
    }
  }
)

function toggleConnection() {
  if (isConnected.value) {
    emit('disconnect')
    return
  }
  if (canConnect.value) {
    emit('connect', selectedNodeIdModel.value, selectedCore.value, networkMode.value)
  }
}

async function copyAddress() {
  await navigator.clipboard.writeText(localProxyAddress.value)
}
</script>

<template>
  <section class="panel relative overflow-hidden rounded-lg p-4 sm:min-h-[32rem] sm:p-6">
    <div class="absolute left-0 top-0 h-1 w-full" :class="statusTone === 'connected' ? 'bg-teal-500' : statusTone === 'failed' ? 'bg-red-500' : 'bg-graphite-200'" />

    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <div class="text-xs font-medium uppercase text-graphite-500">{{ t('client.selectedProfile') }}</div>
        <h2 class="mt-2 break-words text-xl font-semibold tracking-normal sm:text-2xl">{{ selectedNode?.name ?? t('client.noProfile') }}</h2>
        <div class="mt-1 break-all font-mono text-xs text-graphite-500 sm:text-sm">{{ endpoint }}</div>
      </div>
      <span
        class="badge"
        :class="statusTone === 'connected' ? 'badge-ok' : statusTone === 'failed' ? 'badge-warn' : 'badge-muted'"
      >
        {{ statusText }}
      </span>
    </div>

    <div class="grid min-h-[13rem] place-items-center py-6 sm:min-h-[20rem] sm:py-8">
      <button
        class="focus-ring grid h-32 w-32 place-items-center rounded-full border text-white shadow-[0_18px_45px_rgba(16,24,40,0.18)] transition disabled:cursor-not-allowed disabled:opacity-50 sm:h-40 sm:w-40"
        :class="isConnected ? 'border-red-500 bg-red-600 hover:bg-red-700' : 'border-teal-500 bg-teal-600 hover:bg-teal-700'"
        type="button"
        :disabled="!isConnected && !canConnect"
        @click="toggleConnection"
      >
        <span class="grid place-items-center gap-3">
          <PowerIcon v-if="isConnected" class="h-8 w-8 sm:h-10 sm:w-10" />
          <SignalIcon v-else class="h-8 w-8 sm:h-10 sm:w-10" />
          <span class="text-sm font-semibold">{{ isConnected ? t('actions.disconnect') : isBusy ? t('actions.connecting') : t('actions.connect') }}</span>
        </span>
      </button>
    </div>

    <div class="grid gap-3 md:grid-cols-3">
      <label class="grid gap-1.5">
        <span class="text-xs font-medium text-graphite-500">{{ t('client.network') }}</span>
        <select v-model="networkMode" class="focus-ring h-10 rounded-md border border-black/10 bg-white px-3 text-sm">
          <option value="local_proxy">{{ t('modes.localProxy') }}</option>
          <option value="system_proxy">{{ t('modes.systemProxy') }}</option>
          <option value="tun">{{ t('modes.tun') }}</option>
        </select>
      </label>

      <label class="grid gap-1.5">
        <span class="text-xs font-medium text-graphite-500">{{ t('client.core') }}</span>
        <select v-model="selectedCore" class="focus-ring h-10 rounded-md border border-black/10 bg-white px-3 text-sm">
          <option value="sing-box">sing-box</option>
          <option value="xray">xray-core</option>
        </select>
      </label>

      <label class="grid gap-1.5">
        <span class="text-xs font-medium text-graphite-500">{{ t('client.profile') }}</span>
        <select v-model="selectedNodeIdModel" class="focus-ring h-10 rounded-md border border-black/10 bg-white px-3 text-sm">
          <option value="" disabled>{{ t('client.selectProfile') }}</option>
          <option v-for="node in nodes" :key="node.id" :value="node.id">{{ node.name }}</option>
        </select>
      </label>
    </div>

    <dl class="mt-5 grid overflow-hidden rounded-md border border-black/10 text-sm md:grid-cols-3 md:divide-x md:divide-black/10">
      <div class="border-b border-black/10 px-4 py-3 last:border-b-0 md:border-b-0">
        <dt class="text-xs font-medium text-graphite-500">{{ t('client.localProxy') }}</dt>
        <dd class="mt-1 flex items-center gap-2 font-mono text-xs text-graphite-900">
          {{ localProxyAddress }}
          <button class="focus-ring rounded p-1 text-graphite-500 hover:bg-graphite-100" type="button" :title="t('client.copyProxyAddress')" @click="copyAddress">
            <ClipboardDocumentIcon class="h-4 w-4" />
          </button>
        </dd>
      </div>
      <div class="border-b border-black/10 px-4 py-3 last:border-b-0 md:border-b-0">
        <dt class="text-xs font-medium text-graphite-500">{{ t('client.windowsProxy') }}</dt>
        <dd class="mt-1 truncate font-medium">{{ systemProxyLabel }}</dd>
        <dd v-if="systemProxyDetail" class="mt-1 truncate font-mono text-xs text-graphite-500">{{ systemProxyDetail }}</dd>
      </div>
      <div class="border-b border-black/10 px-4 py-3 last:border-b-0 md:border-b-0">
        <dt class="text-xs font-medium text-graphite-500">{{ t('client.mode') }}</dt>
        <dd class="mt-1 font-medium">{{ networkModeLabel }}</dd>
      </div>
    </dl>
    <div class="mt-4 rounded-md border border-black/10 bg-graphite-50 px-3 py-2 text-xs text-graphite-600">
      {{ networkModeHint }}
    </div>

    <div v-if="externalProxyActive" class="mt-4 rounded-md border border-sky-200 bg-sky-50 px-3 py-2 text-xs text-sky-800">
      {{ t('client.externalProxyWarning') }}
    </div>
    <div v-if="networkMode === 'system_proxy'" class="mt-4 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-800">
      {{ t('client.systemProxyWarning') }}
    </div>
    <div v-if="networkMode === 'tun'" class="mt-4 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-800">
      {{ t('client.tunWarning') }}
    </div>
    <div v-if="status?.lastError" class="mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700">
      {{ status.lastError }}
    </div>
  </section>
</template>
