<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { CheckCircleIcon, ExclamationTriangleIcon } from '@heroicons/vue/24/outline'
import { rayflowApi } from '@/services/api'
import type { AppSettings, NetworkMode, RuntimeCapabilities } from '@/types/api'

const props = defineProps<{ settings: AppSettings | null; runtime?: RuntimeCapabilities | null }>()
const emit = defineEmits<{ patch: [payload: Partial<AppSettings>] }>()
const { t } = useI18n()

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
  emit('patch', {
    defaultCore: draft.value.defaultCore,
    localProxyPort: Number(draft.value.localProxyPort),
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

async function validate(core: string, path: string) {
  validation.value[core] = { ok: false, message: t('settings.checking') }
  const result = await rayflowApi.validateCore({ core, path: path.trim() })
  validation.value[core] = {
    ok: result.ok,
    message: result.ok ? result.version || t('settings.validationPassed') : result.error || t('common.failed')
  }
}
</script>

<template>
  <div class="grid w-full gap-5 sm:gap-6">
  <div class="panel rounded-lg p-5">
    <h2 class="text-lg font-semibold">{{ t('settings.general') }}</h2>
    <div class="mt-5 grid gap-4">
      <label class="grid gap-1">
        <span class="text-sm font-medium">{{ t('settings.theme') }}</span>
        <select v-model="theme" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm">
          <option value="system">{{ t('settings.themeOptions.system') }}</option>
          <option value="light">{{ t('settings.themeOptions.light') }}</option>
          <option value="dark">{{ t('settings.themeOptions.dark') }}</option>
        </select>
      </label>
      <label class="grid gap-1">
        <span class="text-sm font-medium">{{ t('settings.language') }}</span>
        <select v-model="language" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm">
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
        <input v-model="autostart" type="checkbox" class="h-5 w-5 accent-teal-600" />
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
          <select v-model="draft.defaultCore" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm">
            <option value="sing-box">sing-box</option>
            <option value="xray">xray-core</option>
          </select>
        </label>
        <label class="grid gap-1">
          <span class="text-sm font-medium">{{ t('settings.localProxyPort') }}</span>
          <input v-model.number="draft.localProxyPort" type="number" min="1" max="65535" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm" />
        </label>
      </div>

      <label class="grid gap-1">
        <span class="text-sm font-medium">{{ t('settings.defaultNetworkMode') }}</span>
        <select v-model="draft.preferredNetworkMode" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm">
          <option value="local_proxy">{{ t('modes.localProxy') }}</option>
          <option value="system_proxy">{{ t('modes.systemProxy') }}</option>
          <option value="tun">{{ t('modes.tun') }}</option>
        </select>
      </label>

      <div class="grid gap-3">
        <label class="grid gap-1">
          <span class="text-sm font-medium">{{ t('settings.singBoxPath') }}</span>
          <div class="grid gap-2 sm:flex">
            <input v-model="draft.singBoxPath" class="focus-ring min-w-0 flex-1 rounded-md border border-black/10 bg-white px-3 py-2 text-sm" placeholder="C:\\Tools\\sing-box\\sing-box.exe" />
            <button class="focus-ring rounded-md border border-black/10 px-3 py-2 text-sm hover:bg-graphite-50" type="button" @click="validate('sing-box', draft.singBoxPath)">{{ t('actions.check') }}</button>
          </div>
        </label>
        <label class="grid gap-1">
          <span class="text-sm font-medium">{{ t('settings.xrayPath') }}</span>
          <div class="grid gap-2 sm:flex">
            <input v-model="draft.xrayPath" class="focus-ring min-w-0 flex-1 rounded-md border border-black/10 bg-white px-3 py-2 text-sm" placeholder="C:\\Tools\\xray\\xray.exe" />
            <button class="focus-ring rounded-md border border-black/10 px-3 py-2 text-sm hover:bg-graphite-50" type="button" @click="validate('xray', draft.xrayPath)">{{ t('actions.check') }}</button>
          </div>
        </label>
        <label class="grid gap-1">
          <span class="text-sm font-medium">{{ t('settings.zapretPath') }}</span>
          <div class="grid gap-2 sm:flex">
            <input v-model="draft.zapretPath" class="focus-ring min-w-0 flex-1 rounded-md border border-black/10 bg-white px-3 py-2 text-sm" placeholder="C:\\Tools\\zapret\\winws.exe" />
            <button class="focus-ring rounded-md border border-black/10 px-3 py-2 text-sm hover:bg-graphite-50" type="button" @click="validate('zapret', draft.zapretPath)">{{ t('actions.check') }}</button>
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
            <input v-model="draft.tunEnabled" type="checkbox" class="h-5 w-5 accent-teal-600" />
          </label>
          <label class="grid gap-1">
            <span class="text-xs font-medium text-graphite-500">{{ t('settings.stack') }}</span>
            <select v-model="draft.tunStack" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm">
              <option value="system">system</option>
              <option value="gvisor">gvisor</option>
              <option value="mixed">mixed</option>
            </select>
          </label>
          <label class="flex items-center justify-between rounded-md bg-graphite-50 px-3 py-2">
            <span class="text-sm font-medium">{{ t('settings.autoRoute') }}</span>
            <input v-model="draft.tunAutoRoute" type="checkbox" class="h-5 w-5 accent-teal-600" />
          </label>
          <label class="flex items-center justify-between rounded-md bg-graphite-50 px-3 py-2">
            <span class="text-sm font-medium">{{ t('settings.strictRoute') }}</span>
            <input v-model="draft.tunStrictRoute" type="checkbox" class="h-5 w-5 accent-teal-600" />
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
        <button class="focus-ring w-full rounded-md bg-teal-600 px-4 py-2 text-sm font-medium text-white hover:bg-teal-700 sm:w-auto" type="button" @click="saveCoreSettings">{{ t('settings.saveCore') }}</button>
      </div>
    </div>
  </div>
  </div>
</template>
