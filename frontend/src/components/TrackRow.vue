<script setup lang="ts">
import { ref } from 'vue'
import type { Track } from '@/types'
import { formatBytes, formatTime } from '@/api/client'
import CoverArt from './CoverArt.vue'
import { usePlayerStore } from '@/stores/player'
import { useLibraryStore } from '@/stores/library'
import { coverOf } from '@/lib/cover'

const props = defineProps<{
  track: Track
  queue?: Track[]
}>()

const player = usePlayerStore()
const library = useLibraryStore()
const busy = ref(false)
const err = ref('')

async function play() {
  const q = props.queue?.length ? props.queue : [props.track]
  await player.playTracks(q, props.track.id)
}

function fav(e: Event) {
  e.stopPropagation()
  library.toggleFavorite(props.track)
}

async function dl(e: Event) {
  e.stopPropagation()
  err.value = ''
  busy.value = true
  try {
    if (props.track.isDownloaded) await library.removeDownload(props.track)
    else await library.download(props.track)
  } catch (ex) {
    err.value = ex instanceof Error ? ex.message : '下载失败'
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div>
    <button
      class="flex w-full items-center gap-3 rounded-2xl px-2 py-2.5 text-left transition active:bg-white/5"
      @click="play"
    >
      <CoverArt :src="coverOf(track)" size="sm" rounded="rounded-[10px]" />
      <div class="min-w-0 flex-1">
        <div class="truncate text-[15px] font-semibold leading-tight">{{ track.title }}</div>
        <div class="mt-0.5 truncate text-[12px] leading-tight text-muted">
          {{ track.artist }}
          <template v-if="track.duration"> · {{ formatTime(track.duration) }}</template>
          <template v-if="track.audioBytes"> · {{ formatBytes(track.audioBytes) }}</template>
          <template v-if="track.isDownloaded"> · 本地</template>
        </div>
      </div>
      <div class="flex shrink-0 items-center gap-0.5">
        <span
          class="inline-flex h-9 w-9 items-center justify-center rounded-full text-[15px]"
          :class="track.isFavorite ? 'text-accent' : 'text-muted'"
          @click="fav"
          >♥</span
        >
        <span
          class="inline-flex h-9 w-9 items-center justify-center rounded-full text-[15px]"
          :class="busy ? 'text-white/40' : 'text-muted'"
          @click="dl"
          >{{ track.isDownloaded ? '✓' : busy ? '…' : '↓' }}</span
        >
      </div>
    </button>
    <p v-if="err" class="px-3 pb-1 text-[11px] text-accent">{{ err }}</p>
  </div>
</template>
