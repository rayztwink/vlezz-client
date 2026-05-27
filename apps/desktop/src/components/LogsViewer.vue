<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { LogEntry } from '@/types/api'

defineProps<{ logs: LogEntry[]; title?: string; embedded?: boolean }>()
const { t } = useI18n()

async function copyLogs(logs: LogEntry[]) {
  const text = logs.map((entry) => `[${entry.createdAt}] ${entry.source} ${entry.level}: ${entry.message}`).join('\n')
  await navigator.clipboard.writeText(text)
}
</script>

<template>
  <div :class="embedded ? 'overflow-hidden' : 'panel overflow-hidden rounded-lg'">
    <div class="flex items-center justify-between border-b border-black/10 px-4 py-3" :class="embedded ? 'hidden' : ''">
      <h2 class="text-base font-semibold">{{ title ?? t('logs.title') }}</h2>
      <div class="flex items-center gap-2">
        <span class="badge badge-muted">{{ t('logs.lines', { count: logs.length }) }}</span>
        <button class="focus-ring rounded-md border border-black/10 px-2 py-1 text-xs hover:bg-graphite-50" type="button" @click="copyLogs(logs)">{{ t('actions.copy') }}</button>
      </div>
    </div>
    <div class="max-h-72 overflow-auto bg-white">
      <div v-if="logs.length === 0" class="px-4 py-8 text-center text-sm text-graphite-500">{{ t('logs.empty') }}</div>
      <div v-for="entry in logs" :key="entry.id" class="grid gap-1 border-b border-black/5 px-4 py-2 font-mono text-xs sm:grid-cols-[6rem_7rem_5rem_1fr] sm:gap-3">
        <span class="text-graphite-500">{{ new Date(entry.createdAt).toLocaleTimeString() }}</span>
        <div class="flex flex-wrap items-center gap-2 sm:contents">
        <span class="truncate text-graphite-500">{{ entry.source }}</span>
        <span :class="entry.level === 'error' ? 'text-red-600' : entry.level === 'warn' ? 'text-amber-600' : 'text-teal-700'">{{ entry.level }}</span>
        </div>
        <span class="break-words text-graphite-700">{{ entry.message }}</span>
      </div>
    </div>
  </div>
</template>
