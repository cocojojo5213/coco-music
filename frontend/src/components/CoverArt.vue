<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    src: string
    alt?: string
    /** preset sizes — prefer this over free-form classes for stable layout */
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'tile' | 'xl' | 'hero' | 'now'
    rounded?: string
  }>(),
  {
    size: 'md',
    rounded: 'rounded-xl',
  },
)

const boxClass = computed(() => {
  switch (props.size) {
    case 'xs':
      return 'h-10 w-10'
    case 'sm':
      return 'h-12 w-12'
    case 'md':
      return 'h-14 w-14'
    case 'lg':
      return 'h-24 w-24'
    case 'tile':
      // horizontal rail card
      return 'h-[118px] w-[118px]'
    case 'xl':
      return 'h-40 w-40'
    case 'hero':
      // featured square, not a wide banner
      return 'mx-auto aspect-square w-[min(68vw,240px)]'
    case 'now':
      return 'mx-auto aspect-square w-[min(70vw,280px)]'
    default:
      return 'h-14 w-14'
  }
})

const shadowClass = computed(() =>
  props.size === 'hero' || props.size === 'now' || props.size === 'xl' || props.size === 'lg'
    ? 'shadow-[0_18px_50px_rgba(0,0,0,0.45)]'
    : 'shadow-[0_6px_18px_rgba(0,0,0,0.35)]',
)
</script>

<template>
  <div
    class="relative shrink-0 overflow-hidden bg-panel-2"
    :class="[boxClass, rounded, shadowClass]"
  >
    <img
      v-if="src"
      :src="src"
      :alt="alt || ''"
      class="absolute inset-0 h-full w-full object-cover"
      loading="lazy"
      decoding="async"
      draggable="false"
      @error="($event.target as HTMLImageElement).style.opacity = '0.15'"
    />
    <div
      v-else
      class="absolute inset-0 flex items-center justify-center bg-gradient-to-br from-white/10 to-white/0 text-lg text-white/30"
    >
      ♪
    </div>
    <div class="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/15 to-transparent" />
  </div>
</template>
