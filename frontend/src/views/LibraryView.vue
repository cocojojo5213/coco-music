<script setup lang="ts">
import { ref } from 'vue'
import { useLibraryStore } from '@/stores/library'
import TrackRow from '@/components/TrackRow.vue'

const library = useLibraryStore()
const tab = ref<'favorites' | 'downloads'>('favorites')
</script>

<template>
  <div class="safe-top px-4">
    <header class="mb-5 pt-2">
      <h1 class="text-3xl font-bold">Library</h1>
      <p class="mt-1 text-sm text-muted">摇摆熊 · 收藏与下载保存在你的浏览器本地</p>
    </header>

    <div class="mb-4 grid grid-cols-2 gap-2 rounded-2xl bg-white/5 p-1">
      <button
        class="rounded-xl px-3 py-2 text-sm font-semibold"
        :class="tab === 'favorites' ? 'bg-white text-black' : 'text-muted'"
        @click="tab = 'favorites'"
      >
        收藏 {{ library.favorites.length }}
      </button>
      <button
        class="rounded-xl px-3 py-2 text-sm font-semibold"
        :class="tab === 'downloads' ? 'bg-white text-black' : 'text-muted'"
        @click="tab = 'downloads'"
      >
        已下载 {{ library.downloads.length }}
      </button>
    </div>

    <div v-if="tab === 'favorites'" class="glass-card rounded-3xl p-2">
      <TrackRow
        v-for="t in library.favorites"
        :key="t.id"
        :track="t"
        :queue="library.favorites"
      />
      <div v-if="!library.favorites.length" class="py-16 text-center text-muted">
        还没有收藏
      </div>
    </div>

    <div v-else class="glass-card rounded-3xl p-2">
      <TrackRow
        v-for="t in library.downloads"
        :key="t.id"
        :track="t"
        :queue="library.downloads"
      />
      <div v-if="!library.downloads.length" class="py-16 text-center text-muted">
        还没有下载。点 ↓ 会保存到本机，并缓存供离线播放。
      </div>
    </div>
  </div>
</template>
