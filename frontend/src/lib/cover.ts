import type { Track } from '@/types'

/** Prefer direct cover art CDN (no host image proxy egress). */
export function coverOf(track?: Track | null): string {
  if (!track) return ''
  // Prefer original artwork/cover CDN; artworkProxyUrl may point at our own hosts.
  return track.coverUrl || track.artwork || track.artworkProxyUrl || ''
}
