export interface Track {
  id: string
  musicId?: string
  title: string
  artist: string
  album?: string
  duration?: number
  durationMs?: number
  coverUrl?: string
  artwork?: string
  artworkProxyUrl?: string
  url: string
  /** CDN direct media URL — browser should fetch this, never via coco-music proxy */
  directUrl?: string
  proxyUrl?: string
  streamUrl?: string
  lrc?: string
  playable?: boolean
  preview?: boolean
  provider?: string
  providerId?: string
  providerName?: string
  source?: string
  audioBytes?: number
  contentType?: string
  // ranking meta
  rank?: number
  chartId?: string
  searchCount?: number
  searchTerm?: string
  // local-only
  isFavorite?: boolean
  isDownloaded?: boolean
  favoriteKey?: string
}

export interface MusicListResponse {
  items: Track[]
  cached?: boolean
  emptyReason?: string
  provider?: string
  providers?: string[]
  stale?: boolean
  note?: string
  ranking?: string
}

export interface SearchResult {
  tracks: Track[]
  albums: unknown[]
  artists: unknown[]
}

export interface LyricResponse {
  lrc?: string
  synced?: boolean
}
