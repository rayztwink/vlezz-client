<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { BeakerIcon } from '@heroicons/vue/24/outline'
import type { DiagnosticCheck } from '@/types/api'

defineProps<{ history: DiagnosticCheck[]; loading?: boolean }>()
const emit = defineEmits<{ check: [target: string, type: string] }>()
const { t } = useI18n()

const target = ref('example.com:443')
const type = ref('tcp')
</script>

<template>
  <div class="panel rounded-lg">
    <div class="border-b border-black/10 px-4 py-3">
      <h2 class="text-sm font-semibold">{{ t('diagnostics.panelTitle') }}</h2>
      <p class="text-xs text-graphite-500">{{ t('diagnostics.panelDescription') }}</p>
    </div>
    <form class="grid gap-2 border-b border-black/10 p-4 sm:grid-cols-[1fr_7rem_auto]" @submit.prevent="emit('check', target, type)">
      <input v-model="target" class="focus-ring min-w-0 rounded-md border border-black/10 bg-white px-3 py-2 text-sm" :placeholder="t('diagnostics.targetPlaceholder')" />
      <select v-model="type" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm">
        <option value="tcp">TCP</option>
        <option value="dns">DNS</option>
        <option value="http">HTTP</option>
      </select>
      <button class="focus-ring grid h-10 w-full place-items-center rounded-md bg-teal-600 text-white hover:bg-teal-700 sm:w-10" :title="t('actions.run')" type="submit">
        <BeakerIcon class="h-5 w-5" />
      </button>
    </form>
    <div class="max-h-[28rem] overflow-auto">
      <div v-if="history.length === 0" class="px-4 py-8 text-center text-sm text-graphite-500">{{ t('diagnostics.empty') }}</div>
      <div v-for="item in history" :key="item.id" class="grid gap-2 border-b border-black/5 px-4 py-3 text-sm sm:grid-cols-[1fr_7rem_7rem_9rem] sm:gap-3">
        <span class="truncate font-medium">{{ item.target }}</span>
        <div class="flex flex-wrap items-center gap-2 sm:contents">
        <span class="badge" :class="item.status === 'ok' ? 'badge-ok' : 'badge-warn'">{{ item.status }}</span>
        <span>{{ item.latencyMs ?? 0 }} ms</span>
        <span class="text-xs text-graphite-500">{{ new Date(item.checkedAt).toLocaleString() }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
