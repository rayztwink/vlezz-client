import { defineStore } from 'pinia'
import { ref } from 'vue'
import { rayflowApi } from '@/services/api'
import type { DiagnosticCheck, Node } from '@/types/api'

export const useNodesStore = defineStore('nodes', () => {
  const nodes = ref<Node[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const lastCheck = ref<DiagnosticCheck | null>(null)

  async function load() {
    loading.value = true
    error.value = null
    try {
      nodes.value = await rayflowApi.nodes()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load nodes'
    } finally {
      loading.value = false
    }
  }

  async function importNode(link: string, name?: string) {
    const node = await rayflowApi.importNode({ link, name })
    nodes.value = [node, ...nodes.value]
  }

  async function deleteNode(id: string) {
    await rayflowApi.deleteNode(id)
    nodes.value = nodes.value.filter((node) => node.id !== id)
  }

  async function checkNode(id: string) {
    lastCheck.value = await rayflowApi.checkNode(id)
    await load()
  }

  async function connectNode(id: string, core = 'sing-box') {
    await rayflowApi.connectNodeWithOptions({ nodeId: id, core })
  }

  async function disconnect() {
    await rayflowApi.disconnect()
  }

  return { nodes, loading, error, lastCheck, load, importNode, deleteNode, checkNode, connectNode, disconnect }
})
