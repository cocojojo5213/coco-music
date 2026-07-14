import { defineStore } from 'pinia'
import { api, type Chart } from '@/api/client'
import type { Track } from '@/types'
import {
  downloadTrackClient,
  idbDeleteAudio,
  loadDownloadMeta,
  loadFavorites,
  saveDownloadMeta,
  saveFavorites,
  trackKey,
} from '@/lib/localLibrary'

export const useLibraryStore = defineStore('library', {
  state: () => ({
    tracks: [] as Track[],
    chart: null as Chart | null,
    favorites: loadFavorites() as Track[],
    downloads: loadDownloadMeta() as Track[],
    loading: false,
    error: '' as string,
    status: '' as string,
  }),
  actions: {
    markFlags(list: Track[]) {
      const fav = new Set(this.favorites.map(trackKey))
      const dl = new Set(this.downloads.map(trackKey))
      return list.map((t) => ({
        ...t,
        isFavorite: fav.has(trackKey(t)),
        isDownloaded: dl.has(trackKey(t)),
        favoriteKey: trackKey(t),
      }))
    },
    async loadHot(_refresh = false) {
      this.loading = true
      this.error = ''
      this.status = '加载站友搜索榜…'
      try {
        try {
          const board = await api.charts()
          const chart = (board.items || [])[0]
          if (chart?.items?.length) {
            this.chart = chart
            this.tracks = this.markFlags(chart.items)
            this.status = chart.description || chart.name
            return
          }
        } catch {
          // fall through
        }
        const data = await api.hot(24, _refresh)
        this.tracks = this.markFlags(data.items || [])
        this.chart = {
          id: data.ranking || 'community-search',
          name: data.note || '站友搜索榜',
          description: '按大家实际搜索次数实时排序',
          items: this.tracks,
        }
        this.status = data.note || data.emptyReason || '站友搜索榜'
      } catch (e) {
        this.error = e instanceof Error ? e.message : 'load failed'
        this.tracks = []
      } finally {
        this.loading = false
      }
    },
    async search(q: string) {
      this.loading = true
      this.error = ''
      this.status = ''
      try {
        const data = await api.search(q)
        this.tracks = this.markFlags(data.items || [])
        this.status = data.emptyReason || (this.tracks.length ? '' : '没有找到可完整播放的音乐')
      } catch (e) {
        this.error = e instanceof Error ? e.message : 'search failed'
        this.tracks = []
      } finally {
        this.loading = false
      }
    },
    toggleFavorite(track: Track) {
      const key = trackKey(track)
      const idx = this.favorites.findIndex((t) => trackKey(t) === key)
      if (idx >= 0) {
        this.favorites.splice(idx, 1)
      } else {
        this.favorites.unshift({ ...track, isFavorite: true })
      }
      saveFavorites(this.favorites)
      this.tracks = this.markFlags(this.tracks)
      this.downloads = this.markFlags(this.downloads)
      if (this.chart) {
        this.chart = { ...this.chart, items: this.markFlags(this.chart.items) }
      }
    },
    isFavorite(track: Track) {
      const key = trackKey(track)
      return this.favorites.some((t) => trackKey(t) === key)
    },
    async download(track: Track) {
      const saved = await downloadTrackClient(track)
      const key = trackKey(saved)
      this.downloads = [{ ...saved, isDownloaded: true }, ...this.downloads.filter((t) => trackKey(t) !== key)]
      saveDownloadMeta(this.downloads)
      this.tracks = this.markFlags(this.tracks)
    },
    async removeDownload(track: Track) {
      const key = trackKey(track)
      await idbDeleteAudio(key)
      this.downloads = this.downloads.filter((t) => trackKey(t) !== key)
      saveDownloadMeta(this.downloads)
      this.tracks = this.markFlags(this.tracks)
    },
  },
})
