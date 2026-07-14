<script setup lang="ts">
import { ref } from 'vue'
import { useLibraryStore } from '@/stores/library'
import TrackRow from '@/components/TrackRow.vue'
import PlayerIcons from '@/components/icons/PlayerIcons.vue'

const library = useLibraryStore()
const tab = ref<'favorites' | 'downloads'>('favorites')
</script>

<template>
  <div class="safe-top px-4">
    <header class="mb-5 pt-2">
      <p class="text-[13px] text-muted">摇摆熊</p>
      <h1 class="text-[28px] font-bold leading-tight tracking-tight">资料库</h1>
      <p class="mt-1 text-sm text-muted">收藏与下载保存在你的浏览器本地</p>
    </header>

    <div class="mb-4 grid grid-cols-2 gap-1 rounded-2xl bg-white/5 p-1">
      <button
        type="button"
        class="rounded-xl px-3 py-2.5 text-sm font-semibold transition"
        :class="tab === 'favorites' ? 'bg-white text-black shadow-sm' : 'text-muted'"
        @click="tab = 'favorites'"
      >
        收藏 {{ library.favorites.length }}
      </button>
      <button
        type="button"
        class="rounded-xl px-3 py-2.5 text-sm font-semibold transition"
        :class="tab === 'downloads' ? 'bg-white text-black shadow-sm' : 'text-muted'"
        @click="tab = 'downloads'"
      >
        已下载 {{ library.downloads.length }}
      </button>
    </div>

    <div v-if="tab === 'favorites'" class="glass-card rounded-3xl p-1.5">
      <TrackRow
        v-for="t in library.favorites"
        :key="t.id"
        :track="t"
        :queue="library.favorites"
      />
      <div v-if="!library.favorites.length" class="px-4 py-16 text-center text-muted">
        <div class="mb-3 flex justify-center text-white/25">
          <PlayerIcons name="heart" :size="28" />
        </div>
        还没有收藏
        <div class="mt-1 text-xs">点列表里的心形即可收藏</div>
      </div>
    </div>

    <div v-else class="glass-card rounded-3xl p-1.5">
      <TrackRow
        v-for="t in library.downloads"
        :key="t.id"
        :track="t"
        :queue="library.downloads"
      />
      <div v-if="!library.downloads.length" class="px-4 py-16 text-center text-muted">
        <div class="mb-3 flex justify-center text-white/25">
          <PlayerIcons name="download" :size="28" />
        </div>
        还没有下载
        <div class="mt-1 text-xs">点下载图标直链保存，并可离线播放</div>
      </div>
    </div>
  </div>
</template>
