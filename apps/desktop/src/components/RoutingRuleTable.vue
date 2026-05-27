<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { PlusIcon, TrashIcon } from '@heroicons/vue/24/outline'
import type { ActiveMode, RoutingRule } from '@/types/api'

defineProps<{ rules: RoutingRule[] }>()
const emit = defineEmits<{ create: [domain: string, mode: ActiveMode]; delete: [id: string] }>()

const domain = ref('')
const mode = ref<ActiveMode>('proxy')
const { t } = useI18n()

function submit() {
  if (!domain.value.trim()) {
    return
  }
  emit('create', domain.value.trim(), mode.value)
  domain.value = ''
}
</script>

<template>
  <div class="panel rounded-lg">
    <div class="border-b border-black/10 px-4 py-3">
      <h2 class="text-sm font-semibold">{{ t('routing.rules') }}</h2>
      <p class="text-xs text-graphite-500">{{ t('routing.description') }}</p>
    </div>
    <form class="grid gap-2 border-b border-black/10 p-4 sm:grid-cols-[1fr_8rem_auto]" @submit.prevent="submit">
      <input v-model="domain" class="focus-ring min-w-0 rounded-md border border-black/10 bg-white px-3 py-2 text-sm" :placeholder="t('routing.domainPlaceholder')" />
      <select v-model="mode" class="focus-ring rounded-md border border-black/10 bg-white px-3 py-2 text-sm">
        <option value="direct">{{ t('modes.direct') }}</option>
        <option value="proxy">{{ t('modes.proxy') }}</option>
        <option value="zapret">{{ t('modes.zapret') }}</option>
      </select>
      <button class="focus-ring grid h-10 w-full place-items-center rounded-md bg-teal-600 text-white hover:bg-teal-700 sm:w-10" :title="t('actions.add')" type="submit">
        <PlusIcon class="h-5 w-5" />
      </button>
    </form>
    <div class="divide-y divide-black/5">
      <div v-if="rules.length === 0" class="px-4 py-8 text-center text-sm text-graphite-500">{{ t('routing.empty') }}</div>
      <div v-for="rule in rules" :key="rule.id" class="grid gap-3 px-4 py-3 sm:flex sm:items-center sm:justify-between">
        <div class="min-w-0">
          <div class="text-sm font-medium">{{ rule.domain }}</div>
          <div class="text-xs text-graphite-500">{{ t('routing.enabled', { value: rule.enabled ? t('common.yes') : t('common.no') }) }}</div>
        </div>
        <div class="flex items-center justify-between gap-2 sm:justify-end">
          <span class="badge" :class="rule.mode === 'proxy' ? 'badge-ok' : rule.mode === 'zapret' ? 'badge-warn' : 'badge-muted'">{{ rule.mode }}</span>
          <button class="focus-ring grid h-8 w-8 place-items-center rounded-md border border-black/10 text-red-600 hover:bg-red-50" :title="t('actions.delete')" @click="emit('delete', rule.id)">
            <TrashIcon class="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
