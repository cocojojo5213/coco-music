<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Track } from '@/types'
import { formatBytes, formatTime } from '@/api/client'
import CoverArt from './CoverArt.vue'
import PlayerIcons from './icons/PlayerIcons.vue'
import { usePlayerStore } from '@/stores/player'
import { useLibraryStore } from '@/stores/library'
import { coverOf } from '@/lib/cover'
import { canClientDirect } from '@/lib/directMedia'
import { trackKey } from '@/lib/localLibrary'

const props = defineProps<{
  track: Track
  queue?: Track[]
}>()

const player = usePlayerStore()
const library = useLibraryStore()
const busy = ref(false)
const err = ref('')

const directOk = computed(() => canClientDirect(props.track) || !!props.track.isDownloaded)
const isFav = computed(() => library.isFavorite(props.track))
const isPlaying = computed(() => player.current && trackKey(player.current) === trackKey(props.track))

async function play() {
  const q = props.queue?.length ? props.queue : [props.track]
  await player.playTracks(q, props.track.id)
}

function fav() {
  library.toggleFavorite(props.track)
}

async function dl() {
  err.value = ''
  if (!props.track.isDownloaded && !canClientDirect(props.track)) {
    err.value = '无直链，无法下载（不经本站中转）'
    return
  }
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
  <div class="rounded-2xl transition" :class="isPlaying ? 'bg-white/[0.06]' : ''">
    <div class="flex w-full items-center gap-2.5 px-1.5 py-2">
      <button
        type="button"
        class="flex min-w-0 flex-1 items-center gap-3 rounded-xl px-1 py-0.5 text-left active:bg-white/5"
        @click="play"
      >
        <div class="relative shrink-0">
          <CoverArt :src="coverOf(track)" size="sm" rounded="rounded-[10px]" />
          <span
            v-if="isPlaying && player.playing"
            class="absolute inset-0 flex items-center justify-center rounded-[10px] bg-black/35 text-[11px] text-white"
            >▶</span
          >
        </div>
        <div class="min-w-0 flex-1">
          <div
            class="truncate text-[15px] font-semibold leading-tight"
            :class="isPlaying ? 'text-accent' : ''"
          >
            {{ track.title }}
          </div>
          <div class="mt-0.5 truncate text-[12px] leading-tight text-muted">
            {{ track.artist }}
            <template v-if="track.duration"> · {{ formatTime(track.duration) }}</template>
            <template v-if="track.audioBytes"> · {{ formatBytes(track.audioBytes) }}</template>
            <template v-if="track.isDownloaded"> · 本地</template>
          </div>
        </div>
      </button>

      <div class="flex shrink-0 items-center">
        <button
          type="button"
          class="transport-btn h-10 w-10"
          :class="isFav ? 'text-accent' : 'text-muted'"
          :aria-label="isFav ? '取消收藏' : '收藏'"
          @click="fav"
        >
          <PlayerIcons :name="isFav ? 'heart-fill' : 'heart'" :size="18" />
        </button>
        <button
          type="button"
          class="transport-btn h-10 w-10 text-[15px] font-semibold"
          :class="
            track.isDownloaded
              ? 'text-accent'
              : !directOk
                ? 'text-white/25'
                : busy
                  ? 'text-white/40'
                  : 'text-muted'
          "
          :title="track.isDownloaded ? '移除下载' : directOk ? '直链下载' : '无CDN直链'"
          :aria-label="track.isDownloaded ? '移除下载' : '下载'"
          :disabled="busy || (!track.isDownloaded && !directOk)"
          @click="dl"
        >
          {{ track.isDownloaded ? '✓' : busy ? '…' : '↓' }}
        </button>
      </div>
    </div>
    <p v-if="err" class="px-3 pb-2 text-[11px] text-accent">{{ err }}</p>
  </div>
</template>
