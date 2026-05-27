<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

defineProps<{ open: boolean }>()

const emit = defineEmits<{
  close: []
  submit: [link: string, name?: string]
}>()

const link = ref('')
const name = ref('')
const { t } = useI18n()

function submit() {
  emit('submit', link.value.trim(), name.value.trim() || undefined)
  link.value = ''
  name.value = ''
}
</script>

<template>
  <div v-if="open" class="fixed inset-0 z-50 grid place-items-center overflow-auto bg-black/35 p-3 backdrop-blur-sm sm:p-4">
    <form class="panel max-h-[calc(100vh-2rem)] w-full max-w-2xl overflow-auto rounded-lg p-4 sm:p-5" @submit.prevent="submit">
      <div class="mb-5">
        <h2 class="text-lg font-semibold">{{ t('import.title') }}</h2>
        <p class="mt-1 text-sm text-graphite-500">{{ t('import.description') }}</p>
      </div>
      <label class="mb-3 block">
        <span class="mb-1 block text-sm font-medium">{{ t('import.displayName') }}</span>
        <input v-model="name" class="focus-ring w-full rounded-md border border-black/10 bg-white px-3 py-2" :placeholder="t('import.optional')" />
      </label>
      <label class="block">
        <span class="mb-1 block text-sm font-medium">{{ t('import.vlessLink') }}</span>
        <textarea v-model="link" class="focus-ring h-32 w-full resize-none rounded-md border border-black/10 bg-white px-3 py-2 font-mono text-sm" placeholder="vless://..." required />
      </label>
      <div class="mt-5 grid gap-2 sm:flex sm:justify-end">
        <button class="focus-ring rounded-md border border-black/10 px-4 py-2 text-sm hover:bg-graphite-50" type="button" @click="emit('close')">{{ t('actions.cancel') }}</button>
        <button class="focus-ring rounded-md bg-teal-600 px-4 py-2 text-sm font-medium text-white hover:bg-teal-700" type="submit">{{ t('actions.import') }}</button>
      </div>
    </form>
  </div>
</template>
