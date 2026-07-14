import { defineStore } from 'pinia'
import { api, mediaSrc } from '@/api/client'
import type { Track } from '@/types'
import { localPlaybackUrl, trackKey } from '@/lib/localLibrary'
import { canClientDirect } from '@/lib/directMedia'
import { coverOf } from '@/lib/cover'

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
    error: '' as string,
    objectUrl: '' as string,
  }),
  getters: {
    current(state): Track | null {
      if (state.index < 0 || state.index >= state.queue.length) return null
      return state.queue[state.index]
    },
    progress(state): number {
      if (!state.duration) return 0
      return Math.min(1, state.currentTime / state.duration)
    },
  },
  actions: {
    ensureAudio() {
      if (this.audio) return this.audio
      const audio = new Audio()
      audio.preload = 'metadata'
      // Do NOT set crossOrigin — some CDNs allow media element play without CORS
      // headers, but break when crossOrigin=anonymous is forced.
      audio.addEventListener('timeupdate', () => {
        this.currentTime = audio.currentTime
        this.patchMediaPosition()
      })
      audio.addEventListener('durationchange', () => {
        this.duration = Number.isFinite(audio.duration) ? audio.duration : 0
      })
      audio.addEventListener('waiting', () => {
        this.loading = true
      })
      audio.addEventListener('canplay', () => {
        this.loading = false
      })
      audio.addEventListener('ended', () => {
        void this.next()
      })
      audio.addEventListener('play', () => {
        this.playing = true
        this.error = ''
      })
      audio.addEventListener('pause', () => {
        this.playing = false
      })
      audio.addEventListener('error', () => {
        this.playing = false
        this.loading = false
        this.error = '播放失败，已跳过'
        // auto-advance on bad CDN link
        window.setTimeout(() => {
          void this.next()
        }, 350)
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
    updateMediaSession(track: Track) {
      if (!('mediaSession' in navigator)) return
      try {
        const art = coverOf(track)
        navigator.mediaSession.metadata = new MediaMetadata({
          title: track.title || '未知歌曲',
          artist: track.artist || '摇摆熊 · Coco Music',
          album: track.album || 'Coco Music',
          artwork: art
            ? [
                { src: art, sizes: '256x256', type: 'image/jpeg' },
                { src: art, sizes: '512x512', type: 'image/jpeg' },
              ]
            : [],
        })
        navigator.mediaSession.setActionHandler('play', () => {
          void this.toggle()
        })
        navigator.mediaSession.setActionHandler('pause', () => {
          void this.toggle()
        })
        navigator.mediaSession.setActionHandler('previoustrack', () => {
          void this.prev()
        })
        navigator.mediaSession.setActionHandler('nexttrack', () => {
          void this.next()
        })
        navigator.mediaSession.setActionHandler('seekto', (d) => {
          if (typeof d.seekTime === 'number') this.seek(d.seekTime)
        })
      } catch {
        // older browsers
      }
    },
    patchMediaPosition() {
      if (!('mediaSession' in navigator) || !this.duration) return
      try {
        navigator.mediaSession.setPositionState({
          duration: this.duration,
          position: Math.min(this.currentTime, this.duration),
          playbackRate: 1,
        })
      } catch {
        // ignore invalid state
      }
    },
    async playTracks(tracks: Track[], startId?: string) {
      const playable = tracks.filter(
        (t) => t.playable !== false && (canClientDirect(t) || t.isDownloaded || t.url || t.directUrl),
      )
      if (!playable.length) {
        this.error = '没有可播放的直链曲目'
        return
      }
      this.queue = [...playable]
      let idx = 0
      if (startId) {
        const found = playable.findIndex(
          (t) => t.id === startId || t.musicId === startId || trackKey(t) === startId,
        )
        idx = found >= 0 ? found : 0
      }
      this.index = idx
      await this.loadCurrent(true)
    },
    async loadCurrent(autoplay: boolean, depth = 0) {
      const track = this.current
      if (!track) return
      if (depth >= this.queue.length) {
        this.loading = false
        this.error = '队列中没有可播放音频'
        return
      }
      const audio = this.ensureAudio()
      this.loading = true
      this.currentTime = 0
      this.duration = 0
      this.error = ''
      this.revokeObjectUrl()
      this.updateMediaSession(track)

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

      if (!src) {
        this.loading = false
        this.index = (this.index + 1) % this.queue.length
        await this.loadCurrent(autoplay, depth + 1)
        return
      }

      audio.src = src
      try {
        if (autoplay) await audio.play()
        void api.playEvent(track)
      } catch {
        // autoplay blocked — still keep src ready
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
      const next = Math.max(0, t)
      audio.currentTime = next
      this.currentTime = next
      this.patchMediaPosition()
    },
  },
})
