import { describe, it, expect, beforeEach, vi } from 'vitest'
import { check, mockUpdaterConfig, setupUpdaterMockIPC } from './mocks/updater'
import { invoke } from '@tauri-apps/api/core'
import { clearMocks } from '@tauri-apps/api/mocks'

describe('Updater Mock Smoke Test', () => {
  beforeEach(() => {
    mockUpdaterConfig.reset()
    clearMocks()
  })

  describe('JavaScript/TypeScript API Check', () => {
    it('should return null when state is no-update', async () => {
      mockUpdaterConfig.setState('no-update')
      const result = await check()
      expect(result).toBeNull()
    })

    it('should return update details when state is update', async () => {
      mockUpdaterConfig.setState('update')
      mockUpdaterConfig.setVersion('2.0.0')

      const result = await check()
      expect(result).not.toBeNull()
      expect(result?.version).toBe('2.0.0')
      expect(result?.body).toBe('Simulated update release notes')

      // Test download progress callbacks
      const progressCallback = vi.fn()
      await result?.downloadAndInstall(progressCallback)

      expect(mockUpdaterConfig.wasDownloadedAndInstalled).toBe(true)
      expect(mockUpdaterConfig.relaunchCount).toBe(1)
      expect(progressCallback).toHaveBeenCalled()
      expect(progressCallback).toHaveBeenCalledWith({ event: 'Started', data: { contentLength: 100 } })
      expect(progressCallback).toHaveBeenLastCalledWith({ event: 'Finished' })
    })

    it('should throw an error when state is error', async () => {
      mockUpdaterConfig.setState('error')
      mockUpdaterConfig.setCheckError('Network connection timed out')

      await expect(check()).rejects.toThrow('Network connection timed out')
    })
  })

  describe('IPC Mocking via mockIPC', () => {
    it('should simulate updater check over IPC', async () => {
      setupUpdaterMockIPC()

      mockUpdaterConfig.setState('update')
      mockUpdaterConfig.setVersion('3.0.0-beta.1')

      const ipcResult = await invoke<{ version: string }>('plugin:updater|check')
      expect(ipcResult).not.toBeNull()
      expect(ipcResult.version).toBe('3.0.0-beta.1')
    })

    it('should simulate no-update over IPC', async () => {
      setupUpdaterMockIPC()

      mockUpdaterConfig.setState('no-update')

      const ipcResult = await invoke('plugin:updater|check')
      expect(ipcResult).toBeNull()
    })

    it('should simulate check error over IPC', async () => {
      setupUpdaterMockIPC()

      mockUpdaterConfig.setState('error')
      mockUpdaterConfig.setCheckError('Server internal error')

      await expect(invoke('plugin:updater|check')).rejects.toThrow('Server internal error')
    })

    it('should simulate download, progress, and relaunch over IPC', async () => {
      setupUpdaterMockIPC()
      
      mockUpdaterConfig.setState('update')

      const progressCallback = vi.fn()
      // Register a callback ID with window.__TAURI_INTERNALS__.transformCallback
      const callbackId = (window as any).__TAURI_INTERNALS__.transformCallback(progressCallback)

      await invoke('plugin:updater|download_and_install', { onProgress: callbackId })

      expect(mockUpdaterConfig.wasDownloadedAndInstalled).toBe(true)
      expect(mockUpdaterConfig.relaunchCount).toBe(1)
      expect(progressCallback).toHaveBeenCalled()
      expect(progressCallback).toHaveBeenCalledWith({ event: 'Started', data: { contentLength: 100 } })
      expect(progressCallback).toHaveBeenLastCalledWith({ event: 'Finished' })
    })
  })
})
