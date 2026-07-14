import { defineStore } from 'pinia'

export type ToastKind = 'info' | 'ok' | 'error'

export interface Toast {
  id: number
  message: string
  kind: ToastKind
}

let seq = 1

export const useUiStore = defineStore('ui', {
  state: () => ({
    toasts: [] as Toast[],
  }),
  actions: {
    toast(message: string, kind: ToastKind = 'info', ms = 2200) {
      const id = seq++
      this.toasts.push({ id, message, kind })
      window.setTimeout(() => {
        this.toasts = this.toasts.filter((t) => t.id !== id)
      }, ms)
    },
    ok(message: string) {
      this.toast(message, 'ok')
    },
    error(message: string) {
      this.toast(message, 'error', 2800)
    },
  },
})
