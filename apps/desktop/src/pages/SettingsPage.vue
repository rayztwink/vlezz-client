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
    <SettingsForm :settings="settings.settings" :runtime="diagnostics.runtime" @patch="settings.patch" />
  </div>
</template>
