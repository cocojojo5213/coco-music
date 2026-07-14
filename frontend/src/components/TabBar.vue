<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const tabs = [
  { name: 'home', label: 'Listen Now', path: '/', icon: '♪' },
  { name: 'search', label: 'Search', path: '/search', icon: '⌕' },
  { name: 'library', label: 'Library', path: '/library', icon: '☰' },
]

function go(path: string) {
  router.push(path)
}
</script>

<template>
  <nav class="glass fixed inset-x-0 bottom-0 z-40 mx-auto h-[58px] max-w-lg safe-bottom">
    <div class="grid h-full grid-cols-3 px-1">
      <button
        v-for="tab in tabs"
        :key="tab.name"
        class="flex flex-col items-center justify-center gap-0.5 rounded-xl transition"
        :class="route.name === tab.name ? 'text-accent' : 'text-muted'"
        @click="go(tab.path)"
      >
        <span class="text-[17px] leading-none">{{ tab.icon }}</span>
        <span class="text-[10px] font-medium leading-none">{{ tab.label }}</span>
      </button>
    </div>
  </nav>
</template>
