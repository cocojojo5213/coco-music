<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import PlayerIcons from './icons/PlayerIcons.vue'

const route = useRoute()
const router = useRouter()

const tabs = [
  { name: 'home', label: '现在就听', path: '/', icon: 'home' as const },
  { name: 'search', label: '搜索', path: '/search', icon: 'search' as const },
  { name: 'library', label: '资料库', path: '/library', icon: 'library' as const },
]

function go(path: string) {
  if (route.path !== path) router.push(path)
}
</script>

<template>
  <nav
    class="glass fixed inset-x-0 bottom-0 z-40 mx-auto max-w-lg border-t border-white/5"
    style="padding-bottom: env(safe-area-inset-bottom, 0px)"
  >
    <div class="grid h-[54px] grid-cols-3 px-1">
      <button
        v-for="tab in tabs"
        :key="tab.name"
        type="button"
        class="flex flex-col items-center justify-center gap-1 rounded-xl transition active:opacity-80"
        :class="route.name === tab.name ? 'text-accent' : 'text-muted'"
        @click="go(tab.path)"
      >
        <PlayerIcons :name="tab.icon" :size="20" />
        <span class="text-[10px] font-medium leading-none tracking-wide">{{ tab.label }}</span>
      </button>
    </div>
  </nav>
</template>
