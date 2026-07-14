<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { usePlayerStore } from '@/stores/player'
import { useLibraryStore } from '@/stores/library'
import { api, formatTime } from '@/api/client'
import CoverArt from './CoverArt.vue'
import LyricsScroller from './LyricsScroller.vue'
import PlayerIcons from './icons/PlayerIcons.vue'
import { coverOf } from '@/lib/cover'
import { canClientDirect } from '@/lib/directMedia'
import { trackKey } from '@/lib/localLibrary'

const player = usePlayerStore()
const library = useLibraryStore()
const track = computed(() => player.current)
const lrc = ref('')
const lyricsLoading = ref(false)
const busy = ref(false)
const focusLyrics = ref(true)
const err = ref('')

const isFav = computed(() => (track.value ? library.isFavorite(track.value) : false))
const isDl = computed(() => {
  if (!track.value) return false
  const key = trackKey(track.value)
  return library.downloads.some((t) => trackKey(t) === key)
})
const directOk = computed(() => (track.value ? canClientDirect(track.value) || isDl.value : false))
const duration = computed(
  () => player.duration || track.value?.duration || 0,
)

watch(
  () => [player.showNowPlaying, track.value?.id] as const,
  async ([open, id]) => {
    if (!open || !id || !track.value) return
    err.value = ''
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

function fav() {
  if (!track.value) return
  library.toggleFavorite(track.value)
}

async function dl() {
  if (!track.value) return
  err.value = ''
  if (!isDl.value && !canClientDirect(track.value)) {
    err.value = '无CDN直链，无法下载'
    return
  }
  busy.value = true
  try {
    if (isDl.value) await library.removeDownload(track.value)
    else await library.download(track.value)
  } catch (e) {
    err.value = e instanceof Error ? e.message : '下载失败'
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
      class="fixed inset-0 z-50 flex items-end justify-center bg-black/60 backdrop-blur-[3px]"
      @click.self="player.showNowPlaying = false"
    >
      <div
        class="relative flex h-[min(96dvh,860px)] w-full max-w-lg flex-col overflow-hidden rounded-t-[30px] bg-gradient-to-b from-[#2c1c22] via-ink to-ink px-5 pt-3"
        style="padding-bottom: max(1rem, env(safe-area-inset-bottom, 0px))"
      >
        <button
          type="button"
          class="mx-auto mb-3 block h-1.5 w-12 shrink-0 rounded-full bg-white/25"
          aria-label="收起"
          @click="player.showNowPlaying = false"
        />

        <!-- compact header + optional cover -->
        <div class="flex shrink-0 items-center gap-3">
          <button
            type="button"
            class="shrink-0 transition active:scale-[0.98]"
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
          <div v-if="focusLyrics" class="flex shrink-0 items-center gap-1">
            <button
              type="button"
              class="transport-btn h-10 w-10 rounded-full bg-white/10"
              :class="isFav ? 'text-accent' : 'text-white/90'"
              aria-label="收藏"
              @click="fav"
            >
              <PlayerIcons :name="isFav ? 'heart-fill' : 'heart'" :size="18" />
            </button>
            <button
              type="button"
              class="h-10 rounded-full bg-white/10 px-3 text-[12px] font-medium text-white/90 transition active:scale-95 disabled:opacity-40"
              :disabled="busy || (!isDl && !directOk)"
              @click="dl"
            >
              {{ isDl ? '已下载' : busy ? '…' : '下载' }}
            </button>
          </div>
        </div>

        <div v-if="!focusLyrics" class="mt-4 shrink-0 text-center">
          <h2 class="truncate px-2 text-[22px] font-bold leading-tight">{{ track.title }}</h2>
          <p class="mt-1 truncate px-2 text-[15px] text-accent">{{ track.artist }}</p>
          <div class="mt-3 flex justify-center gap-2">
            <button
              type="button"
              class="transport-btn h-10 w-10 rounded-full bg-white/10"
              :class="isFav ? 'text-accent' : 'text-white/90'"
              aria-label="收藏"
              @click="fav"
            >
              <PlayerIcons :name="isFav ? 'heart-fill' : 'heart'" :size="18" />
            </button>
            <button
              type="button"
              class="h-10 rounded-full bg-white/10 px-4 text-sm font-medium transition active:scale-95 disabled:opacity-40"
              :disabled="busy || (!isDl && !directOk)"
              @click="dl"
            >
              {{ isDl ? '已下载' : busy ? '下载中' : '下载' }}
            </button>
            <button
              type="button"
              class="h-10 rounded-full bg-white/10 px-4 text-sm font-medium transition active:scale-95"
              @click="focusLyrics = true"
            >
              歌词
            </button>
          </div>
        </div>

        <p v-if="err" class="mt-2 text-center text-[11px] text-accent">{{ err }}</p>

        <!-- scrolling lyrics stage -->
        <div
          class="mt-3 min-h-0 flex-1 overflow-hidden rounded-2xl bg-white/[0.04] px-2"
          :class="focusLyrics ? '' : 'max-h-[28vh]'"
        >
          <LyricsScroller
            :lrc="lrc"
            :current-time="player.currentTime"
            :loading="lyricsLoading"
          />
        </div>

        <!-- transport -->
        <div class="shrink-0 pt-4">
          <input
            class="progress w-full"
            type="range"
            min="0"
            :max="duration > 0 ? duration : 1"
            step="0.1"
            :value="Math.min(player.currentTime, duration || 0)"
            :disabled="!duration"
            @input="onSeek"
          />
          <div class="mt-2 flex justify-between text-[11px] tabular-nums text-muted">
            <span>{{ formatTime(player.currentTime) }}</span>
            <span>{{ formatTime(duration) }}</span>
          </div>

          <div class="mt-5 flex items-center justify-center gap-10">
            <button
              type="button"
              class="transport-btn h-12 w-12 text-white/90"
              aria-label="上一首"
              @click="player.prev()"
            >
              <PlayerIcons name="prev" :size="30" />
            </button>

            <button
              type="button"
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
              type="button"
              class="transport-btn h-12 w-12 text-white/90"
              aria-label="下一首"
              @click="player.next()"
            >
              <PlayerIcons name="next" :size="30" />
            </button>
          </div>

          <button
            type="button"
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
