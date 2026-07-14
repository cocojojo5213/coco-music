<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { activeLyricIndex, parseLyrics, type LyricLine } from '@/lib/lyrics'

const props = defineProps<{
  lrc: string
  currentTime: number
  loading?: boolean
}>()

const emit = defineEmits<{
  seek: [time: number]
}>()

const scroller = ref<HTMLElement | null>(null)
const lines = computed<LyricLine[]>(() => parseLyrics(props.lrc, true))
const synced = computed(() => lines.value.some((l) => l.time != null))
const active = computed(() =>
  synced.value ? activeLyricIndex(lines.value, props.currentTime) : -1,
)

const userHolding = ref(false)
let resumeTimer: number | undefined
let lastUserScroll = 0

function onUserScroll() {
  // ignore programmatic scrolls shortly after our own scrollTo
  if (Date.now() - lastUserScroll < 40) return
  userHolding.value = true
  window.clearTimeout(resumeTimer)
  resumeTimer = window.setTimeout(() => {
    userHolding.value = false
  }, 2600)
}

function markProgrammatic() {
  lastUserScroll = Date.now()
}

async function centerLine(idx: number, smooth = true) {
  const root = scroller.value
  if (!root || idx < 0) return
  await nextTick()
  const el = root.querySelector<HTMLElement>(`[data-lyric-index="${idx}"]`)
  if (!el) return
  const rootRect = root.getBoundingClientRect()
  const elRect = el.getBoundingClientRect()
  const offset =
    elRect.top - rootRect.top - rootRect.height / 2 + elRect.height / 2 + root.scrollTop
  markProgrammatic()
  root.scrollTo({
    top: Math.max(0, offset),
    behavior:
      smooth && !window.matchMedia('(prefers-reduced-motion: reduce)').matches
        ? 'smooth'
        : 'auto',
  })
}

watch(active, (idx, prev) => {
  if (idx < 0 || idx === prev || userHolding.value) return
  void centerLine(idx)
})

watch(
  () => props.lrc,
  () => {
    userHolding.value = false
    void centerLine(Math.max(0, active.value), false)
  },
)

function onLineClick(line: LyricLine) {
  if (line.time == null) return
  emit('seek', line.time)
  userHolding.value = false
}

onBeforeUnmount(() => {
  window.clearTimeout(resumeTimer)
})
</script>

<template>
  <div class="flex h-full min-h-0 flex-col">
    <div class="mb-2 flex items-center justify-between px-1">
      <div class="text-[11px] font-semibold uppercase tracking-wider text-muted">Lyrics</div>
      <div class="text-[11px] text-muted">
        <span v-if="loading">加载中</span>
        <span v-else-if="userHolding">滑动浏览中</span>
        <span v-else-if="synced">点歌词可跳转</span>
        <span v-else-if="lines.length">纯文本</span>
      </div>
    </div>

    <div
      ref="scroller"
      class="lyric-scroller min-h-0 flex-1 overflow-y-auto overscroll-contain px-1 py-6"
      @scroll.passive="onUserScroll"
      @touchstart.passive="onUserScroll"
    >
      <div v-if="loading" class="py-10 text-center text-sm text-muted">歌词加载中…</div>
      <div v-else-if="!lines.length" class="py-10 text-center text-sm text-muted">暂无匹配歌词</div>
      <div v-else class="space-y-4">
        <div class="h-10" aria-hidden="true" />
        <button
          v-for="(line, i) in lines"
          :key="i"
          type="button"
          class="lyric-line block w-full px-2 text-center transition-all duration-300"
          :data-lyric-index="i"
          :class="{
            'is-active': i === active,
            'is-passed': synced && active >= 0 && i < active,
            'is-upcoming': synced && active >= 0 && i > active,
            'is-plain': !synced,
            'is-seekable': synced && line.time != null,
          }"
          @click="onLineClick(line)"
        >
          {{ line.text }}
        </button>
        <div class="h-16" aria-hidden="true" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.lyric-scroller {
  mask-image: linear-gradient(to bottom, transparent, #000 10%, #000 90%, transparent);
  -webkit-mask-image: linear-gradient(to bottom, transparent, #000 10%, #000 90%, transparent);
  -webkit-overflow-scrolling: touch;
}
.lyric-line {
  font-size: 15px;
  line-height: 1.5;
  color: rgba(255, 255, 255, 0.34);
  font-weight: 500;
  background: transparent;
  border: 0;
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
.lyric-line.is-seekable:active {
  opacity: 0.75;
}
</style>
