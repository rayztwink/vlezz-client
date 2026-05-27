<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import SettingsForm from '@/components/SettingsForm.vue'
import { useDiagnosticsStore } from '@/stores/diagnostics'
import { useSettingsStore } from '@/stores/settings'

const settings = useSettingsStore()
const diagnostics = useDiagnosticsStore()
const { t } = useI18n()

onMounted(() => {
  void Promise.all([settings.load(), diagnostics.loadRuntime()])
})
</script>

<template>
  <div class="mx-auto grid w-full max-w-7xl gap-4 sm:gap-5">
    <div>
      <h1 class="text-xl font-semibold tracking-normal sm:text-2xl">{{ t('settings.title') }}</h1>
      <p class="mt-1 text-sm text-graphite-500">{{ t('settings.subtitle') }}</p>
    </div>
    <div v-if="settings.error" class="flex gap-2 rounded-lg border border-amber-600/20 bg-amber-50 p-4 text-sm text-amber-800">
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="h-5 w-5 shrink-0 text-amber-600">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
      </svg>
      <div>
        <span class="font-medium">Error:</span>
        <span class="ml-1">{{ settings.error }}</span>
      </div>
    </div>
    <SettingsForm :settings="settings.settings" :runtime="diagnostics.runtime" @patch="settings.patch" />
  </div>
</template>
