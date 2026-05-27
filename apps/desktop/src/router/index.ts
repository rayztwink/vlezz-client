import { createRouter, createWebHistory } from 'vue-router'
import VlessPage from '@/pages/VlessPage.vue'
import ZapretPage from '@/pages/ZapretPage.vue'
import RoutingPage from '@/pages/RoutingPage.vue'
import DiagnosticsPage from '@/pages/DiagnosticsPage.vue'
import SettingsPage from '@/pages/SettingsPage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/vless' },
    { path: '/dashboard', redirect: '/vless' },
    { path: '/vless', component: VlessPage },
    { path: '/zapret', component: ZapretPage },
    { path: '/routing', component: RoutingPage },
    { path: '/diagnostics', component: DiagnosticsPage },
    { path: '/settings', component: SettingsPage }
  ]
})

export default router
