<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useUiStore } from '@/stores/ui'

const ui = useUiStore()
const { toasts } = storeToRefs(ui)
</script>

<template>
  <div
    class="pointer-events-none fixed inset-x-0 top-0 z-[60] mx-auto flex max-w-lg flex-col items-center gap-2 px-4"
    style="padding-top: max(0.75rem, env(safe-area-inset-top))"
  >
    <TransitionGroup name="toast">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="pointer-events-auto rounded-2xl border border-white/10 bg-[#1c1c24]/92 px-4 py-2.5 text-[13px] text-white shadow-xl backdrop-blur-xl"
        :class="t.kind === 'error' ? 'text-accent' : t.kind === 'ok' ? 'text-white' : 'text-white/90'"
      >
        {{ t.message }}
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.22s ease;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-10px) scale(0.98);
}
</style>
