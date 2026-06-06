<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { CheckCircleIcon, ExclamationTriangleIcon } from '@heroicons/vue/24/outline'
import { rayflowApi } from '@/services/api'
import type { AppSettings, NetworkMode, RuntimeCapabilities } from '@/types/api'
import { useUpdaterStore } from '@/stores/updater'

const props = defineProps<{ settings: AppSettings | null; runtime?: RuntimeCapabilities | null }>()
const emit = defineEmits<{ patch: [payload: Partial<AppSettings>] }>()
const { t } = useI18n()

const updaterStore = useUpdaterStore()

onMounted(() => {
  void updaterStore.loadCurrentVersion()
})

const statusText = computed(() => {
  if (updaterStore.status === 'error') {
    return updaterStore.error || 'Error checking or downloading update'
  }
  if (updaterStore.checking) {
    const val = t('settings.updater.checking')
    if (val === 'settings.updater.checking') {
      return 'checking'
    }
    return val
  }
  if (updaterStore.downloadedAndInstalled) {
    const val = t('settings.updater.installing')
    if (val === 'settings.updater.installing') {
      return 'installing'
    }
    return val
  }
  if (updaterStore.downloading) {
    const val = t('settings.updater.downloading')
    if (val === 'settings.updater.downloading') {
      return 'downloading'
    }
    return val
  }
  if (updaterStore.updateAvailable && updaterStore.newVersion) {
    return updaterStore.newVersion
  }
  if (updaterStore.status === 'up-to-date') {
    const val = t('settings.updater.upToDate')
    if (val === 'settings.updater.upToDate') {
      return 'up to date'
    }
    return val
  }
  return ''
})

const isSupportedPlatform = computed(() => {
  const p = props.runtime?.platform
  if (!p) return true
  return ['win32', 'darwin', 'linux', 'windows', 'macos'].includes(p.toLowerCase())
})

const isFormDisabled = computed(() => {
  return (updaterStore.downloading || (updaterStore.downloadedAndInstalled && !updaterStore.relaunchCancelled)) && !updaterStore.error
})

function formatBytes(bytes: number) {
  if (isNaN(bytes) || !isFinite(bytes)) return '0 bytes'
  const sign = bytes < 0 ? '-' : ''
  const absBytes = Math.abs(bytes)
  if (absBytes === 0) return '0 bytes'
  
  const k = 1024
  const sizes = ['bytes', 'KB', 'MB', 'GB', 'TB', 'PB']
  let i = 0
  if (absBytes >= 1) {
    i = Math.floor(Math.log(absBytes) / Math.log(k))
  }
  if (i >= sizes.length) {
    i = sizes.length - 1
  }
  const val = parseFloat((absBytes / Math.pow(k, i)).toFixed(2))
  return `${sign}${val} ${sizes[i]}`
}

const formattedProgress = computed(() => {
  const downloaded = updaterStore.downloadedBytes
  const total = updaterStore.totalBytes
  
  if (total && total > 0) {
    const pct = updaterStore.progressPercent
    return `${pct}% (${formatBytes(downloaded)} / ${formatBytes(total)})`
  } else {
    return `${formatBytes(downloaded)}`
  }
})

const draft = ref({
  defaultCore: 'sing-box',
  localProxyPort: 2080,
  singBoxPath: '',
  xrayPath: '',
    zapretPath: '',
    preferredNetworkMode: 'local_proxy',
  tunEnabled: false,
  tunStack: 'system',
  tunAutoRoute: true,
  tunStrictRoute: true
})
const validation = ref<Record<string, { ok: boolean; message: string }>>({})

watch(
  () => props.settings,
  (settings) => {
    if (!settings) {
      return
    }
    draft.value = {
      defaultCore: settings.defaultCore || 'sing-box',
      localProxyPort: settings.localProxyPort || 2080,
      singBoxPath: settings.singBoxPath || '',
      xrayPath: settings.xrayPath || '',
      zapretPath: settings.zapretPath || '',
      preferredNetworkMode: settings.preferredNetworkMode || 'local_proxy',
      tunEnabled: settings.tunEnabled || false,
      tunStack: settings.tunStack || 'system',
      tunAutoRoute: settings.tunAutoRoute ?? true,
      tunStrictRoute: settings.tunStrictRoute ?? true
    }
  },
  { immediate: true }
)

const theme = computed({
  get: () => props.settings?.theme ?? 'system',
  set: (value) => emit('patch', { theme: value })
})

const language = computed({
  get: () => props.settings?.language ?? 'system',
  set: (value) => emit('patch', { language: value })
})

const autostart = computed({
  get: () => props.settings?.autostart ?? false,
  set: (value) => emit('patch', { autostart: value })
})

function saveCoreSettings() {
  const port = Number(draft.value.localProxyPort)
  if (isNaN(port) || !Number.isInteger(port) || port < 1 || port > 65535) {
    validation.value['localProxyPort'] = {
      ok: false,
      message: 'Port must be a valid integer between 1 and 65535'
    }
    return
  }
  delete validation.value['localProxyPort']

  emit('patch', {
    defaultCore: draft.value.defaultCore,
    localProxyPort: port,
    singBoxPath: draft.value.singBoxPath.trim(),
    xrayPath: draft.value.xrayPath.trim(),
    zapretPath: draft.value.zapretPath.trim(),
    preferredNetworkMode: draft.value.preferredNetworkMode as NetworkMode,
    tunEnabled: draft.value.tunEnabled,
    tunStack: draft.value.tunStack,
    tunAutoRoute: draft.value.tunAutoRoute,
    tunStrictRoute: draft.value.tunStrictRoute
  })
}

watch(
  () => draft.value.singBoxPath,
  () => {
    delete validation.value['sing-box']
  }
)

watch(
  () => draft.value.xrayPath,
  () => {
    delete validation.value['xray']
  }
)

watch(
  () => draft.value.zapretPath,
  () => {
    delete validation.value['zapret']
  }
)

watch(
  () => draft.value.localProxyPort,
  () => {
    delete validation.value['localProxyPort']
  }
)

async function validate(core: string, path: string) {
  validation.value[core] = { ok: false, message: t('settings.checking') }
  try {
    const result = await rayflowApi.validateCore({ core, path: path.trim() })
    validation.value[core] = {
      ok: result.ok,
      message: result.ok ? result.version || t('settings.validationPassed') : result.error || t('common.failed')
    }
  } catch (err: any) {
    console.error('Validation failed:', err)
    let errMsg = t('common.failed')
    if (err?.response?.status === 401) {
      errMsg = 'Unauthorized (401): Please restart the rayflowd backend process to sync authentication tokens.'
    } else if (err?.response?.data?.error) {
      errMsg = err.response.data.error
    } else if (err?.message) {
      errMsg = err.message
    }
    validation.value[core] = {
      ok: false,
      message: errMsg
    }
  }
}

async function pasteFromClipboard(field: 'singBoxPath' | 'xrayPath' | 'zapretPath') {
  try {
    const text = await navigator.clipboard.readText()
    if (text) {
      draft.value[field] = text.trim()
    }
  } catch (err) {
    console.error('Failed to read clipboard', err)
  }
}

defineExpose({
  formatBytes,
  draft,
  validation,
  saveCoreSettings
})
</script>

<template>
  <div class="grid w-full gap-5 sm:gap-6">
  <div class="panel rounded-lg p-5">
    <h2 class="text-lg font-semibold">{{ t('settings.general') }}</h2>
    <div class="mt-5 grid gap-4">
      <label class="grid gap-1">
        <span class="text-sm font-medium">{{ t('settings.theme') }}</span>
        <select v-model="theme" :disabled="isFormDisabled" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm">
          <option value="system">{{ t('settings.themeOptions.system') }}</option>
          <option value="light">{{ t('settings.themeOptions.light') }}</option>
          <option value="dark">{{ t('settings.themeOptions.dark') }}</option>
        </select>
      </label>
      <label class="grid gap-1">
        <span class="text-sm font-medium">{{ t('settings.language') }}</span>
        <select v-model="language" :disabled="isFormDisabled" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm">
          <option value="system">{{ t('settings.languageOptions.system') }}</option>
          <option value="ru">{{ t('settings.languageOptions.ru') }}</option>
          <option value="en">{{ t('settings.languageOptions.en') }}</option>
        </select>
      </label>
      <label class="grid gap-3 rounded-lg border border-black/10 px-3 py-2 sm:flex sm:items-center sm:justify-between">
        <span class="min-w-0">
          <span class="block text-sm font-medium">{{ t('settings.autostart') }}</span>
          <span class="text-xs text-graphite-500">{{ t('settings.autostartDescription') }}</span>
        </span>
        <input v-model="autostart" :disabled="isFormDisabled" type="checkbox" class="h-5 w-5 accent-teal-600" />
      </label>
    </div>
  </div>

  <div class="panel rounded-lg p-5">
    <div class="mb-5">
      <h2 class="text-lg font-semibold">{{ t('settings.coreRuntime') }}</h2>
      <p class="mt-1 text-sm text-graphite-500">{{ t('settings.coreDescription') }}</p>
    </div>

    <div class="grid gap-4">
      <div class="grid gap-4 md:grid-cols-2">
        <label class="grid gap-1">
          <span class="text-sm font-medium">{{ t('settings.defaultCore') }}</span>
          <select v-model="draft.defaultCore" :disabled="isFormDisabled" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm">
            <option value="sing-box">sing-box</option>
            <option value="xray">xray-core</option>
          </select>
        </label>
        <label class="grid gap-1">
          <span class="text-sm font-medium">{{ t('settings.localProxyPort') }}</span>
          <input v-model.number="draft.localProxyPort" :disabled="isFormDisabled" type="number" min="1" max="65535" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm" />
        </label>
      </div>

      <label class="grid gap-1">
        <span class="text-sm font-medium">{{ t('settings.defaultNetworkMode') }}</span>
        <select v-model="draft.preferredNetworkMode" :disabled="isFormDisabled" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm">
          <option value="local_proxy">{{ t('modes.localProxy') }}</option>
          <option value="system_proxy">{{ t('modes.systemProxy') }}</option>
          <option value="tun">{{ t('modes.tun') }}</option>
        </select>
      </label>

      <div class="grid gap-3">
        <label class="grid gap-1">
          <span class="text-sm font-medium">{{ t('settings.singBoxPath') }}</span>
          <div class="grid gap-2 sm:flex">
            <input v-model="draft.singBoxPath" :disabled="isFormDisabled" class="focus-ring min-w-0 flex-1 rounded-md border border-black/10 bg-white px-3 py-2 text-sm" placeholder="C:\\Tools\\sing-box\\sing-box.exe" />
            <button :disabled="isFormDisabled" class="focus-ring rounded-md border border-black/10 px-3 py-2 text-sm hover:bg-graphite-50" type="button" @click="pasteFromClipboard('singBoxPath')">{{ t('actions.paste') }}</button>
            <button :disabled="isFormDisabled" class="focus-ring rounded-md border border-black/10 px-3 py-2 text-sm hover:bg-graphite-50" type="button" @click="validate('sing-box', draft.singBoxPath)">{{ t('actions.check') }}</button>
          </div>
        </label>
        <label class="grid gap-1">
          <span class="text-sm font-medium">{{ t('settings.xrayPath') }}</span>
          <div class="grid gap-2 sm:flex">
            <input v-model="draft.xrayPath" :disabled="isFormDisabled" class="focus-ring min-w-0 flex-1 rounded-md border border-black/10 bg-white px-3 py-2 text-sm" placeholder="C:\\Tools\\xray\\xray.exe" />
            <button :disabled="isFormDisabled" class="focus-ring rounded-md border border-black/10 px-3 py-2 text-sm hover:bg-graphite-50" type="button" @click="pasteFromClipboard('xrayPath')">{{ t('actions.paste') }}</button>
            <button :disabled="isFormDisabled" class="focus-ring rounded-md border border-black/10 px-3 py-2 text-sm hover:bg-graphite-50" type="button" @click="validate('xray', draft.xrayPath)">{{ t('actions.check') }}</button>
          </div>
        </label>
        <label class="grid gap-1">
          <span class="text-sm font-medium">{{ t('settings.zapretPath') }}</span>
          <div class="grid gap-2 sm:flex">
            <input v-model="draft.zapretPath" :disabled="isFormDisabled" class="focus-ring min-w-0 flex-1 rounded-md border border-black/10 bg-white px-3 py-2 text-sm" placeholder="C:\\Tools\\zapret\\winws.exe" />
            <button :disabled="isFormDisabled" class="focus-ring rounded-md border border-black/10 px-3 py-2 text-sm hover:bg-graphite-50" type="button" @click="pasteFromClipboard('zapretPath')">{{ t('actions.paste') }}</button>
            <button :disabled="isFormDisabled" class="focus-ring rounded-md border border-black/10 px-3 py-2 text-sm hover:bg-graphite-50" type="button" @click="validate('zapret', draft.zapretPath)">{{ t('actions.check') }}</button>
          </div>
        </label>
      </div>

      <div class="rounded-lg border border-black/10 p-4">
        <div class="mb-4">
          <h3 class="text-sm font-semibold">{{ t('settings.tunMode') }}</h3>
          <p class="mt-1 text-xs text-graphite-500">{{ t('settings.tunDescription') }}</p>
          <p v-if="runtime" class="mt-2 text-xs text-graphite-500">
            {{ t('settings.tunStatus', { platform: runtime.platform, admin: runtime.isAdmin ? t('common.yes') : t('common.no'), systemProxy: runtime.systemProxySupported ? t('common.yes') : t('common.no') }) }}
          </p>
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <label class="flex items-center justify-between rounded-md bg-graphite-50 px-3 py-2">
            <span class="text-sm font-medium">{{ t('settings.enableTun') }}</span>
            <input v-model="draft.tunEnabled" :disabled="isFormDisabled" type="checkbox" class="h-5 w-5 accent-teal-600" />
          </label>
          <label class="grid gap-1">
            <span class="text-xs font-medium text-graphite-500">{{ t('settings.stack') }}</span>
            <select v-model="draft.tunStack" :disabled="isFormDisabled" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm">
              <option value="system">system</option>
              <option value="gvisor">gvisor</option>
              <option value="mixed">mixed</option>
            </select>
          </label>
          <label class="flex items-center justify-between rounded-md bg-graphite-50 px-3 py-2">
            <span class="text-sm font-medium">{{ t('settings.autoRoute') }}</span>
            <input v-model="draft.tunAutoRoute" :disabled="isFormDisabled" type="checkbox" class="h-5 w-5 accent-teal-600" />
          </label>
          <label class="flex items-center justify-between rounded-md bg-graphite-50 px-3 py-2">
            <span class="text-sm font-medium">{{ t('settings.strictRoute') }}</span>
            <input v-model="draft.tunStrictRoute" :disabled="isFormDisabled" type="checkbox" class="h-5 w-5 accent-teal-600" />
          </label>
        </div>
      </div>

      <div v-if="Object.keys(validation).length" class="grid gap-2 rounded-lg border border-black/10 bg-graphite-50 p-3 text-sm">
        <div v-for="(item, key) in validation" :key="key" class="flex gap-2">
          <CheckCircleIcon v-if="item.ok" class="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" />
          <ExclamationTriangleIcon v-else class="mt-0.5 h-4 w-4 shrink-0 text-amber-600" />
          <span class="font-medium">{{ key }}:</span>
          <span class="break-words text-graphite-600">{{ item.message }}</span>
        </div>
      </div>

      <div class="flex justify-end">
        <button :disabled="isFormDisabled" class="focus-ring w-full rounded-md bg-teal-600 px-4 py-2 text-sm font-medium text-white hover:bg-teal-700 sm:w-auto" type="button" @click="saveCoreSettings">{{ t('settings.saveCore') }}</button>
      </div>
    </div>
  </div>

  <!-- Software Updates Panel -->
  <div class="panel rounded-lg p-5">
    <h2 class="text-lg font-semibold">{{ t('settings.updater.title') }}</h2>
    <p class="mt-1 text-sm text-graphite-500">{{ t('settings.updater.description') }}</p>
    
    <div class="mt-5 grid gap-4">
      <!-- Current Version Display -->
      <div class="text-sm font-medium text-graphite-700">
        {{ t('settings.updater.currentVersion') }}: 
        <span data-testid="current-version" class="font-semibold text-graphite-900">{{ updaterStore.currentVersion }}</span>
      </div>

      <!-- Update Status Text -->
      <div class="text-sm text-graphite-600">
        <span data-testid="update-status">{{ statusText }}</span>
      </div>

      <!-- Error Message Block -->
      <div v-if="updaterStore.error" data-testid="update-error" class="rounded-md bg-rose-50 p-3 text-sm text-rose-800">
        {{ updaterStore.error }}
      </div>

      <!-- Release Notes Display -->
      <div v-if="updaterStore.updateAvailable && updaterStore.updateBody" class="rounded-lg border border-black/10 bg-graphite-50 p-4">
        <h3 class="text-sm font-semibold text-graphite-800 mb-2">{{ t('settings.updater.releaseNotes') }}</h3>
        <p data-testid="release-notes" class="text-sm text-graphite-600 whitespace-pre-wrap">{{ updaterStore.updateBody }}</p>
      </div>

      <div class="flex flex-wrap gap-2 mt-2">
        <!-- Check for Updates Button -->
        <button 
          v-if="!updaterStore.updateAvailable && !updaterStore.downloadedAndInstalled"
          :disabled="!isSupportedPlatform || updaterStore.checking || updaterStore.cooldownActive || isFormDisabled"
          data-testid="check-update-btn"
          type="button" 
          class="focus-ring rounded-md bg-teal-600 px-4 py-2 text-sm font-medium text-white hover:bg-teal-700 disabled:opacity-50" 
          @click="updaterStore.checkForUpdates"
        >
          {{ t('settings.updater.checkButton') }}
        </button>
        
        <!-- Download & Install Button -->
        <button 
          v-if="updaterStore.updateAvailable && !updaterStore.downloading && !updaterStore.downloadedAndInstalled"
          :disabled="isFormDisabled"
          data-testid="download-install-btn"
          type="button" 
          class="focus-ring rounded-md bg-teal-600 px-4 py-2 text-sm font-medium text-white hover:bg-teal-700" 
          @click="updaterStore.downloadAndInstall"
        >
          {{ t('settings.updater.downloadAndInstall') }}
        </button>

        <!-- Relaunch Now Button -->
        <button 
          v-if="updaterStore.downloadedAndInstalled"
          data-testid="relaunch-btn"
          type="button" 
          class="focus-ring rounded-md bg-teal-600 px-4 py-2 text-sm font-medium text-white hover:bg-teal-700" 
          @click="updaterStore.relaunchApp"
        >
          {{ t('settings.updater.relaunch') }}
        </button>
      </div>

      <!-- Download Progress Bar -->
      <div v-if="updaterStore.downloading || updaterStore.downloadedAndInstalled" data-testid="progress-bar-container" class="grid gap-1.5 rounded-lg border border-black/10 bg-graphite-50 p-4">
        <div class="flex justify-between text-xs font-semibold text-graphite-700">
          <span>{{ t('settings.updater.downloading') }}</span>
          <span data-testid="progress-bar">{{ formattedProgress }}</span>
        </div>
        <div class="w-full bg-graphite-200 rounded-full h-2">
          <div class="bg-teal-600 h-2 rounded-full transition-all duration-150" :style="{ width: updaterStore.progressPercent + '%' }"></div>
        </div>
      </div>
    </div>
  </div>
  </div>
</template>
