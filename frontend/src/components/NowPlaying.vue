<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { usePlayerStore } from '@/stores/player'
import { useLibraryStore } from '@/stores/library'
import { useUiStore } from '@/stores/ui'
import { api, formatTime } from '@/api/client'
import CoverArt from './CoverArt.vue'
import LyricsScroller from './LyricsScroller.vue'
import PlayerIcons from './icons/PlayerIcons.vue'
import { coverOf } from '@/lib/cover'
import { canClientDirect } from '@/lib/directMedia'
import { tick } from '@/lib/haptics'

const player = usePlayerStore()
const library = useLibraryStore()
const ui = useUiStore()
const track = computed(() => player.current)
const lrc = ref('')
const lyricsLoading = ref(false)
const busy = ref(false)
const focusLyrics = ref(true)
const err = ref('')

// drag-to-dismiss
const dragY = ref(0)
const dragging = ref(false)
let startY = 0
let startX = 0
let activePointer = false

const isFav = computed(() => (track.value ? library.isFavorite(track.value) : false))
const isDl = computed(() => (track.value ? library.isDownloaded(track.value) : false))
const directOk = computed(() => (track.value ? canClientDirect(track.value) || isDl.value : false))
const duration = computed(() => player.duration || track.value?.duration || 0)
const progressPct = computed(() => {
  if (!duration.value) return 0
  return Math.min(100, (player.currentTime / duration.value) * 100)
})
const sheetStyle = computed(() => {
  if (!dragY.value) return undefined
  return {
    transform: `translateY(${dragY.value}px)`,
    transition: dragging.value ? 'none' : 'transform 0.28s cubic-bezier(0.22,1,0.36,1)',
  }
})
const backdropStyle = computed(() => {
  const o = Math.max(0.15, 0.6 - dragY.value / 700)
  return { opacity: String(o) }
})

watch(
  () => [player.showNowPlaying, track.value?.id] as const,
  async ([open, id]) => {
    document.body.style.overflow = open ? 'hidden' : ''
    dragY.value = 0
    dragging.value = false
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
  const was = library.isFavorite(track.value)
  library.toggleFavorite(track.value)
  tick('success')
  ui.ok(was ? '已取消收藏' : '已收藏')
}

async function dl() {
  if (!track.value) return
  err.value = ''
  if (!isDl.value && !canClientDirect(track.value)) {
    err.value = '无CDN直链，无法下载'
    ui.error(err.value)
    return
  }
  busy.value = true
  try {
    if (isDl.value) {
      await library.removeDownload(track.value)
      ui.ok('已移除下载')
    } else {
      await library.download(track.value)
      tick('success')
      ui.ok('已下载到本机')
    }
  } catch (e) {
    err.value = e instanceof Error ? e.message : '下载失败'
    ui.error(err.value)
  } finally {
    busy.value = false
  }
}

function onSeek(e: Event) {
  const v = Number((e.target as HTMLInputElement).value)
  player.seek(v)
}

function seekLyric(t: number) {
  player.seek(Math.max(0, t))
}

function close() {
  player.showNowPlaying = false
  dragY.value = 0
}

function onPointerDown(e: PointerEvent) {
  // only from top handle / header zone (data-drag)
  const t = e.target as HTMLElement
  if (!t.closest('[data-drag-handle]')) return
  activePointer = true
  dragging.value = true
  startY = e.clientY
  startX = e.clientX
  dragY.value = 0
  ;(e.currentTarget as HTMLElement).setPointerCapture?.(e.pointerId)
}

function onPointerMove(e: PointerEvent) {
  if (!activePointer) return
  const dy = e.clientY - startY
  const dx = Math.abs(e.clientX - startX)
  if (dx > 48 && dy < 20) return
  dragY.value = Math.max(0, dy)
}

function onPointerUp() {
  if (!activePointer) return
  activePointer = false
  dragging.value = false
  if (dragY.value > 120) {
    tick('light')
    close()
  } else {
    dragY.value = 0
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="sheet">
      <div
        v-if="player.showNowPlaying && track"
        class="fixed inset-0 z-50 flex items-end justify-center"
        @click.self="close"
      >
        <div class="absolute inset-0 bg-black/60 backdrop-blur-[3px]" :style="backdropStyle" @click="close" />
        <div
          class="relative flex h-[min(96dvh,860px)] w-full max-w-lg flex-col overflow-hidden rounded-t-[30px] bg-gradient-to-b from-[#321f27] via-ink to-ink px-5 pt-3"
          :style="[
            { paddingBottom: 'max(1rem, env(safe-area-inset-bottom, 0px))' },
            sheetStyle || {},
          ]"
          @pointerdown="onPointerDown"
          @pointermove="onPointerMove"
          @pointerup="onPointerUp"
          @pointercancel="onPointerUp"
        >
          <div data-drag-handle class="shrink-0 pb-1 pt-0.5">
            <button
              type="button"
              class="mx-auto mb-2 block h-1.5 w-12 rounded-full bg-white/25"
              aria-label="下滑关闭"
              @click="close"
            />
          </div>

          <div data-drag-handle class="flex shrink-0 items-center gap-3">
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
                class="inline-flex h-10 items-center gap-1 rounded-full bg-white/10 px-3 text-[12px] font-medium text-white/90 transition active:scale-95 disabled:opacity-40"
                :disabled="busy || (!isDl && !directOk)"
                @click="dl"
              >
                <PlayerIcons v-if="busy" name="spinner" :size="14" />
                <PlayerIcons v-else-if="isDl" name="check" :size="14" />
                <PlayerIcons v-else name="download" :size="14" />
                {{ isDl ? '已下载' : busy ? '下载中' : '下载' }}
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
                class="inline-flex h-10 items-center gap-1.5 rounded-full bg-white/10 px-4 text-sm font-medium transition active:scale-95 disabled:opacity-40"
                :disabled="busy || (!isDl && !directOk)"
                @click="dl"
              >
                <PlayerIcons v-if="busy" name="spinner" :size="15" />
                <PlayerIcons v-else-if="isDl" name="check" :size="15" />
                <PlayerIcons v-else name="download" :size="15" />
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

          <p v-if="err || player.error" class="mt-2 text-center text-[11px] text-accent">
            {{ err || player.error }}
          </p>

          <div
            class="mt-3 min-h-0 flex-1 overflow-hidden rounded-2xl bg-white/[0.04] px-2"
            :class="focusLyrics ? '' : 'max-h-[28vh]'"
          >
            <LyricsScroller
              :lrc="lrc"
              :current-time="player.currentTime"
              :loading="lyricsLoading"
              @seek="seekLyric"
            />
          </div>

          <div class="shrink-0 pt-4">
            <div class="relative h-4">
              <div class="absolute inset-x-0 top-1/2 h-1 -translate-y-1/2 rounded-full bg-white/15">
                <div
                  class="h-full rounded-full bg-accent transition-[width] duration-100"
                  :style="{ width: progressPct + '%' }"
                />
              </div>
              <input
                class="progress-abs absolute inset-0 w-full"
                type="range"
                min="0"
                :max="duration > 0 ? duration : 1"
                step="0.1"
                :value="Math.min(player.currentTime, duration || 0)"
                :disabled="!duration"
                @input="onSeek"
              />
            </div>
            <div class="mt-1 flex justify-between text-[11px] tabular-nums text-muted">
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
                <PlayerIcons v-if="player.loading" name="spinner" :size="28" />
                <span v-else :class="player.playing ? '' : 'translate-x-[1.5px]'">
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

            <div class="mt-4 flex items-center justify-center gap-4 text-[12px] text-muted/90">
              <button type="button" class="active:opacity-70" @click="focusLyrics = !focusLyrics">
                {{ focusLyrics ? '显示大封面' : '专注歌词' }}
              </button>
              <span class="text-white/15">·</span>
              <span class="tabular-nums">{{ player.index + 1 }}/{{ player.queue.length }}</span>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.sheet-enter-active,
.sheet-leave-active {
  transition: opacity 0.22s ease;
}
.sheet-enter-active .relative,
.sheet-leave-active .relative {
  transition: transform 0.28s cubic-bezier(0.22, 1, 0.36, 1);
}
.sheet-enter-from,
.sheet-leave-to {
  opacity: 0;
}
.sheet-enter-from .relative,
.sheet-leave-to .relative {
  transform: translateY(18%);
}

.progress-abs {
  appearance: none;
  background: transparent;
  height: 100%;
  margin: 0;
  cursor: pointer;
}
.progress-abs::-webkit-slider-thumb {
  appearance: none;
  width: 14px;
  height: 14px;
  border-radius: 999px;
  background: white;
  box-shadow: 0 0 0 3px rgba(252, 60, 68, 0.28);
}
.progress-abs::-webkit-slider-runnable-track {
  height: 100%;
  background: transparent;
}
.progress-abs::-moz-range-thumb {
  width: 14px;
  height: 14px;
  border: 0;
  border-radius: 999px;
  background: white;
}
.progress-abs::-moz-range-track {
  background: transparent;
  height: 4px;
}
</style>
