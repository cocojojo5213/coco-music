<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { usePlayerStore } from '@/stores/player'
import { useLibraryStore } from '@/stores/library'
import { api, formatTime } from '@/api/client'
import CoverArt from './CoverArt.vue'
import LyricsScroller from './LyricsScroller.vue'
import PlayerIcons from './icons/PlayerIcons.vue'
import { coverOf } from '@/lib/cover'

const player = usePlayerStore()
const library = useLibraryStore()
const track = computed(() => player.current)
const lrc = ref('')
const lyricsLoading = ref(false)
const busy = ref(false)
const focusLyrics = ref(true)

watch(
  () => [player.showNowPlaying, track.value?.id] as const,
  async ([open, id]) => {
    if (!open || !id || !track.value) return
    lrc.value = track.value.lrc || ''
    if (lrc.value) {
      lyricsLoading.value = false
      return
    }
    lyricsLoading.value = true
    try {
      const res = await api.lyrics(track.value)
      if (player.current?.id === id) {
        lrc.value = res.lrc || ''
        if (res.lrc) track.value.lrc = res.lrc
      }
    } catch {
      if (player.current?.id === id) lrc.value = ''
    } finally {
      if (player.current?.id === id) lyricsLoading.value = false
    }
  },
  { immediate: true },
)

async function fav() {
  if (!track.value) return
  library.toggleFavorite(track.value)
}

async function dl() {
  if (!track.value) return
  busy.value = true
  try {
    if (track.value.isDownloaded) await library.removeDownload(track.value)
    else await library.download(track.value)
  } finally {
    busy.value = false
  }
}

function onSeek(e: Event) {
  const v = Number((e.target as HTMLInputElement).value)
  player.seek(v)
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="player.showNowPlaying && track"
      class="fixed inset-0 z-50 flex items-end justify-center bg-black/55 backdrop-blur-[2px]"
      @click.self="player.showNowPlaying = false"
    >
      <div
        class="relative flex h-[min(96dvh,860px)] w-full max-w-lg flex-col overflow-hidden rounded-t-[30px] bg-gradient-to-b from-[#2c1c22] via-ink to-ink px-5 pt-3 safe-bottom"
      >
        <button
          class="mx-auto mb-3 block h-1.5 w-12 shrink-0 rounded-full bg-white/25"
          aria-label="收起"
          @click="player.showNowPlaying = false"
        />

        <!-- compact header + optional cover -->
        <div class="flex shrink-0 items-center gap-3">
          <button
            class="shrink-0 active:scale-[0.98] transition"
            :class="focusLyrics ? '' : 'mx-auto'"
            @click="focusLyrics = !focusLyrics"
          >
            <CoverArt
              :src="coverOf(track)"
              :size="focusLyrics ? 'sm' : 'now'"
              :rounded="focusLyrics ? 'rounded-xl' : 'rounded-[26px]'"
            />
          </button>
          <div v-if="focusLyrics" class="min-w-0 flex-1">
            <h2 class="truncate text-[17px] font-bold leading-tight">{{ track.title }}</h2>
            <p class="truncate text-[13px] text-accent">{{ track.artist }}</p>
          </div>
          <div v-if="focusLyrics" class="flex shrink-0 gap-1.5">
            <button
              class="transport-btn h-9 w-9 rounded-full bg-white/10"
              :class="track.isFavorite ? 'text-accent' : 'text-white/90'"
              aria-label="收藏"
              @click="fav"
            >
              <PlayerIcons :name="track.isFavorite ? 'heart-fill' : 'heart'" :size="18" />
            </button>
            <button
              class="h-9 rounded-full bg-white/10 px-3 text-[12px] font-medium text-white/90 active:scale-95 transition"
              :disabled="busy"
              @click="dl"
            >
              {{ track.isDownloaded ? '已下载' : busy ? '…' : '下载' }}
            </button>
          </div>
        </div>

        <div v-if="!focusLyrics" class="mt-4 shrink-0 text-center">
          <h2 class="truncate text-[22px] font-bold leading-tight">{{ track.title }}</h2>
          <p class="mt-1 truncate text-[15px] text-accent">{{ track.artist }}</p>
          <div class="mt-3 flex justify-center gap-2">
            <button
              class="transport-btn h-10 w-10 rounded-full bg-white/10"
              :class="track.isFavorite ? 'text-accent' : 'text-white/90'"
              aria-label="收藏"
              @click="fav"
            >
              <PlayerIcons :name="track.isFavorite ? 'heart-fill' : 'heart'" :size="18" />
            </button>
            <button
              class="h-10 rounded-full bg-white/10 px-4 text-sm font-medium active:scale-95 transition"
              :disabled="busy"
              @click="dl"
            >
              {{ track.isDownloaded ? '已下载' : busy ? '下载中' : '下载' }}
            </button>
            <button
              class="h-10 rounded-full bg-white/10 px-4 text-sm font-medium active:scale-95 transition"
              @click="focusLyrics = true"
            >
              歌词
            </button>
          </div>
        </div>

        <!-- scrolling lyrics stage -->
        <div
          class="mt-3 min-h-0 flex-1 rounded-2xl bg-white/[0.04] px-2"
          :class="focusLyrics ? '' : 'max-h-[28vh]'"
        >
          <LyricsScroller
            :lrc="lrc"
            :current-time="player.currentTime"
            :loading="lyricsLoading"
          />
        </div>

        <!-- transport -->
        <div class="shrink-0 pb-5 pt-4">
          <input
            class="progress w-full"
            type="range"
            min="0"
            :max="player.duration || track.duration || 0"
            step="0.1"
            :value="player.currentTime"
            @input="onSeek"
          />
          <div class="mt-2 flex justify-between text-[11px] tabular-nums text-muted">
            <span>{{ formatTime(player.currentTime) }}</span>
            <span>{{ formatTime(player.duration || track.duration || 0) }}</span>
          </div>

          <div class="mt-5 flex items-center justify-center gap-10">
            <button
              class="transport-btn h-12 w-12 text-white/90"
              aria-label="上一首"
              @click="player.prev()"
            >
              <PlayerIcons name="prev" :size="30" />
            </button>

            <button
              class="transport-btn play-btn h-[68px] w-[68px] rounded-full bg-white text-black shadow-[0_10px_30px_rgba(0,0,0,0.35)]"
              :aria-label="player.playing ? '暂停' : '播放'"
              @click="player.toggle()"
            >
              <span :class="player.playing ? '' : 'translate-x-[1.5px]'">
                <PlayerIcons
                  :name="player.playing ? 'pause' : 'play'"
                  :size="player.playing ? 28 : 30"
                />
              </span>
            </button>

            <button
              class="transport-btn h-12 w-12 text-white/90"
              aria-label="下一首"
              @click="player.next()"
            >
              <PlayerIcons name="next" :size="30" />
            </button>
          </div>

          <button
            class="mx-auto mt-4 block text-[12px] text-muted/90 active:opacity-70"
            @click="focusLyrics = !focusLyrics"
          >
            {{ focusLyrics ? '显示大封面' : '专注歌词' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
