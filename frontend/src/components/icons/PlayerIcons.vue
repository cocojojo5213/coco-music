<script setup lang="ts">
withDefaults(
  defineProps<{
    name: 'play' | 'pause' | 'prev' | 'next' | 'heart' | 'heart-fill'
    size?: number
  }>(),
  { size: 24 },
)
</script>

<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="currentColor"
    xmlns="http://www.w3.org/2000/svg"
    aria-hidden="true"
    class="block"
  >
    <!-- Apple-ish filled media glyphs -->
    <path
      v-if="name === 'play'"
      d="M8.2 5.35c0-.9.98-1.45 1.74-.98l10.1 6.15a1.12 1.12 0 0 1 0 1.96l-10.1 6.15c-.76.47-1.74-.08-1.74-.98V5.35Z"
    />
    <path
      v-else-if="name === 'pause'"
      d="M7.4 5.5c-.77 0-1.4.63-1.4 1.4v10.2c0 .77.63 1.4 1.4 1.4h2.3c.77 0 1.4-.63 1.4-1.4V6.9c0-.77-.63-1.4-1.4-1.4H7.4Zm6.9 0c-.77 0-1.4.63-1.4 1.4v10.2c0 .77.63 1.4 1.4 1.4h2.3c.77 0 1.4-.63 1.4-1.4V6.9c0-.77-.63-1.4-1.4-1.4h-2.3Z"
    />
    <path
      v-else-if="name === 'prev'"
      d="M6.2 6.1c0-.6-.5-1.1-1.1-1.1S4 5.5 4 6.1v11.8c0 .6.5 1.1 1.1 1.1s1.1-.5 1.1-1.1V6.1Zm13.5.38c.72-.45.72-1.51 0-1.96-.23-.14-.5-.22-.78-.22-.3 0-.58.09-.82.25L9.2 10.05a1.7 1.7 0 0 0 0 2.9l8.9 5.5c.72.45 1.66-.07 1.66-.95V6.48Z"
    />
    <path
      v-else-if="name === 'next'"
      d="M17.8 6.1c0-.6.5-1.1 1.1-1.1s1.1.5 1.1 1.1v11.8c0 .6-.5 1.1-1.1 1.1s-1.1-.5-1.1-1.1V6.1ZM4.3 6.48c-.72-.45-.72-1.51 0-1.96.23-.14.5-.22.78-.22.3 0 .58.09.82.25l8.9 5.5a1.7 1.7 0 0 1 0 2.9l-8.9 5.5c-.72.45-1.66-.07-1.66-.95V6.48Z"
    />
    <path
      v-else-if="name === 'heart'"
      fill="none"
      stroke="currentColor"
      stroke-width="1.8"
      stroke-linecap="round"
      stroke-linejoin="round"
      d="M12 20.2s-6.8-4.18-8.55-7.86C1.9 9.1 3.2 6.2 6.05 5.55c1.62-.37 3.2.2 4.15 1.45.95-1.25 2.53-1.82 4.15-1.45 2.85.65 4.15 3.55 2.6 6.79C18.8 16.02 12 20.2 12 20.2Z"
    />
    <path
      v-else-if="name === 'heart-fill'"
      d="M12 20.2s-6.8-4.18-8.55-7.86C1.9 9.1 3.2 6.2 6.05 5.55c1.62-.37 3.2.2 4.15 1.45.95-1.25 2.53-1.82 4.15-1.45 2.85.65 4.15 3.55 2.6 6.79C18.8 16.02 12 20.2 12 20.2Z"
    />
  </svg>
</template>
