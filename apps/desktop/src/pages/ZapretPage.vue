<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowPathIcon } from '@heroicons/vue/24/outline'
import LogsViewer from '@/components/LogsViewer.vue'
import ZapretPresetList from '@/components/ZapretPresetList.vue'
import { useZapretStore } from '@/stores/zapret'

const zapret = useZapretStore()
const { t } = useI18n()

onMounted(() => {
  void zapret.load()
})
</script>

<template>
  <div class="mx-auto grid max-w-7xl gap-4 sm:gap-5">
    <div class="grid gap-3 sm:flex sm:items-center sm:justify-between">
      <div class="min-w-0">
        <h1 class="text-xl font-semibold tracking-normal sm:text-2xl">{{ t('zapret.title') }}</h1>
        <p class="mt-1 text-sm text-graphite-500">{{ t('zapret.subtitle') }}</p>
      </div>
      <button class="focus-ring inline-flex w-full items-center justify-center gap-2 rounded-md border border-black/10 bg-white px-3 py-2 text-sm hover:bg-graphite-50 sm:w-auto" type="button" @click="zapret.updatePresets">
        <ArrowPathIcon class="h-4 w-4" />
        {{ t('zapret.updatePresets') }}
      </button>
    </div>

    <section class="grid gap-5 xl:grid-cols-[0.9fr_1.1fr]">
      <ZapretPresetList :presets="zapret.presets" @start="zapret.startPreset" @stop="zapret.stop" />
      <LogsViewer :logs="zapret.logs" :title="t('logs.zapret')" />
    </section>
  </div>
</template>
