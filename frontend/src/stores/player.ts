import { defineStore } from 'pinia'
import { api, mediaSrc } from '@/api/client'
import type { Track } from '@/types'
import { localPlaybackUrl } from '@/lib/localLibrary'

export const usePlayerStore = defineStore('player', {
  state: () => ({
    queue: [] as Track[],
    index: -1,
    playing: false,
    currentTime: 0,
    duration: 0,
    showNowPlaying: false,
    audio: null as HTMLAudioElement | null,
    loading: false,
    usingProxy: false,
    objectUrl: '' as string,
  }),
  getters: {
    current(state): Track | null {
      if (state.index < 0 || state.index >= state.queue.length) return null
      return state.queue[state.index]
    },
  },
  actions: {
    ensureAudio() {
      if (this.audio) return this.audio
      const audio = new Audio()
      audio.preload = 'metadata'
      audio.addEventListener('timeupdate', () => {
        this.currentTime = audio.currentTime
      })
      audio.addEventListener('durationchange', () => {
        this.duration = Number.isFinite(audio.duration) ? audio.duration : 0
      })
      audio.addEventListener('ended', () => this.next())
      audio.addEventListener('play', () => {
        this.playing = true
      })
      audio.addEventListener('pause', () => {
        this.playing = false
      })
      audio.addEventListener('error', () => {
        // Intentionally no fallback to same-origin /api/proxy — that would
        // pull full audio through coco-music and burn server egress.
        // Direct CDN failure is surfaced as a stalled track instead.
        this.playing = false
      })
      this.audio = audio
      return audio
    },
    revokeObjectUrl() {
      if (this.objectUrl) {
        URL.revokeObjectURL(this.objectUrl)
        this.objectUrl = ''
      }
    },
    async playTracks(tracks: Track[], startId?: string) {
      const playable = tracks.filter((t) => t.playable !== false && (t.url || t.proxyUrl))
      if (!playable.length) return
      this.queue = [...playable]
      const idx = startId ? playable.findIndex((t) => t.id === startId) : 0
      this.index = idx >= 0 ? idx : 0
      await this.loadCurrent(true)
    },
    async loadCurrent(autoplay: boolean) {
      const track = this.current
      if (!track) return
      const audio = this.ensureAudio()
      this.loading = true
      this.currentTime = 0
      this.usingProxy = false
      this.revokeObjectUrl()

      // prefer client-local download blob
      let src = ''
      try {
        const local = await localPlaybackUrl(track)
        if (local) {
          this.objectUrl = local
          src = local
        }
      } catch {
        // ignore
      }
      if (!src) src = mediaSrc(track, false)

      audio.src = src
      try {
        if (autoplay) await audio.play()
        // fire-and-forget play stats to upstream via BFF
        void api.playEvent(track)
      } catch {
        // autoplay blocked
      } finally {
        this.loading = false
      }
    },
    async toggle() {
      const track = this.current
      if (!track) return
      const audio = this.ensureAudio()
      if (!audio.src) {
        await this.loadCurrent(true)
        return
      }
      if (audio.paused) await audio.play()
      else audio.pause()
    },
    async next() {
      if (!this.queue.length) return
      this.index = (this.index + 1) % this.queue.length
      await this.loadCurrent(true)
    },
    async prev() {
      if (!this.queue.length) return
      if (this.currentTime > 3) {
        this.seek(0)
        return
      }
      this.index = (this.index - 1 + this.queue.length) % this.queue.length
      await this.loadCurrent(true)
    },
    seek(t: number) {
      const audio = this.ensureAudio()
      audio.currentTime = t
      this.currentTime = t
    },
  },
})
