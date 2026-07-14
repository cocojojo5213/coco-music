<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import PlayerIcons from './icons/PlayerIcons.vue'

const props = withDefaults(
  defineProps<{
    src: string
    alt?: string
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'tile' | 'xl' | 'hero' | 'now'
    rounded?: string
  }>(),
  {
    size: 'md',
    rounded: 'rounded-xl',
  },
)

const broken = ref(false)
watch(
  () => props.src,
  () => {
    broken.value = false
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
      return 'h-[118px] w-[118px]'
    case 'xl':
      return 'h-40 w-40'
    case 'hero':
      return 'mx-auto aspect-square w-[min(68vw,240px)]'
    case 'now':
      return 'mx-auto aspect-square w-[min(70vw,280px)]'
    default:
      return 'h-14 w-14'
  }
})

const iconSize = computed(() => {
  switch (props.size) {
    case 'xs':
      return 14
    case 'sm':
      return 16
    case 'tile':
    case 'lg':
      return 28
    case 'hero':
    case 'now':
    case 'xl':
      return 42
    default:
      return 20
  }
})

const shadowClass = computed(() =>
  props.size === 'hero' || props.size === 'now' || props.size === 'xl' || props.size === 'lg'
    ? 'shadow-[0_18px_50px_rgba(0,0,0,0.45)]'
    : 'shadow-[0_6px_18px_rgba(0,0,0,0.35)]',
)

const showImg = computed(() => !!props.src && !broken.value)
</script>

<template>
  <div
    class="relative shrink-0 overflow-hidden bg-gradient-to-br from-[#2a2a34] via-panel-2 to-[#121218]"
    :class="[boxClass, rounded, shadowClass]"
  >
    <img
      v-if="showImg"
      :src="src"
      :alt="alt || ''"
      class="absolute inset-0 h-full w-full object-cover transition-opacity duration-300"
      loading="lazy"
      decoding="async"
      draggable="false"
      @error="broken = true"
    />
    <div
      v-else
      class="absolute inset-0 flex items-center justify-center text-white/25"
    >
      <PlayerIcons name="music" :size="iconSize" />
    </div>
    <div class="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/20 via-transparent to-white/5" />
  </div>
</template>
