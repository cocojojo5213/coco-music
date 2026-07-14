import type { Track } from '@/types'
import { normalizeTrack } from '@/api/client'
import { directMediaCandidates } from '@/lib/directMedia'

const FAV_KEY = 'coco-music-favorites-v1'
const DL_META_KEY = 'coco-music-downloads-v1'
const DB_NAME = 'coco-music'
const DB_STORE = 'audio'

export function trackKey(t: Pick<Track, 'id' | 'musicId' | 'title' | 'artist' | 'url'>) {
  const base = [t.musicId || '', t.id || '', t.title || '', t.artist || ''].join('|')
  let h = 2166136261
  for (let i = 0; i < base.length; i++) {
    h ^= base.charCodeAt(i)
    h = Math.imul(h, 16777619)
  }
  return `t-${(h >>> 0).toString(36)}`
}

export function loadFavorites(): Track[] {
  try {
    const raw = JSON.parse(localStorage.getItem(FAV_KEY) || '[]')
    if (!Array.isArray(raw)) return []
    return raw.map((t) => normalizeTrack({ ...t, isFavorite: true })).filter((t) => t.url && t.title)
  } catch {
    return []
  }
}

export function saveFavorites(list: Track[]) {
  const slim = list.slice(0, 200).map((t) => ({
    id: t.id,
    musicId: t.musicId,
    title: t.title,
    artist: t.artist,
    album: t.album,
    duration: t.duration,
    durationMs: t.durationMs,
    coverUrl: t.coverUrl,
    artwork: t.artwork,
    artworkProxyUrl: t.artworkProxyUrl,
    url: t.url,
    directUrl: t.directUrl,
    proxyUrl: t.proxyUrl,
    provider: t.provider,
    providerId: t.providerId,
    providerName: t.providerName,
    source: t.source,
    audioBytes: t.audioBytes,
    contentType: t.contentType,
    favoriteKey: trackKey(t),
    favoriteAddedAt: new Date().toISOString(),
  }))
  localStorage.setItem(FAV_KEY, JSON.stringify(slim))
}

export function loadDownloadMeta(): Track[] {
  try {
    const raw = JSON.parse(localStorage.getItem(DL_META_KEY) || '[]')
    if (!Array.isArray(raw)) return []
    return raw.map((t) => normalizeTrack({ ...t, isDownloaded: true }))
  } catch {
    return []
  }
}

export function saveDownloadMeta(list: Track[]) {
  const slim = list.slice(0, 100).map((t) => ({
    id: t.id,
    musicId: t.musicId,
    title: t.title,
    artist: t.artist,
    album: t.album,
    duration: t.duration,
    coverUrl: t.coverUrl,
    artwork: t.artwork,
    url: t.url,
    directUrl: t.directUrl,
    proxyUrl: t.proxyUrl,
    contentType: t.contentType,
    audioBytes: t.audioBytes,
    provider: t.provider,
    favoriteKey: trackKey(t),
  }))
  localStorage.setItem(DL_META_KEY, JSON.stringify(slim))
}

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, 1)
    req.onupgradeneeded = () => {
      const db = req.result
      if (!db.objectStoreNames.contains(DB_STORE)) {
        db.createObjectStore(DB_STORE)
      }
    }
    req.onsuccess = () => resolve(req.result)
    req.onerror = () => reject(req.error)
  })
}

export async function idbPutAudio(key: string, blob: Blob) {
  const db = await openDB()
  return new Promise<void>((resolve, reject) => {
    const tx = db.transaction(DB_STORE, 'readwrite')
    tx.objectStore(DB_STORE).put(blob, key)
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error)
  })
}

export async function idbGetAudio(key: string): Promise<Blob | null> {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(DB_STORE, 'readonly')
    const req = tx.objectStore(DB_STORE).get(key)
    req.onsuccess = () => resolve((req.result as Blob) || null)
    req.onerror = () => reject(req.error)
  })
}

export async function idbDeleteAudio(key: string) {
  const db = await openDB()
  return new Promise<void>((resolve, reject) => {
    const tx = db.transaction(DB_STORE, 'readwrite')
    tx.objectStore(DB_STORE).delete(key)
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error)
  })
}

/**
 * Client-side download: browser fetches CDN direct URL only.
 * Never uses coco-music /api/proxy (avoids server egress charges).
 */
export async function downloadTrackClient(track: Track): Promise<Track> {
  const key = trackKey(track)
  const candidates = directMediaCandidates(track)
  if (!candidates.length) {
    throw new Error('当前音源没有可用直链，无法绕过服务器下载')
  }

  let blob: Blob | null = null
  let lastErr = 'download failed'
  for (const src of candidates) {
    try {
      // Third-party media hosts; requires CORS for blob download.
      const res = await fetch(src, { mode: 'cors', credentials: 'omit', referrerPolicy: 'no-referrer' })
      if (!res.ok) {
        lastErr = `直链下载失败: ${res.status}`
        continue
      }
      blob = await res.blob()
      if (blob.size > 0) break
      lastErr = 'empty body'
      blob = null
    } catch (e) {
      lastErr = e instanceof Error ? e.message : '直链下载失败（可能被源站 CORS 拦截）'
    }
  }
  if (!blob) throw new Error(lastErr)
  await idbPutAudio(key, blob)

  // also offer a real file download to user device
  const ext = (track.contentType || blob.type || '').includes('flac')
    ? 'flac'
    : (track.contentType || blob.type || '').includes('mp4')
      ? 'm4a'
      : 'mp3'
  const filename = `${sanitize(track.artist)} - ${sanitize(track.title)}.${ext}`
  const objectUrl = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = objectUrl
  a.download = filename
  a.rel = 'noopener'
  document.body.appendChild(a)
  a.click()
  a.remove()
  setTimeout(() => URL.revokeObjectURL(objectUrl), 30_000)

  return normalizeTrack({ ...track, isDownloaded: true, audioBytes: blob.size })
}

export async function localPlaybackUrl(track: Track): Promise<string | null> {
  const key = trackKey(track)
  const blob = await idbGetAudio(key)
  if (!blob) return null
  return URL.createObjectURL(blob)
}

function sanitize(s: string) {
  return String(s || 'track')
    .replace(/[\\/:*?"<>|]/g, '_')
    .slice(0, 80)
}
