import { defineStore } from 'pinia'
import { ref } from 'vue'
import { check } from '@tauri-apps/plugin-updater'
import { getVersion } from '@tauri-apps/api/app'
import { invoke } from '@tauri-apps/api/core'

function compareSemver(v1: string, v2: string): number {
  const clean = (v: string) => v.trim().replace(/^v/, '');
  const parts1 = clean(v1).split('.');
  const parts2 = clean(v2).split('.');
  const maxLen = Math.max(parts1.length, parts2.length);
  for (let i = 0; i < maxLen; i++) {
    const p1 = parts1[i] || '0';
    const p2 = parts2[i] || '0';
    
    const subParts1 = p1.split('-');
    const subParts2 = p2.split('-');
    
    const num1 = parseInt(subParts1[0], 10);
    const num2 = parseInt(subParts2[0], 10);
    
    const isNum1 = !isNaN(num1);
    const isNum2 = !isNaN(num2);
    
    if (isNum1 && isNum2) {
      if (num1 !== num2) {
        return num1 > num2 ? 1 : -1;
      }
    } else {
      if (subParts1[0] !== subParts2[0]) {
        return subParts1[0] > subParts2[0] ? 1 : -1;
      }
    }
    
    if (subParts1.length > 1 || subParts2.length > 1) {
      if (subParts1.length === 1) return 1;
      if (subParts2.length === 1) return -1;
      
      const tag1 = subParts1.slice(1).join('-');
      const tag2 = subParts2.slice(1).join('-');
      if (tag1 !== tag2) {
        const match1 = tag1.match(/^([a-zA-Z\-]+)(\d+)$/);
        const match2 = tag2.match(/^([a-zA-Z\-]+)(\d+)$/);
        if (match1 && match2) {
          const alpha1 = match1[1];
          const alpha2 = match2[1];
          if (alpha1 !== alpha2) {
            return alpha1 > alpha2 ? 1 : -1;
          }
          const val1 = parseInt(match1[2], 10);
          const val2 = parseInt(match2[2], 10);
          if (val1 !== val2) {
            return val1 > val2 ? 1 : -1;
          }
        } else {
          return tag1 > tag2 ? 1 : -1;
        }
      }
    }
  }
  return 0;
}

export const useUpdaterStore = defineStore('updater', () => {
  const currentVersion = ref<string>('0.1.0')
  const checking = ref(false)
  const cooldownActive = ref(false)
  const downloading = ref(false)
  const progressPercent = ref(0)
  const downloadedBytes = ref(0)
  const totalBytes = ref(0)
  const updateAvailable = ref(false)
  const newVersion = ref<string>('')
  const updateBody = ref<string>('')
  const downloadedAndInstalled = ref(false)
  const error = ref<string | null>(null)
  const status = ref<string>('idle') // 'idle', 'checking', 'up-to-date', 'available', 'downloading', 'installing', 'error'
  const relaunchCancelled = ref(false)

  let pendingUpdate: any = null

  async function loadCurrentVersion() {
    try {
      currentVersion.value = await getVersion()
    } catch (err: any) {
      console.warn('Failed to retrieve application version from Tauri:', err)
      currentVersion.value = '0.1.0'
    }
  }

  async function checkForUpdates() {
    if (downloading.value) return
    if (checking.value) return
    checking.value = true
    error.value = null
    updateAvailable.value = false
    newVersion.value = ''
    updateBody.value = ''
    pendingUpdate = null
    status.value = 'checking'
    relaunchCancelled.value = false

    try {
      const update = await check()
      if (update) {
        const isUpgrade = compareSemver(update.version, currentVersion.value) > 0
        if (isUpgrade) {
          updateAvailable.value = true
          newVersion.value = update.version
          updateBody.value = update.body || ''
          pendingUpdate = update
          status.value = 'available'
        } else {
          updateAvailable.value = false
          status.value = 'up-to-date'
        }
      } else {
        updateAvailable.value = false
        status.value = 'up-to-date'
      }
    } catch (err: any) {
      console.error('Failed checking for updates:', err)
      error.value = err?.message || String(err)
      status.value = 'error'
    } finally {
      setTimeout(() => {
        checking.value = false
      }, 0)
      cooldownActive.value = true
      setTimeout(() => {
        cooldownActive.value = false
      }, 30000)
    }
  }

  async function downloadAndInstall() {
    if (downloading.value) return
    if (!pendingUpdate) return
    downloading.value = true
    progressPercent.value = 0
    downloadedBytes.value = 0
    totalBytes.value = 0
    error.value = null
    status.value = 'downloading'
    relaunchCancelled.value = false

    try {
      let contentLength = 0
      let downloaded = 0

      await pendingUpdate.downloadAndInstall((event: any) => {
        if (!event) return
        if (event.event === 'Started') {
          contentLength = event.data?.contentLength || 0
          totalBytes.value = contentLength
          downloaded = 0
          downloadedBytes.value = 0
        } else if (event.event === 'Progress') {
          downloaded += event.data?.chunkLength || 0
          downloadedBytes.value = downloaded
          if (contentLength > 0) {
            progressPercent.value = Math.round((downloaded / contentLength) * 100)
          }
        } else if (event.event === 'Finished') {
          progressPercent.value = 100
          status.value = 'installing'
        }
      })

      // Download and install completed
      progressPercent.value = 100
      downloading.value = false
      downloadedAndInstalled.value = true
      status.value = 'installing'

      // Include automatic confirm() prompt on successful download/install completion to relaunch
      let userConfirmed = false
      try {
        userConfirmed = confirm('Update downloaded and installed successfully. Relaunch now?')
      } catch (confirmErr) {
        console.warn('window.confirm is not supported or failed:', confirmErr)
      }
      if (userConfirmed) {
        await relaunchApp()
      } else {
        relaunchCancelled.value = true
      }
    } catch (err: any) {
      console.error('Failed to download or install update:', err)
      error.value = err?.message || String(err)
      status.value = 'error'
      downloading.value = false
    }
  }

  async function relaunchApp() {
    try {
      // The relaunch action should invoke the custom command 'relaunch' with a fallback to 'plugin:process|relaunch'
      await invoke('relaunch')
    } catch (err: any) {
      console.error('Custom relaunch command failed, trying fallback:', err)
      try {
        await invoke('plugin:process|relaunch')
      } catch (fallbackErr: any) {
        console.error('Fallback relaunch command failed:', fallbackErr)
        error.value = fallbackErr?.message || String(fallbackErr)
        status.value = 'error'
      }
    }
  }

  return {
    currentVersion,
    checking,
    cooldownActive,
    downloading,
    progressPercent,
    downloadedBytes,
    totalBytes,
    updateAvailable,
    newVersion,
    updateBody,
    downloadedAndInstalled,
    error,
    status,
    relaunchCancelled,
    loadCurrentVersion,
    checkForUpdates,
    downloadAndInstall,
    relaunchApp
  }
})
