<script setup lang="ts">
import { ref } from 'vue'
import { api } from '@/api/client'
import type { Track } from '@/types'
import TrackRow from '@/components/TrackRow.vue'
import { useLibraryStore } from '@/stores/library'

const q = ref('')
const loading = ref(false)
const tracks = ref<Track[]>([])
const status = ref('')
const library = useLibraryStore()
let timer: number | undefined

function onInput() {
  window.clearTimeout(timer)
  timer = window.setTimeout(() => void runSearch(), 280)
}

async function runSearch() {
  const query = q.value.trim()
  if (!query) {
    tracks.value = []
    status.value = ''
    return
  }
  loading.value = true
  status.value = '搜索中…'
  try {
    const data = await api.search(query)
    tracks.value = library.markFlags(data.items || [])
    status.value = data.emptyReason || (tracks.value.length ? '' : '没有结果')
  } catch (e) {
    status.value = e instanceof Error ? e.message : '搜索失败'
    tracks.value = []
  } finally {
    loading.value = false
  }
}

async function submit() {
  window.clearTimeout(timer)
  await runSearch()
}
</script>

<template>
  <div class="safe-top px-4">
    <header class="mb-4 pt-2">
      <h1 class="mb-3 text-3xl font-bold">Search</h1>
      <p class="mb-3 -mt-2 text-sm text-muted">摇摆熊</p>
      <form class="rounded-2xl bg-white/10 px-4 py-3" @submit.prevent="submit">
        <input
          v-model="q"
          class="w-full bg-transparent text-base outline-none placeholder:text-muted"
          placeholder="艺人、歌曲或专辑"
          type="search"
          enterkeyhint="search"
          @input="onInput"
        />
      </form>
    </header>

    <div v-if="loading" class="py-16 text-center text-muted">搜索完整音频…</div>
    <div v-else-if="!q.trim()" class="py-16 text-center text-muted">试试 稻香 / 起风了 / 周杰伦</div>
    <template v-else>
      <p v-if="status" class="mb-2 text-xs text-muted">{{ status }}</p>
      <div class="glass-card rounded-3xl p-2">
        <TrackRow v-for="t in tracks" :key="t.id" :track="t" :queue="tracks" />
        <div v-if="!tracks.length && !loading" class="py-16 text-center text-muted">没有结果</div>
      </div>
    </template>
  </div>
</template>
