<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import DiagnosticsPanel from '@/components/DiagnosticsPanel.vue'
import LatencyChart from '@/components/LatencyChart.vue'
import { useDiagnosticsStore } from '@/stores/diagnostics'

const diagnostics = useDiagnosticsStore()
const { t } = useI18n()

async function runCheck(target: string, type: string) {
  await diagnostics.check(target, type)
}

onMounted(() => {
  void diagnostics.load()
})
</script>

<template>
  <div class="mx-auto grid max-w-7xl gap-4 sm:gap-5">
    <div>
      <h1 class="text-xl font-semibold tracking-normal sm:text-2xl">{{ t('diagnostics.title') }}</h1>
      <p class="mt-1 text-sm text-graphite-500">{{ t('diagnostics.subtitle') }}</p>
    </div>

    <section class="grid gap-5 xl:grid-cols-[1fr_0.9fr]">
      <DiagnosticsPanel :history="diagnostics.history" :loading="diagnostics.loading" @check="runCheck" />
      <LatencyChart :data="diagnostics.history" />
    </section>
  </div>
</template>
