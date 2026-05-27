<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowPathIcon } from '@heroicons/vue/24/outline'
import type { ConnectionStatus, IPCheckResult, RuntimeCapabilities } from '@/types/api'

const props = defineProps<{
  status: ConnectionStatus | null
  direct: IPCheckResult | null
  rayflow: IPCheckResult | null
  tun: IPCheckResult | null
  runtime: RuntimeCapabilities | null
  tunEnabled?: boolean
  defaultCore?: string
  loadingRoute?: string | null
}>()

const emit = defineEmits<{
  check: [route: 'direct' | 'rayflow_proxy' | 'tun']
}>()

const { t } = useI18n()

const systemProxyLabel = computed(() => {
  const proxy = props.status?.systemProxy
  if (!proxy?.proxyEnable) {
    return t('common.off')
  }
  return proxy.enabledByRayflow ? t('ipCheck.systemProxyOwned') : t('ipCheck.systemProxyExternal')
})

const systemProxyDetail = computed(() => {
  const proxy = props.status?.systemProxy
  if (!proxy?.proxyEnable) {
    return ''
  }
  return proxy.proxyServer || proxy.currentProxyServer
})

const tunAvailable = computed(() => {
  const core = props.status?.selectedCore || props.defaultCore || ''
  return Boolean(props.runtime?.isAdmin && props.tunEnabled && core !== 'xray' && core !== 'xray-core')
})

function checkedAt(result: IPCheckResult | null) {
  return result?.checkedAt ? new Date(result.checkedAt).toLocaleString() : t('ipCheck.notChecked')
}
</script>

<template>
  <section class="panel rounded-lg p-4 sm:p-5">
    <div class="mb-4 flex flex-wrap items-start justify-between gap-3">
      <div>
        <h2 class="text-base font-semibold">{{ t('ipCheck.title') }}</h2>
        <p class="text-xs text-graphite-500">{{ t('ipCheck.subtitle') }}</p>
      </div>
      <button class="focus-ring inline-flex items-center gap-2 rounded-md border border-black/10 px-3 py-2 text-sm hover:bg-graphite-50" type="button" @click="emit('check', 'direct'); emit('check', 'rayflow_proxy')">
        <ArrowPathIcon class="h-4 w-4" />
        {{ t('ipCheck.runAll') }}
      </button>
    </div>

    <div class="grid gap-3 md:grid-cols-3">
      <div class="rounded-md border border-black/10 p-3">
        <div class="flex items-center justify-between gap-2">
          <h3 class="text-sm font-semibold">{{ t('ipCheck.direct') }}</h3>
          <button class="focus-ring rounded-md border border-black/10 px-2 py-1 text-xs hover:bg-graphite-50" type="button" :disabled="loadingRoute === 'direct'" @click="emit('check', 'direct')">
            {{ loadingRoute === 'direct' ? t('actions.checking') : t('actions.check') }}
          </button>
        </div>
        <div class="mt-3 font-mono text-sm">{{ direct?.ip ?? t('common.unknown') }}</div>
        <div class="mt-1 text-xs text-graphite-500">{{ direct?.country || direct?.provider || checkedAt(direct) }}</div>
        <div v-if="direct?.error" class="mt-2 break-words text-xs text-red-700">{{ direct.error }}</div>
      </div>

      <div class="rounded-md border border-black/10 p-3">
        <div class="flex items-center justify-between gap-2">
          <h3 class="text-sm font-semibold">{{ t('ipCheck.rayflow') }}</h3>
          <button class="focus-ring rounded-md border border-black/10 px-2 py-1 text-xs hover:bg-graphite-50" type="button" :disabled="loadingRoute === 'rayflow_proxy'" @click="emit('check', 'rayflow_proxy')">
            {{ loadingRoute === 'rayflow_proxy' ? t('actions.checking') : t('actions.check') }}
          </button>
        </div>
        <div class="mt-3 font-mono text-sm">{{ rayflow?.ip ?? t('common.unknown') }}</div>
        <div class="mt-1 text-xs text-graphite-500">{{ rayflow?.country || rayflow?.provider || checkedAt(rayflow) }}</div>
        <div v-if="rayflow?.error" class="mt-2 break-words text-xs text-red-700">{{ rayflow.error }}</div>
      </div>

      <div class="rounded-md border border-black/10 p-3">
        <div class="flex items-center justify-between gap-2">
          <h3 class="text-sm font-semibold">{{ t('ipCheck.tun') }}</h3>
          <button class="focus-ring rounded-md border border-black/10 px-2 py-1 text-xs hover:bg-graphite-50 disabled:cursor-not-allowed disabled:opacity-50" type="button" :disabled="!tunAvailable || loadingRoute === 'tun'" @click="emit('check', 'tun')">
            {{ loadingRoute === 'tun' ? t('actions.checking') : t('actions.check') }}
          </button>
        </div>
        <div class="mt-3 font-mono text-sm">{{ tun?.ip ?? t('common.unknown') }}</div>
        <div class="mt-1 text-xs text-graphite-500">{{ tunAvailable ? (tun?.country || tun?.provider || checkedAt(tun)) : t('ipCheck.tunUnavailable') }}</div>
        <div v-if="tun?.error" class="mt-2 break-words text-xs text-red-700">{{ tun.error }}</div>
      </div>
    </div>

    <div class="mt-3 rounded-md border border-black/10 bg-graphite-50 px-3 py-2 text-xs text-graphite-600">
      <span class="font-medium">{{ t('ipCheck.windowsProxy') }}:</span>
      {{ systemProxyLabel }}
      <span v-if="systemProxyDetail" class="font-mono"> · {{ systemProxyDetail }}</span>
    </div>
  </section>
</template>
