import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import SearchView from '@/views/SearchView.vue'
import LibraryView from '@/views/LibraryView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/search', name: 'search', component: SearchView },
    { path: '/library', name: 'library', component: LibraryView },
  ],
  scrollBehavior(to, from, saved) {
    if (saved) return saved
    if (to.path === from.path) return false
    return { top: 0, behavior: 'smooth' }
  },
})

export default router
