<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { PlusIcon } from '@heroicons/vue/24/outline'
import ConnectionPanel from '@/components/ConnectionPanel.vue'
import IPCheckPanel from '@/components/IPCheckPanel.vue'
import ImportNodeModal from '@/components/ImportNodeModal.vue'
import LogsViewer from '@/components/LogsViewer.vue'
import NodeTable from '@/components/NodeTable.vue'
import SubscriptionTable from '@/components/SubscriptionTable.vue'
import { useConnectionStore } from '@/stores/connection'
import { useDiagnosticsStore } from '@/stores/diagnostics'
import { useNodesStore } from '@/stores/nodes'
import { useSettingsStore } from '@/stores/settings'
import { useSubscriptionsStore } from '@/stores/subscriptions'
import type { NetworkMode } from '@/types/api'

const connection = useConnectionStore()
const diagnostics = useDiagnosticsStore()
const nodes = useNodesStore()
const settings = useSettingsStore()
const subscriptions = useSubscriptionsStore()
const importOpen = ref(false)
const selectedNodeId = ref('')
const subscriptionName = ref('')
const subscriptionUrl = ref('')
let pollTimer: number | undefined
const { t } = useI18n()

async function importNode(link: string, name?: string) {
  await nodes.importNode(link, name)
  importOpen.value = false
}

async function addSubscription() {
  if (!subscriptionName.value.trim() || !subscriptionUrl.value.trim()) {
    return
  }
  await subscriptions.create(subscriptionName.value.trim(), subscriptionUrl.value.trim())
  subscriptionName.value = ''
  subscriptionUrl.value = ''
}

async function updateSubscription(id: string) {
  await subscriptions.update(id)
  await nodes.load()
}

async function connect(nodeId: string, core: string, networkMode: NetworkMode) {
  selectedNodeId.value = nodeId
  await connection.connect(nodeId, core, networkMode)
}

async function checkIP(route: 'direct' | 'rayflow_proxy' | 'tun') {
  await diagnostics.checkIP(route, connection.status?.localProxyAddress || '127.0.0.1:2080', 'socks5')
}

function selectNode(nodeId: string) {
  selectedNodeId.value = nodeId
}

onMounted(() => {
  void Promise.all([nodes.load(), subscriptions.load(), connection.load(), settings.load(), diagnostics.loadRuntime()])
  pollTimer = window.setInterval(() => {
    void connection.load()
  }, 2500)
})

onUnmounted(() => {
  if (pollTimer) {
    window.clearInterval(pollTimer)
  }
})

watch(
  () => nodes.nodes,
  (profiles) => {
    if (!selectedNodeId.value && profiles.length > 0) {
      selectedNodeId.value = profiles[0].id
    }
  },
  { immediate: true }
)

watch(
  () => connection.status?.selectedNodeId,
  (nodeId) => {
    if (nodeId) {
      selectedNodeId.value = nodeId
    }
  },
  { immediate: true }
)
</script>

<template>
  <div class="mx-auto grid max-w-7xl gap-4 sm:gap-5">
    <div class="grid gap-3 sm:flex sm:items-center sm:justify-between">
      <div class="min-w-0">
        <h1 class="text-xl font-semibold tracking-normal sm:text-2xl">{{ t('client.title') }}</h1>
        <p class="mt-1 text-sm text-graphite-500">{{ t('client.subtitle') }}</p>
      </div>
      <button class="focus-ring inline-flex h-10 w-full items-center justify-center gap-2 rounded-md bg-teal-600 px-4 text-sm font-medium text-white hover:bg-teal-700 sm:w-auto" type="button" @click="importOpen = true">
        <PlusIcon class="h-4 w-4" />
        {{ t('actions.import') }}
      </button>
    </div>

    <section class="grid gap-5 xl:grid-cols-[22rem_1fr]">
      <NodeTable class="order-2 xl:order-1" :nodes="nodes.nodes" :selected-node-id="selectedNodeId" :loading="nodes.loading" @select="selectNode" @check="nodes.checkNode" @delete="nodes.deleteNode" />

      <div class="order-1 grid gap-5 xl:order-2">
        <ConnectionPanel
          :nodes="nodes.nodes"
          :selected-node-id="selectedNodeId"
          :status="connection.status"
          :loading="connection.loading"
          :default-core="settings.settings?.defaultCore"
          :default-network-mode="settings.settings?.preferredNetworkMode"
          @connect="connect"
          @disconnect="connection.disconnect"
          @select="selectNode"
        />
        <IPCheckPanel
          :status="connection.status"
          :direct="diagnostics.ipChecks.direct"
          :rayflow="diagnostics.ipChecks.rayflow_proxy"
          :tun="diagnostics.ipChecks.tun"
          :runtime="diagnostics.runtime"
          :tun-enabled="settings.settings?.tunEnabled"
          :default-core="settings.settings?.defaultCore"
          :loading-route="diagnostics.ipLoadingRoute"
          @check="checkIP"
        />
      </div>
    </section>

    <details class="panel rounded-lg">
      <summary class="focus-ring flex cursor-pointer list-none items-center justify-between gap-3 px-4 py-3 text-base font-semibold">
        {{ t('subscriptions.title') }}
        <span class="badge badge-muted">{{ subscriptions.subscriptions.length }}</span>
      </summary>
      <section class="grid gap-4 border-t border-black/10 p-4 xl:grid-cols-[0.9fr_1.1fr]">
        <form class="grid gap-3" @submit.prevent="addSubscription">
          <div>
            <h2 class="text-sm font-semibold">{{ t('subscriptions.addTitle') }}</h2>
            <p class="mt-1 text-xs text-graphite-500">{{ t('subscriptions.description') }}</p>
          </div>
          <input v-model="subscriptionName" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm" :placeholder="t('subscriptions.namePlaceholder')" />
          <input v-model="subscriptionUrl" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm" :placeholder="t('subscriptions.urlPlaceholder')" />
          <button class="focus-ring rounded-md bg-teal-600 px-3 py-2 text-sm font-medium text-white hover:bg-teal-700" type="submit">{{ t('subscriptions.add') }}</button>
          <div v-if="subscriptions.lastUpdate" class="rounded-md border border-teal-200 bg-teal-50 p-3 text-xs text-teal-800">
            {{ t('subscriptions.result', { imported: subscriptions.lastUpdate.imported, skipped: subscriptions.lastUpdate.skipped, failed: subscriptions.lastUpdate.failed }) }}
          </div>
        </form>
        <SubscriptionTable :subscriptions="subscriptions.subscriptions" embedded @update="updateSubscription" @delete="subscriptions.remove" />
      </section>
    </details>

    <details class="panel rounded-lg">
      <summary class="focus-ring flex cursor-pointer list-none items-center justify-between gap-3 px-4 py-3 text-base font-semibold">
        {{ t('logs.connection') }}
        <span class="badge badge-muted">{{ t('logs.lines', { count: connection.logs.length }) }}</span>
      </summary>
      <div class="border-t border-black/10">
        <LogsViewer :logs="connection.logs" :title="t('logs.connection')" embedded />
      </div>
    </details>
    <ImportNodeModal :open="importOpen" @close="importOpen = false" @submit="importNode" />
  </div>
</template>
