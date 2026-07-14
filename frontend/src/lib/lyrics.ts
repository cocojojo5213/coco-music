export interface LyricLine {
  time: number | null // seconds; null = plain text
  text: string
}

const TIME_TAG = /\[(\d{1,2}):(\d{1,2})(?:[.:](\d{1,3}))?\]/g

function parseTimestamp(m: RegExpMatchArray): number {
  const min = Number(m[1]) || 0
  const sec = Number(m[2]) || 0
  const frac = m[3] || '0'
  // support centiseconds (.xx) and milliseconds (.xxx)
  const ms = Number(frac.padEnd(3, '0').slice(0, 3)) / 1000
  return min * 60 + sec + ms
}

function isMetaLine(text: string): boolean {
  const t = text.trim()
  if (!t) return true
  const lower = t.toLowerCase()
  const meta = [
    '作词',
    '作曲',
    '编曲',
    '制作人',
    '混音',
    '录音',
    '和声',
    'lyricist',
    'composer',
    'arranger',
    'produced by',
  ]
  return meta.some((k) => lower.startsWith(k) || (t.includes(' : ') && lower.includes(k) && t.length < 40))
}

/** Parse LRC or plain lyrics into ordered lines. */
export function parseLyrics(raw: string, preferSynced = true): LyricLine[] {
  const src = String(raw || '').replace(/\r\n?/g, '\n')
  if (!src.trim()) return []

  const timed: LyricLine[] = []
  const plain: LyricLine[] = []

  for (const line of src.split('\n')) {
    const trimmed = line.trim()
    if (!trimmed) continue

    const tags = [...trimmed.matchAll(TIME_TAG)]
    const text = trimmed.replace(TIME_TAG, '').trim()
    if (!text || isMetaLine(text)) continue

    if (preferSynced && tags.length) {
      for (const tag of tags) {
        timed.push({ time: parseTimestamp(tag), text })
      }
    } else if (tags.length) {
      plain.push({ time: null, text })
    } else if (!/^\[[a-z]+:/i.test(trimmed)) {
      plain.push({ time: null, text })
    }
  }

  if (timed.length) {
    return timed.sort((a, b) => (a.time ?? 0) - (b.time ?? 0))
  }
  return plain
}

/** Active line index for current playback time (with small lead-in). */
export function activeLyricIndex(lines: LyricLine[], currentTime: number, lead = 0.18): number {
  if (!lines.length) return -1
  const hasTime = lines.some((l) => l.time != null)
  if (!hasTime) return -1
  const t = currentTime + lead
  let idx = 0
  for (let i = 0; i < lines.length; i++) {
    const lt = lines[i].time
    if (lt != null && lt <= t) idx = i
    else if (lt != null && lt > t) break
  }
  return idx
}
