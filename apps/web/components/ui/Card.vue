<script setup lang="ts">
withDefaults(
  defineProps<{
    tone?: 'paper' | 'ink' | 'plain'
  }>(),
  {
    tone: 'paper'
  }
)
</script>

<template>
  <section
    :class="[
      'rounded-[2rem]',
      tone === 'paper' && 'paper-panel',
      tone === 'ink' && 'bg-ink text-rice shadow-[0_24px_80px_rgba(22,33,28,0.24)]',
      tone === 'plain' && 'border border-ink/10 bg-white/65'
    ]"
  >
    <slot />
  </section>
</template>
