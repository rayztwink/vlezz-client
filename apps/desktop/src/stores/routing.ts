import { defineStore } from 'pinia'
import { ref } from 'vue'
import { rayflowApi } from '@/services/api'
import type { ActiveMode, RoutingRule } from '@/types/api'

export const useRoutingStore = defineStore('routing', () => {
  const rules = ref<RoutingRule[]>([])
  const loading = ref(false)

  async function load() {
    loading.value = true
    try {
      rules.value = await rayflowApi.routingRules()
    } finally {
      loading.value = false
    }
  }

  async function create(domain: string, mode: ActiveMode, enabled = true) {
    const rule = await rayflowApi.createRoutingRule({ domain, mode, enabled })
    rules.value = [...rules.value, rule]
  }

  async function remove(id: string) {
    await rayflowApi.deleteRoutingRule(id)
    rules.value = rules.value.filter((rule) => rule.id !== id)
  }

  return { rules, loading, load, create, remove }
})

