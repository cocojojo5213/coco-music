<script setup lang="ts">
import { computed } from 'vue'
import { useLibraryStore } from '@/stores/library'
import { usePlayerStore } from '@/stores/player'
import TrackRow from '@/components/TrackRow.vue'
import CoverArt from '@/components/CoverArt.vue'
import { coverOf } from '@/lib/cover'

const library = useLibraryStore()
const player = usePlayerStore()

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 12) return '早上好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const hero = computed(() => library.tracks[0] || null)
const rail = computed(() => library.tracks.slice(0, 12))

async function playAll() {
  await player.playTracks(library.tracks)
  player.showNowPlaying = true
}

async function refresh() {
  await library.loadHot(true)
}

async function playOne(id: string) {
  await player.playTracks(library.tracks, id)
  player.showNowPlaying = true
}

function searchCount(t: { searchCount?: number }) {
  const n = Number(t.searchCount || 0)
  if (!n) return ''
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k 次搜索`
  return `${n} 次搜索`
}
</script>

<template>
  <div class="safe-top px-4">
    <header class="mb-5 flex items-end justify-between gap-3 pt-2">
      <div class="min-w-0">
        <p class="text-[13px] text-muted">{{ greeting }} · 摇摆熊</p>
        <h1 class="text-[28px] font-bold leading-tight tracking-tight">Listen Now</h1>
      </div>
      <div class="flex shrink-0 gap-2">
        <button class="h-9 rounded-full bg-white/10 px-3 text-[13px]" @click="refresh">刷新</button>
        <button
          class="h-9 rounded-full bg-accent px-4 text-[13px] font-semibold shadow-lg shadow-accent/30"
          @click="playAll"
        >
          播放
        </button>
      </div>
    </header>

    <section v-if="library.loading" class="py-20 text-center text-muted">加载站友搜索榜…</section>
    <section v-else-if="library.error" class="py-16 text-center">
      <p class="text-accent">{{ library.error }}</p>
      <button class="mt-3 h-9 rounded-full bg-white/10 px-4 text-sm" @click="refresh">重试</button>
    </section>

    <template v-else>
      <section class="mb-5">
        <div class="glass-card rounded-3xl px-4 py-3">
          <div class="flex items-center justify-between gap-3">
            <div class="min-w-0">
              <div class="text-[11px] font-semibold uppercase tracking-[0.14em] text-accent">
                Ranking
              </div>
              <h2 class="truncate text-[22px] font-bold leading-tight">
                {{ library.chart?.name || '站友搜索榜' }}
              </h2>
              <p class="truncate text-[12px] text-muted">
                {{ library.chart?.description || '按大家实际搜索次数实时排序' }}
              </p>
            </div>
            <div class="shrink-0 rounded-2xl bg-accent/15 px-3 py-2 text-center">
              <div class="text-[18px] font-bold tabular-nums text-accent">{{ library.tracks.length }}</div>
              <div class="text-[10px] text-muted">首</div>
            </div>
          </div>
        </div>
      </section>

      <section v-if="hero" class="mb-7">
        <button class="w-full text-center" @click="playOne(hero.id)">
          <CoverArt :src="coverOf(hero)" size="hero" rounded="rounded-[22px]" />
          <div class="mt-3 px-2">
            <div class="text-[11px] font-semibold uppercase tracking-[0.14em] text-accent">
              TOP 1 · {{ searchCount(hero) || '站友热搜' }}
            </div>
            <div class="mt-1 truncate text-[18px] font-bold leading-tight">{{ hero.title }}</div>
            <div class="truncate text-[13px] text-muted">{{ hero.artist }}</div>
          </div>
        </button>
      </section>

      <section v-if="rail.length" class="mb-7">
        <div class="mb-3 flex items-center justify-between">
          <h2 class="text-[20px] font-bold leading-none">封面速览</h2>
        </div>
        <div class="no-scrollbar -mx-4 flex gap-3 overflow-x-auto px-4 pb-1">
          <button
            v-for="(t, idx) in rail"
            :key="'rail-' + t.id"
            class="w-[118px] shrink-0 text-left"
            @click="playOne(t.id)"
          >
            <div class="relative">
              <CoverArt :src="coverOf(t)" size="tile" rounded="rounded-2xl" />
              <span
                class="absolute left-2 top-2 rounded-md bg-black/55 px-1.5 py-0.5 text-[11px] font-bold"
                >{{ idx + 1 }}</span
              >
            </div>
            <div class="mt-2 truncate text-[13px] font-semibold leading-tight">{{ t.title }}</div>
            <div class="truncate text-[11px] text-muted">{{ searchCount(t) || t.artist }}</div>
          </button>
        </div>
      </section>

      <section class="mb-4">
        <div class="mb-2 flex items-center justify-between">
          <h2 class="text-[20px] font-bold leading-none">排行榜</h2>
          <span class="text-[11px] text-muted">搜索越多越靠前</span>
        </div>
        <div class="glass-card rounded-3xl p-1.5">
          <div v-for="(t, i) in library.tracks" :key="t.id + '-' + i" class="flex items-center gap-1">
            <div class="w-10 shrink-0 text-center">
              <div
                class="text-[15px] font-bold tabular-nums leading-none"
                :class="i < 3 ? 'text-accent' : 'text-muted'"
              >
                {{ i + 1 }}
              </div>
              <div v-if="t.searchCount" class="mt-0.5 text-[9px] tabular-nums text-muted/80">
                {{ t.searchCount }}
              </div>
            </div>
            <div class="min-w-0 flex-1">
              <TrackRow :track="t" :queue="library.tracks" />
            </div>
          </div>
          <div v-if="!library.tracks.length" class="py-16 text-center text-muted">暂无曲目</div>
        </div>
      </section>
    </template>
  </div>
</template>
