<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import type { DiagnosticCheck } from '@/types/api'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent])

const props = defineProps<{ data: DiagnosticCheck[] }>()
const { t } = useI18n()

const option = computed(() => {
  const points = props.data
    .filter((item) => typeof item.latencyMs === 'number')
    .slice(0, 20)
    .reverse()
  return {
    grid: { left: 28, right: 12, top: 18, bottom: 24 },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: points.map((item) => new Date(item.checkedAt).toLocaleTimeString()),
      axisLabel: { fontSize: 10 }
    },
    yAxis: { type: 'value', axisLabel: { fontSize: 10 } },
    series: [
      {
        type: 'line',
        smooth: true,
        symbolSize: 6,
        lineStyle: { width: 3, color: '#0f766e' },
        itemStyle: { color: '#0f766e' },
        areaStyle: { color: 'rgba(15, 118, 110, 0.10)' },
        data: points.map((item) => item.latencyMs ?? 0)
      }
    ]
  }
})
</script>

<template>
  <div class="panel rounded-lg p-4">
    <div class="mb-3 flex items-center justify-between">
      <div>
        <h2 class="text-sm font-semibold">{{ t('chart.latency') }}</h2>
        <p class="text-xs text-graphite-500 dark:text-graphite-300">{{ t('chart.recent') }}</p>
      </div>
      <span class="badge badge-muted">{{ t('chart.checks', { count: data.length }) }}</span>
    </div>
    <VChart class="h-64 w-full" :option="option" autoresize />
  </div>
</template>
