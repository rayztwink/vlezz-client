<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { SignalIcon, TrashIcon } from '@heroicons/vue/24/outline'
import type { Node } from '@/types/api'

defineProps<{ nodes: Node[]; selectedNodeId?: string; loading?: boolean }>()

const emit = defineEmits<{
  select: [id: string]
  check: [id: string]
  delete: [id: string]
}>()

const { t } = useI18n()

function maskUuid(uuid: string) {
  return uuid.length > 8 ? `${uuid.slice(0, 8)}...` : uuid
}
</script>

<template>
  <section class="panel flex min-h-[18rem] flex-col overflow-hidden rounded-lg sm:min-h-[32rem]">
    <div class="border-b border-black/10 px-4 py-3">
      <div class="flex items-center justify-between gap-3">
        <h2 class="text-base font-semibold">{{ t('profiles.title') }}</h2>
        <span class="badge badge-muted">{{ nodes.length }}</span>
      </div>
    </div>

    <div class="min-h-0 flex-1 overflow-auto p-2">
      <div v-if="loading" class="grid h-36 place-items-center text-sm text-graphite-500">{{ t('profiles.loading') }}</div>
      <div v-else-if="nodes.length === 0" class="grid h-36 place-items-center px-4 text-center text-sm text-graphite-500">
        {{ t('profiles.empty') }}
      </div>

      <button
        v-for="node in nodes"
        v-else
        :key="node.id"
        class="focus-ring mb-2 grid w-full gap-2 rounded-md border px-3 py-3 text-left transition"
        :class="selectedNodeId === node.id ? 'border-teal-300 bg-teal-50' : 'border-transparent hover:border-black/10 hover:bg-graphite-50'"
        type="button"
        @click="emit('select', node.id)"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="truncate text-sm font-semibold">{{ node.name }}</div>
            <div class="mt-1 truncate font-mono text-xs text-graphite-500">{{ node.address }}:{{ node.port }}</div>
          </div>
          <span v-if="selectedNodeId === node.id" class="badge badge-ok shrink-0">{{ t('common.active') }}</span>
        </div>

        <div class="flex flex-wrap items-center gap-2 text-xs text-graphite-500">
          <span class="font-mono">{{ maskUuid(node.uuid) }}</span>
          <span class="rounded bg-white px-1.5 py-0.5">{{ node.security }}</span>
          <span class="rounded bg-white px-1.5 py-0.5">{{ node.transport }}</span>
          <span class="ml-auto font-medium text-graphite-700">{{ node.latencyMs ? `${node.latencyMs} ms` : t('common.unknown') }}</span>
        </div>

        <div class="flex justify-end gap-2">
          <button class="focus-ring grid h-8 w-8 place-items-center rounded-md border border-black/10 bg-white hover:bg-graphite-50" type="button" :title="t('profiles.checkLatency')" @click.stop="emit('check', node.id)">
            <SignalIcon class="h-4 w-4" />
          </button>
          <button class="focus-ring grid h-8 w-8 place-items-center rounded-md border border-black/10 bg-white text-red-600 hover:bg-red-50" type="button" :title="t('actions.delete')" @click.stop="emit('delete', node.id)">
            <TrashIcon class="h-4 w-4" />
          </button>
        </div>
      </button>
    </div>
  </section>
</template>
