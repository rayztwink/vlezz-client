<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ArrowPathIcon, TrashIcon } from '@heroicons/vue/24/outline'
import type { Subscription } from '@/types/api'

defineProps<{ subscriptions: Subscription[]; embedded?: boolean }>()
const emit = defineEmits<{ update: [id: string]; delete: [id: string] }>()
const { t } = useI18n()

function maskUrl(url: string) {
  try {
    const parsed = new URL(url)
    return `${parsed.origin}/...`
  } catch {
    return '...'
  }
}
</script>

<template>
  <div :class="embedded ? 'overflow-hidden rounded-md border border-black/10' : 'panel overflow-hidden rounded-lg'">
    <div class="border-b border-black/10 px-4 py-3">
      <h2 class="text-base font-semibold">{{ t('subscriptions.title') }}</h2>
      <p class="text-xs text-graphite-500">{{ t('subscriptions.masked') }}</p>
    </div>
    <div class="divide-y divide-black/5">
      <div v-if="subscriptions.length === 0" class="px-4 py-8 text-center text-sm text-graphite-500">{{ t('subscriptions.empty') }}</div>
      <div v-for="sub in subscriptions" :key="sub.id" class="grid gap-3 px-4 py-3 sm:flex sm:items-center sm:justify-between sm:gap-4">
        <div class="min-w-0">
          <div class="truncate text-sm font-medium">{{ sub.name }}</div>
          <div class="truncate text-xs text-graphite-500">{{ maskUrl(sub.url) }}</div>
        </div>
        <div class="flex shrink-0 items-center justify-between gap-2 sm:justify-end">
          <span class="badge badge-muted">{{ sub.updateInterval }}m</span>
          <button class="focus-ring grid h-8 w-8 place-items-center rounded-md border border-black/10 hover:bg-graphite-50" :title="t('subscriptions.updateTitle')" @click="emit('update', sub.id)">
            <ArrowPathIcon class="h-4 w-4" />
          </button>
          <button class="focus-ring grid h-8 w-8 place-items-center rounded-md border border-black/10 text-red-600 hover:bg-red-50" :title="t('actions.delete')" @click="emit('delete', sub.id)">
            <TrashIcon class="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
