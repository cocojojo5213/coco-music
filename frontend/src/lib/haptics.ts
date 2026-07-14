/** Best-effort light haptic for mobile (no-op when unsupported). */
export function tick(style: 'light' | 'medium' | 'success' = 'light') {
  try {
    if (typeof navigator !== 'undefined' && 'vibrate' in navigator) {
      if (style === 'success') navigator.vibrate([8, 20, 8])
      else if (style === 'medium') navigator.vibrate(14)
      else navigator.vibrate(8)
    }
  } catch {
    // ignore
  }
}
