<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  Cog6ToothIcon,
  CommandLineIcon,
  QueueListIcon,
  ServerStackIcon,
  ShieldCheckIcon
} from '@heroicons/vue/24/outline'
import { useSettingsStore } from '@/stores/settings'

const route = useRoute()
const settings = useSettingsStore()
const { t } = useI18n()

const navItems = [
  { to: '/vless', labelKey: 'nav.client', icon: ServerStackIcon },
  { to: '/zapret', labelKey: 'nav.zapret', icon: ShieldCheckIcon },
  { to: '/routing', labelKey: 'nav.routing', icon: QueueListIcon },
  { to: '/diagnostics', labelKey: 'nav.checks', icon: CommandLineIcon },
  { to: '/settings', labelKey: 'nav.settings', icon: Cog6ToothIcon }
]

const currentSection = computed(() => t(navItems.find((item) => item.to === route.path)?.labelKey ?? 'app.name'))

onMounted(async () => {
  await settings.load()
  if (settings.settings?.theme !== 'light') {
    await settings.patch({ theme: 'light' })
  }
})
</script>

<template>
  <div class="flex h-screen flex-col bg-[#f4f4f1] text-graphite-900">
    <header class="shrink-0 border-b border-black/10 bg-white/95 px-3 sm:px-5">
      <div class="mx-auto flex max-w-7xl flex-col gap-3 py-3 sm:flex-row sm:items-center sm:justify-between sm:py-0">
        <div class="flex min-w-0 items-center justify-between gap-3 sm:h-16">
          <div class="flex min-w-0 items-center gap-3">
          <div class="grid h-9 w-9 place-items-center rounded-md bg-[#13998f] text-sm font-semibold text-white">RF</div>
          <div class="min-w-0">
            <div class="text-sm font-semibold">{{ t('app.name') }}</div>
            <div class="text-xs text-graphite-500">{{ currentSection }}</div>
          </div>
          </div>
          <span class="badge badge-ok shrink-0 sm:hidden">{{ t('app.api') }}</span>
        </div>

        <nav class="-mx-1 flex min-w-0 max-w-full items-center gap-1 overflow-x-auto rounded-md border border-black/10 bg-graphite-50 p-1 sm:mx-0">
          <RouterLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="focus-ring inline-flex h-9 shrink-0 items-center gap-2 rounded px-3 text-sm font-medium transition"
            :class="route.path === item.to ? 'bg-white text-graphite-900 shadow-sm' : 'text-graphite-600 hover:text-graphite-900'"
          >
            <component :is="item.icon" class="h-4 w-4" />
            <span>{{ t(item.labelKey) }}</span>
          </RouterLink>
        </nav>

        <span class="badge badge-ok hidden shrink-0 sm:inline-flex">{{ t('app.apiFull') }}</span>
      </div>
    </header>

    <main class="min-h-0 flex-1 overflow-auto px-3 py-4 sm:px-5 sm:py-5">
      <RouterView />
    </main>

    <footer class="shrink-0 border-t border-black/10 bg-white/80 px-3 py-2 text-center text-[11px] leading-relaxed text-graphite-500 sm:px-5 sm:text-xs">
      {{ t('app.legal') }}
    </footer>
  </div>
</template>
