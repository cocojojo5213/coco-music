/** Helpers so downloads/play prefer third-party CDN and never burn coco-music egress. */

const OWN_HOST_RE = /(52131415\.xyz|coco5213\.(me|io)|localhost|127\.0\.0\.1|^100\.)/i

export function isAbsoluteHttp(url: string): boolean {
  return /^https?:\/\//i.test(url)
}

export function isOwnHost(url: string): boolean {
  try {
    const u = new URL(url, typeof location !== 'undefined' ? location.origin : 'https://music.52131415.xyz')
    return OWN_HOST_RE.test(u.hostname)
  } catch {
    return false
  }
}

/** Extract a browser-direct CDN URL. Never returns coco-music /api proxy paths. */
export function extractDirectMediaUrl(raw?: string | null): string {
  const s = String(raw || '').trim()
  if (!s) return ''

  // relative or absolute /api/proxy?url=CDN
  try {
    const base = typeof location !== 'undefined' ? location.origin : 'https://music.52131415.xyz'
    const u = new URL(s, base)
    if (u.pathname === '/api/proxy') {
      const inner = u.searchParams.get('url') || ''
      if (isAbsoluteHttp(inner) && !isOwnHost(inner)) return inner
      return ''
    }
    // own stream endpoints are not direct CDN
    if (u.pathname === '/api/music/stream' || u.pathname.startsWith('/api/music/stream/')) {
      return ''
    }
    if (isAbsoluteHttp(u.href) && !isOwnHost(u.href)) {
      return u.href
    }
  } catch {
    // ignore
  }
  return ''
}

export function directMediaCandidates(track: {
  directUrl?: string
  url?: string
  proxyUrl?: string
  streamUrl?: string
  clientDirect?: boolean
}): string[] {
  const out: string[] = []
  const seen = new Set<string>()
  for (const raw of [track.directUrl, track.url, track.streamUrl, track.proxyUrl]) {
    const d = extractDirectMediaUrl(raw)
    if (d && !seen.has(d)) {
      seen.add(d)
      out.push(d)
    }
  }
  return out
}

/** Whether this track can be downloaded/played without our server egress. */
export function canClientDirect(track: {
  directUrl?: string
  url?: string
  proxyUrl?: string
  streamUrl?: string
  clientDirect?: boolean
}): boolean {
  if (track.clientDirect && extractDirectMediaUrl(track.directUrl || track.url)) return true
  return directMediaCandidates(track).length > 0
}
