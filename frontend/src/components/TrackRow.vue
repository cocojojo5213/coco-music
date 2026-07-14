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

const directOk = computed(() => canClientDirect(props.track) || library.isDownloaded(props.track))
const isFav = computed(() => library.isFavorite(props.track))
const isDl = computed(() => library.isDownloaded(props.track))
const isCurrent = computed(
  () => !!player.current && trackKey(player.current) === trackKey(props.track),
)
const isPlaying = computed(() => isCurrent.value && player.playing)

async function play() {
  const q = props.queue?.length ? props.queue : [props.track]
  await player.playTracks(q, props.track.id)
}

function fav() {
  library.toggleFavorite(props.track)
}

async function dl() {
  err.value = ''
  if (!isDl.value && !canClientDirect(props.track)) {
    err.value = '无直链，无法下载'
    return
  }
  busy.value = true
  try {
    if (isDl.value) await library.removeDownload(props.track)
    else await library.download(props.track)
  } catch (ex) {
    err.value = ex instanceof Error ? ex.message : '下载失败'
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div
    class="rounded-2xl transition-colors"
    :class="isCurrent ? 'bg-white/[0.06]' : 'hover:bg-white/[0.03]'"
  >
    <div class="flex w-full items-center gap-1.5 px-1 py-1.5">
      <button
        type="button"
        class="flex min-w-0 flex-1 items-center gap-3 rounded-xl px-1.5 py-1 text-left active:bg-white/5"
        @click="play"
      >
        <div class="relative shrink-0">
          <CoverArt :src="coverOf(track)" size="sm" rounded="rounded-[10px]" />
          <span
            v-if="isPlaying"
            class="absolute inset-0 flex items-center justify-center rounded-[10px] bg-black/40 text-white"
          >
            <span class="eq" aria-hidden="true">
              <i /><i /><i />
            </span>
          </span>
        </div>
        <div class="min-w-0 flex-1">
          <div
            class="truncate text-[15px] font-semibold leading-tight"
            :class="isCurrent ? 'text-accent' : ''"
          >
            {{ track.title }}
          </div>
          <div class="mt-0.5 truncate text-[12px] leading-tight text-muted">
            {{ track.artist }}
            <template v-if="track.duration"> · {{ formatTime(track.duration) }}</template>
            <template v-if="track.audioBytes"> · {{ formatBytes(track.audioBytes) }}</template>
            <template v-if="isDl"> · 本地</template>
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
          class="transport-btn h-10 w-10"
          :class="isDl ? 'text-accent' : !directOk ? 'text-white/25' : busy ? 'text-white/40' : 'text-muted'"
          :title="isDl ? '移除下载' : directOk ? '直链下载' : '无CDN直链'"
          :aria-label="isDl ? '移除下载' : '下载'"
          :disabled="busy || (!isDl && !directOk)"
          @click="dl"
        >
          <PlayerIcons v-if="busy" name="spinner" :size="18" />
          <PlayerIcons v-else-if="isDl" name="check" :size="18" />
          <PlayerIcons v-else name="download" :size="18" />
        </button>
      </div>
    </div>
    <p v-if="err" class="px-3 pb-2 text-[11px] text-accent">{{ err }}</p>
  </div>
</template>

<style scoped>
.eq {
  display: flex;
  align-items: flex-end;
  gap: 2px;
  height: 12px;
}
.eq i {
  display: block;
  width: 2.5px;
  border-radius: 1px;
  background: #fff;
  animation: eq 0.9s ease-in-out infinite;
}
.eq i:nth-child(1) {
  height: 5px;
  animation-delay: 0s;
}
.eq i:nth-child(2) {
  height: 11px;
  animation-delay: 0.15s;
}
.eq i:nth-child(3) {
  height: 7px;
  animation-delay: 0.3s;
}
@keyframes eq {
  0%,
  100% {
    transform: scaleY(0.45);
  }
  50% {
    transform: scaleY(1);
  }
}
</style>
