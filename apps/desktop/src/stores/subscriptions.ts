import { defineStore } from 'pinia'
import { ref } from 'vue'
import { rayflowApi } from '@/services/api'
import type { Subscription, SubscriptionUpdateResult } from '@/types/api'

export const useSubscriptionsStore = defineStore('subscriptions', () => {
  const subscriptions = ref<Subscription[]>([])
  const loading = ref(false)
  const lastUpdate = ref<SubscriptionUpdateResult | null>(null)

  async function load() {
    loading.value = true
    try {
      subscriptions.value = await rayflowApi.subscriptions()
    } finally {
      loading.value = false
    }
  }

  async function create(name: string, url: string, updateInterval = 1440) {
    const sub = await rayflowApi.createSubscription({ name, url, updateInterval })
    subscriptions.value = [sub, ...subscriptions.value]
  }

  async function update(id: string) {
    lastUpdate.value = await rayflowApi.updateSubscription(id)
    await load()
  }

  async function remove(id: string) {
    await rayflowApi.deleteSubscription(id)
    subscriptions.value = subscriptions.value.filter((sub) => sub.id !== id)
  }

  return { subscriptions, loading, lastUpdate, load, create, update, remove }
})
