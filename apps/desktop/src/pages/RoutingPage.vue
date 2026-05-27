<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import RoutingRuleTable from '@/components/RoutingRuleTable.vue'
import { useRoutingStore } from '@/stores/routing'

const routing = useRoutingStore()
const { t } = useI18n()

onMounted(() => {
  void routing.load()
})
</script>

<template>
  <div class="mx-auto grid max-w-7xl gap-4 sm:gap-5">
    <div>
      <h1 class="text-xl font-semibold tracking-normal sm:text-2xl">{{ t('routing.title') }}</h1>
      <p class="mt-1 text-sm text-graphite-500">{{ t('routing.subtitle') }}</p>
    </div>
    <RoutingRuleTable :rules="routing.rules" @create="routing.create" @delete="routing.remove" />
  </div>
</template>
