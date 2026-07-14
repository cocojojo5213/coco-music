import type { Track } from '@/types'

/** Prefer proxied artwork when available (more reliable on public demo). */
export function coverOf(track?: Track | null): string {
  if (!track) return ''
  return (
    track.artworkProxyUrl ||
    track.coverUrl ||
    track.artwork ||
    ''
  )
}
