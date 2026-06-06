import { mockIPC } from '@tauri-apps/api/mocks'

// Global mock state
export type MockUpdaterState = 'update' | 'no-update' | 'error';

let currentState: MockUpdaterState = 'no-update';
let currentVersion = '1.0.1';
let currentCheckError: string | null = null;
let relaunchCount = 0;
let downloadProgressCalls: any[] = [];
let wasDownloadedAndInstalled = false;
let checkDelay = 0;
let downloadError: string | null = null;

export const mockUpdaterConfig = {
  get state() {
    return currentState;
  },
  get version() {
    return currentVersion;
  },
  get relaunchCount() {
    return relaunchCount;
  },
  get wasDownloadedAndInstalled() {
    return wasDownloadedAndInstalled;
  },
  get downloadProgressCalls() {
    return downloadProgressCalls;
  },
  get checkDelay() {
    return checkDelay;
  },
  get downloadError() {
    return downloadError;
  },
  setState(state: MockUpdaterState) {
    currentState = state;
  },
  setVersion(version: string) {
    currentVersion = version;
  },
  setCheckError(errMessage: string) {
    currentCheckError = errMessage;
  },
  setCheckDelay(delayMs: number) {
    checkDelay = delayMs;
  },
  setDownloadError(errMessage: string | null) {
    downloadError = errMessage;
  },
  reset() {
    currentState = 'no-update';
    currentVersion = '1.0.1';
    currentCheckError = null;
    relaunchCount = 0;
    downloadProgressCalls = [];
    wasDownloadedAndInstalled = false;
    checkDelay = 0;
    downloadError = null;
  }
};

// Simulated update object
export interface MockUpdate {
  version: string;
  date: string;
  body: string;
  downloadAndInstall: (
    onProgress?: (event: DownloadProgressEvent) => void
  ) => Promise<void>;
  close: () => Promise<void>;
}

export type DownloadProgressEvent =
  | { event: 'Started'; data: { contentLength?: number } }
  | { event: 'Progress'; data: { chunkLength: number } }
  | { event: 'Finished' };

// Mocked implementation of check()
export async function check(): Promise<MockUpdate | null> {
  if (checkDelay > 0) {
    await new Promise((resolve) => setTimeout(resolve, checkDelay));
  }
  if (currentState === 'error') {
    throw new Error(currentCheckError || 'Failed to check for updates');
  }
  if (currentState === 'no-update') {
    return null;
  }

  return {
    version: currentVersion,
    date: new Date().toISOString(),
    body: 'Simulated update release notes',
    downloadAndInstall: async (onProgress) => {
      if (downloadError) {
        throw new Error(downloadError);
      }
      wasDownloadedAndInstalled = true;
      if (onProgress) {
        onProgress({ event: 'Started', data: { contentLength: 100 } });
        downloadProgressCalls.push({ event: 'Started', data: { contentLength: 100 } });

        // Simulate chunk progress
        for (let i = 0; i < 5; i++) {
          onProgress({ event: 'Progress', data: { chunkLength: 20 } });
          downloadProgressCalls.push({ event: 'Progress', data: { chunkLength: 20 } });
        }

        onProgress({ event: 'Finished' });
        downloadProgressCalls.push({ event: 'Finished' });
      }
      relaunchCount++;
    },
    close: async () => {}
  };
}

// IPC Simulation setup helper
export function setupUpdaterMockIPC() {
  mockIPC((cmd, payload: any) => {
    switch (cmd) {
      case 'plugin:updater|check':
        if (checkDelay > 0) {
          return new Promise((resolve, reject) => {
            setTimeout(() => {
              if (currentState === 'error') {
                reject(new Error(currentCheckError || 'Failed to check for updates'));
              } else if (currentState === 'no-update') {
                resolve(null);
              } else {
                resolve({
                  version: currentVersion,
                  date: new Date().toISOString(),
                  body: 'Simulated update release notes via IPC'
                });
              }
            }, checkDelay);
          });
        }
        if (currentState === 'error') {
          throw new Error(currentCheckError || 'Failed to check for updates');
        }
        if (currentState === 'no-update') {
          return null;
        }
        return {
          version: currentVersion,
          date: new Date().toISOString(),
          body: 'Simulated update release notes via IPC'
        };

      case 'plugin:updater|download_and_install':
        if (downloadError) {
          throw new Error(downloadError);
        }
        wasDownloadedAndInstalled = true;
        // In Tauri v2, download_and_install usually receives a Channel/callback ID.
        // If we want to simulate progress callbacks, let's call the channel.
        const callbackId = payload?.onProgress || payload?.handler;
        if (callbackId && typeof window !== 'undefined' && (window as any).__TAURI_INTERNALS__?.runCallback) {
          const runCallback = (window as any).__TAURI_INTERNALS__.runCallback;
          runCallback(callbackId, { event: 'Started', data: { contentLength: 100 } });
          for (let i = 0; i < 5; i++) {
            runCallback(callbackId, { event: 'Progress', data: { chunkLength: 20 } });
          }
          runCallback(callbackId, { event: 'Finished' });
        }
        relaunchCount++;
        return;

      case 'plugin:process|relaunch':
      case 'plugin:updater|relaunch':
      case 'relaunch':
        relaunchCount++;
        return;

      case 'get_api_token':
        return 'mock-api-token';

      default:
        return;
    }
  });
}
