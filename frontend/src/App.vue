<script setup lang="ts">
import { onMounted } from 'vue'
import { useLibraryStore } from '@/stores/library'
import TabBar from '@/components/TabBar.vue'
import MiniPlayer from '@/components/MiniPlayer.vue'
import NowPlaying from '@/components/NowPlaying.vue'

const library = useLibraryStore()

onMounted(() => {
  library.loadHot()
})
</script>

<template>
  <div class="min-h-dvh bg-ink text-white">
    <main class="mx-auto min-h-dvh max-w-lg pb-[calc(7.5rem+env(safe-area-inset-bottom,0px))]">
      <RouterView />
      <footer class="px-4 pb-4 pt-10 text-center">
        <div class="text-[12px] font-semibold tracking-wide text-white/70">摇摆熊 · Coco Music</div>
        <div class="mt-1 text-[10px] text-muted">Made by 摇摆熊</div>
      </footer>
    </main>
    <MiniPlayer />
    <TabBar />
    <NowPlaying />
  </div>
</template>
