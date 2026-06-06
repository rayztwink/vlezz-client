import { describe, it, expect } from 'vitest'
import fs from 'fs'
import path from 'path'

const cargoPath = path.resolve(__dirname, '../../src-tauri/Cargo.toml')
const tauriConfPath = path.resolve(__dirname, '../../src-tauri/tauri.conf.json')

describe('Configuration Tests', () => {
  // F1-T1-1: Cargo.toml dependency check
  it('F1-T1-1: should verify Cargo.toml exists and lists Tauri dependencies correctly', () => {
    expect(fs.existsSync(cargoPath)).toBe(true)
    const content = fs.readFileSync(cargoPath, 'utf8')
    
    // Check for [dependencies] section
    expect(content).toContain('[dependencies]')
    
    // Check for tauri and tauri-plugin-updater
    const lines = content.split('\n')
    let inDependencies = false
    let hasTauri = false
    let hasUpdater = false
    
    for (const line of lines) {
      const trimmed = line.trim()
      if (trimmed.startsWith('[dependencies]')) {
        inDependencies = true
        continue
      }
      if (trimmed.startsWith('[') && trimmed.endsWith(']')) {
        inDependencies = false
      }
      if (inDependencies) {
        if (trimmed.startsWith('tauri ')) {
          hasTauri = true
        }
        if (trimmed.startsWith('tauri-plugin-updater')) {
          hasUpdater = true
        }
      }
    }
    
    expect(hasTauri).toBe(true)
    expect(hasUpdater).toBe(true)
  })

  // F1-T1-2: tauri.conf.json JSON parse validation
  it('F1-T1-2: should parse tauri.conf.json as valid JSON', () => {
    expect(fs.existsSync(tauriConfPath)).toBe(true)
    const content = fs.readFileSync(tauriConfPath, 'utf8')
    expect(() => JSON.parse(content)).not.toThrow()
  })

  // F1-T1-3: Cargo.toml version vs tauri.conf.json version check
  it('F1-T1-3: should verify Cargo.toml version matches tauri.conf.json version', () => {
    const cargoContent = fs.readFileSync(cargoPath, 'utf8')
    const tauriConf = JSON.parse(fs.readFileSync(tauriConfPath, 'utf8'))
    
    // Extract version from Cargo.toml [package]
    const packageSection = cargoContent.match(/\[package\]([\s\S]*?)(?:\[\w+\]|$)/)
    expect(packageSection).not.toBeNull()
    const versionMatch = packageSection![1].match(/version\s*=\s*"([^"]+)"/)
    expect(versionMatch).not.toBeNull()
    
    const cargoVersion = versionMatch![1]
    const tauriVersion = tauriConf.version
    
    expect(cargoVersion).toBe(tauriVersion)
  })

  // F1-T1-4: tauri.conf.json plugins updater schema check
  it('F1-T1-4: should verify tauri.conf.json plugins updater schema is correct', () => {
    const tauriConf = JSON.parse(fs.readFileSync(tauriConfPath, 'utf8'))
    
    expect(tauriConf.plugins).toBeDefined()
    expect(tauriConf.plugins.updater).toBeDefined()
    expect(Array.isArray(tauriConf.plugins.updater.endpoints)).toBe(true)
    expect(tauriConf.plugins.updater.endpoints.length).toBeGreaterThan(0)
    expect(typeof tauriConf.plugins.updater.pubkey).toBe('string')
  })

  // F1-T1-5: tauri.conf.json icons bundle check
  it('F1-T1-5: should verify tauri.conf.json bundle contains correct icons', () => {
    const tauriConf = JSON.parse(fs.readFileSync(tauriConfPath, 'utf8'))
    
    expect(tauriConf.bundle).toBeDefined()
    expect(Array.isArray(tauriConf.bundle.icon)).toBe(true)
    expect(tauriConf.bundle.icon).toContain('icons/icon.ico')
    expect(tauriConf.bundle.icon).toContain('icons/icon.icns')
  })

  // F1-T2-1: Tauri config custom endpoint syntax check
  it('F1-T2-1: should verify that custom updater endpoints are valid URLs', () => {
    const tauriConf = JSON.parse(fs.readFileSync(tauriConfPath, 'utf8'))
    const endpoints: string[] = tauriConf.plugins?.updater?.endpoints || []
    
    expect(endpoints.length).toBeGreaterThan(0)
    for (const endpoint of endpoints) {
      expect(() => new URL(endpoint)).not.toThrow()
    }
  })

  // F1-T2-2: Secure HTTPS update endpoint check in tauri.conf.json (production)
  it('F1-T2-2: should verify that production update endpoints use HTTPS', () => {
    const tauriConf = JSON.parse(fs.readFileSync(tauriConfPath, 'utf8'))
    const endpoints: string[] = tauriConf.plugins?.updater?.endpoints || []
    
    for (const endpoint of endpoints) {
      const url = new URL(endpoint)
      // Only allow secure protocol for production update checks
      expect(url.protocol).toBe('https:')
    }
  })
})
