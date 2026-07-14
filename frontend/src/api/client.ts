import type { LyricResponse, MusicListResponse, SearchResult, Track } from '@/types'
import { directMediaCandidates, extractDirectMediaUrl } from '@/lib/directMedia'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { Accept: 'application/json', ...(init?.headers || {}) },
    ...init,
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || res.statusText)
  }
  return res.json() as Promise<T>
}

export function normalizeTrack(raw: Track): Track {
  const cover = raw.coverUrl || raw.artwork || raw.artworkProxyUrl || ''
  const direct =
    extractDirectMediaUrl(raw.directUrl) ||
    extractDirectMediaUrl(raw.url) ||
    extractDirectMediaUrl(raw.streamUrl) ||
    extractDirectMediaUrl(raw.proxyUrl) ||
    ''
  // Prefer CDN direct; never normalize toward same-origin /api/proxy.
  const url = direct || raw.url || raw.streamUrl || ''
  const duration =
    raw.duration ||
    (raw.durationMs ? Math.round(raw.durationMs / 1000) : 0)
  return {
    ...raw,
    title: raw.title || '未知歌曲',
    artist: raw.artist || '未知歌手',
    coverUrl: cover,
    artwork: raw.artwork || cover,
    directUrl: direct || raw.directUrl,
    url,
    streamUrl: url,
    duration,
    playable: raw.playable !== false && !!(url || direct),
  }
}

export interface Chart {
  id: string
  name: string
  description?: string
  updatedAt?: string
  items: Track[]
}

export const api = {
  health: () => request<{ status: string; features?: Record<string, boolean> }>('/api/health'),
  hot: (count = 24, refresh = false) =>
    request<MusicListResponse>(
      `/api/music/hot?count=${count}${refresh ? '&refresh=1' : ''}`,
    ).then((d) => ({ ...d, items: (d.items || []).map(normalizeTrack) })),
  charts: () =>
    request<{ items: Chart[] }>('/api/music/charts').then((d) => ({
      items: (d.items || []).map((c) => ({
        ...c,
        items: (c.items || []).map(normalizeTrack),
      })),
    })),
  chart: (id: string) =>
    request<Chart>(`/api/music/charts/${encodeURIComponent(id)}`).then((c) => ({
      ...c,
      items: (c.items || []).map(normalizeTrack),
    })),
  search: (q: string, refresh = false) =>
    request<MusicListResponse>(
      `/api/music/search?q=${encodeURIComponent(q)}${refresh ? '&refresh=1' : ''}`,
    ).then((d) => ({ ...d, items: (d.items || []).map(normalizeTrack) })),
  legacySearch: async (q: string): Promise<SearchResult> => {
    const d = await api.search(q)
    return { tracks: d.items, albums: [], artists: [] }
  },
  lyrics: (track: Track) =>
    request<LyricResponse>('/api/music/lyrics', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        track: {
          id: track.id,
          musicId: track.musicId || '',
          providerId: track.providerId || '',
          provider: track.provider || '',
          title: track.title,
          artist: track.artist,
          url: track.url,
          lrc: track.lrc || '',
        },
      }),
    }),
  playEvent: (track: Track) =>
    fetch('/api/music/play', {
      method: 'POST',
      headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
      body: JSON.stringify({
        track: {
          id: track.id,
          musicId: track.musicId || '',
          title: track.title,
          artist: track.artist,
          url: track.url,
          artwork: track.artwork || track.coverUrl || '',
          providerId: track.providerId || '',
          providerName: track.providerName || '',
          provider: track.provider || '',
        },
      }),
      keepalive: true,
    }).catch(() => undefined),
}

export function formatTime(sec: number) {
  if (!Number.isFinite(sec) || sec < 0) return '0:00'
  const m = Math.floor(sec / 60)
  const s = Math.floor(sec % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

export function formatBytes(n?: number) {
  if (!n || n <= 0) return ''
  if (n >= 1024 * 1024) return `${(n / 1024 / 1024).toFixed(1)} MB`
  return `${Math.round(n / 1024)} KB`
}

export function mediaSrc(track: Track, useProxy = false) {
  // Default: CDN direct only (zero host egress). Never return stream/proxy on our hosts.
  const direct = directMediaCandidates(track)[0]
  if (direct) return direct
  if (useProxy) {
    // even explicit proxy fallback is disabled for egress safety
    return ''
  }
  return ''
}
