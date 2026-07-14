<script setup lang="ts">
import { computed } from 'vue'
import { usePlayerStore } from '@/stores/player'
import CoverArt from './CoverArt.vue'
import PlayerIcons from './icons/PlayerIcons.vue'
import { coverOf } from '@/lib/cover'

const player = usePlayerStore()
const track = computed(() => player.current)
const progress = computed(() => {
  if (!player.duration) return 0
  return Math.min(100, (player.currentTime / player.duration) * 100)
})
</script>

<template>
  <div
    v-if="track"
    class="glass fixed inset-x-3 z-40 mx-auto max-w-lg overflow-hidden rounded-2xl border border-white/10 shadow-[0_8px_28px_rgba(0,0,0,0.35)]"
    style="bottom: calc(3.6rem + env(safe-area-inset-bottom))"
  >
    <div class="h-[2px] bg-white/10">
      <div class="h-full bg-accent transition-[width] duration-200 ease-linear" :style="{ width: progress + '%' }" />
    </div>
    <div class="flex h-14 items-center gap-1.5 px-2">
      <button class="flex min-w-0 flex-1 items-center gap-2.5 pl-0.5" @click="player.showNowPlaying = true">
        <CoverArt :src="coverOf(track)" size="xs" rounded="rounded-lg" />
        <div class="min-w-0 text-left">
          <div class="truncate text-[13px] font-semibold leading-tight">{{ track.title }}</div>
          <div class="truncate text-[11px] leading-tight text-muted">{{ track.artist }}</div>
        </div>
      </button>

      <button
        class="transport-btn h-11 w-11 text-white"
        :aria-label="player.playing ? '暂停' : '播放'"
        @click="player.toggle()"
      >
        <PlayerIcons :name="player.playing ? 'pause' : 'play'" :size="22" />
      </button>
      <button
        class="transport-btn h-11 w-11 text-white/80"
        aria-label="下一首"
        @click="player.next()"
      >
        <PlayerIcons name="next" :size="22" />
      </button>
    </div>
  </div>
</template>
