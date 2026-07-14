<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { activeLyricIndex, parseLyrics, type LyricLine } from '@/lib/lyrics'

const props = defineProps<{
  lrc: string
  currentTime: number
  loading?: boolean
}>()

const scroller = ref<HTMLElement | null>(null)
const lines = computed<LyricLine[]>(() => parseLyrics(props.lrc, true))
const synced = computed(() => lines.value.some((l) => l.time != null))
const active = computed(() =>
  synced.value ? activeLyricIndex(lines.value, props.currentTime) : -1,
)

watch(active, async (idx, prev) => {
  if (idx < 0 || idx === prev) return
  await nextTick()
  const root = scroller.value
  if (!root) return
  const el = root.querySelector<HTMLElement>(`[data-lyric-index="${idx}"]`)
  if (!el) return
  const rootRect = root.getBoundingClientRect()
  const elRect = el.getBoundingClientRect()
  const offset =
    elRect.top - rootRect.top - rootRect.height / 2 + elRect.height / 2 + root.scrollTop
  root.scrollTo({
    top: Math.max(0, offset),
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
  })
})
</script>

<template>
  <div class="flex h-full min-h-0 flex-col">
    <div class="mb-2 flex items-center justify-between">
      <div class="text-[11px] font-semibold uppercase tracking-wider text-muted">Lyrics</div>
      <div class="text-[11px] text-muted">
        <span v-if="loading">加载中</span>
        <span v-else-if="synced">随播放滚动</span>
        <span v-else-if="lines.length">纯文本</span>
      </div>
    </div>

    <div
      ref="scroller"
      class="lyric-scroller min-h-0 flex-1 overflow-y-auto overscroll-contain px-1 py-6"
    >
      <div v-if="loading" class="py-10 text-center text-sm text-muted">歌词加载中…</div>
      <div v-else-if="!lines.length" class="py-10 text-center text-sm text-muted">暂无匹配歌词</div>
      <div v-else class="space-y-4">
        <!-- top spacer so first lines can center -->
        <div class="h-8" aria-hidden="true" />
        <p
          v-for="(line, i) in lines"
          :key="i"
          class="lyric-line px-2 text-center transition-all duration-300"
          :data-lyric-index="i"
          :class="{
            'is-active': i === active,
            'is-passed': synced && active >= 0 && i < active,
            'is-upcoming': synced && active >= 0 && i > active,
            'is-plain': !synced,
          }"
        >
          {{ line.text }}
        </p>
        <div class="h-16" aria-hidden="true" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.lyric-scroller {
  mask-image: linear-gradient(to bottom, transparent, #000 12%, #000 88%, transparent);
  -webkit-mask-image: linear-gradient(to bottom, transparent, #000 12%, #000 88%, transparent);
  scroll-behavior: smooth;
}
.lyric-line {
  font-size: 15px;
  line-height: 1.45;
  color: rgba(255, 255, 255, 0.34);
  font-weight: 500;
}
.lyric-line.is-plain {
  color: rgba(255, 255, 255, 0.78);
  text-align: left;
}
.lyric-line.is-passed {
  color: rgba(255, 255, 255, 0.28);
}
.lyric-line.is-upcoming {
  color: rgba(255, 255, 255, 0.38);
}
.lyric-line.is-active {
  color: #fff;
  font-size: 18px;
  font-weight: 700;
  transform: scale(1.03);
  text-shadow: 0 0 24px rgba(252, 60, 68, 0.25);
}
</style>
