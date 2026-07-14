<script setup lang="ts">
import { computed } from 'vue'
import { usePlayerStore } from '@/stores/player'
import CoverArt from './CoverArt.vue'
import PlayerIcons from './icons/PlayerIcons.vue'
import { coverOf } from '@/lib/cover'

const player = usePlayerStore()
const track = computed(() => player.current)
const progressPct = computed(() => Math.round((player.progress || 0) * 1000) / 10)
</script>

<template>
  <Transition name="mini">
    <div
      v-if="track && !player.showNowPlaying"
      class="glass fixed inset-x-3 z-40 mx-auto max-w-lg overflow-hidden rounded-2xl border border-white/10 shadow-[0_8px_28px_rgba(0,0,0,0.35)]"
      style="bottom: calc(54px + env(safe-area-inset-bottom, 0px) + 8px)"
    >
      <div class="h-[2px] bg-white/10">
        <div
          class="h-full bg-accent transition-[width] duration-150 ease-linear"
          :style="{ width: progressPct + '%' }"
        />
      </div>
      <div class="flex h-14 items-center gap-0.5 px-2">
        <button
          type="button"
          class="flex min-w-0 flex-1 items-center gap-2.5 pl-0.5 text-left"
          @click="player.showNowPlaying = true"
        >
          <div class="relative">
            <CoverArt :src="coverOf(track)" size="xs" rounded="rounded-lg" />
            <span
              v-if="player.loading"
              class="absolute inset-0 flex items-center justify-center rounded-lg bg-black/35 text-white"
            >
              <PlayerIcons name="spinner" :size="14" />
            </span>
          </div>
          <div class="min-w-0">
            <div class="truncate text-[13px] font-semibold leading-tight">{{ track.title }}</div>
            <div class="truncate text-[11px] leading-tight text-muted">
              {{ player.error || track.artist }}
              <template v-if="!player.error && player.queue.length > 1">
                · {{ player.index + 1 }}/{{ player.queue.length }}
              </template>
            </div>
          </div>
        </button>

        <button
          type="button"
          class="transport-btn h-11 w-11 text-white"
          :aria-label="player.playing ? '暂停' : '播放'"
          @click="player.toggle()"
        >
          <PlayerIcons v-if="player.loading" name="spinner" :size="20" />
          <span v-else :class="player.playing ? '' : 'translate-x-[1px]'">
            <PlayerIcons :name="player.playing ? 'pause' : 'play'" :size="22" />
          </span>
        </button>
        <button
          type="button"
          class="transport-btn h-11 w-11 text-white/80"
          aria-label="下一首"
          @click="player.next()"
        >
          <PlayerIcons name="next" :size="22" />
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.mini-enter-active,
.mini-leave-active {
  transition: all 0.22s ease;
}
.mini-enter-from,
.mini-leave-to {
  opacity: 0;
  transform: translateY(12px);
}
</style>
