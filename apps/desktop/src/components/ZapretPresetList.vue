<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { PlayIcon, StopIcon } from '@heroicons/vue/24/outline'
import type { ZapretPreset } from '@/types/api'

defineProps<{ presets: ZapretPreset[] }>()
const emit = defineEmits<{ start: [id: string]; stop: [] }>()
const { t } = useI18n()
</script>

<template>
  <div class="panel rounded-lg">
    <div class="flex items-center justify-between border-b border-black/10 px-4 py-3 dark:border-white/10">
      <div>
        <h2 class="text-sm font-semibold">{{ t('zapret.presets') }}</h2>
        <p class="text-xs text-graphite-500 dark:text-graphite-300">{{ t('zapret.presetsDescription') }}</p>
      </div>
      <button class="focus-ring grid h-8 w-8 place-items-center rounded-md border border-black/10 hover:bg-black/5 dark:border-white/10 dark:hover:bg-white/8" :title="t('zapret.stopTitle')" @click="emit('stop')">
        <StopIcon class="h-4 w-4" />
      </button>
    </div>
    <div class="divide-y divide-black/5 dark:divide-white/8">
      <div v-if="presets.length === 0" class="px-4 py-8 text-center text-sm text-graphite-500">{{ t('zapret.empty') }}</div>
      <div v-for="preset in presets" :key="preset.id" class="flex items-center justify-between gap-4 px-4 py-3">
        <div class="min-w-0">
          <div class="flex items-center gap-2">
            <span class="truncate text-sm font-medium">{{ preset.name }}</span>
            <span v-if="preset.isActive" class="badge badge-ok">{{ t('common.active') }}</span>
          </div>
          <div class="mt-1 truncate text-xs text-graphite-500">{{ preset.description || preset.source }}</div>
        </div>
        <button class="focus-ring grid h-8 w-8 shrink-0 place-items-center rounded-md bg-teal-600 text-white hover:bg-teal-700" :title="t('zapret.startTitle')" @click="emit('start', preset.id)">
          <PlayIcon class="h-4 w-4" />
        </button>
      </div>
    </div>
  </div>
</template>
