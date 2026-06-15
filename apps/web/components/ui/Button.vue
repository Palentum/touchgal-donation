<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    type?: 'button' | 'submit' | 'reset'
    variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'quiet'
    size?: 'sm' | 'md' | 'lg'
    disabled?: boolean
    loading?: boolean
  }>(),
  {
    type: 'button',
    variant: 'primary',
    size: 'md',
    disabled: false,
    loading: false
  }
)

const variantClass: Record<string, string> = {
  primary: 'bg-ink text-rice shadow-[0_16px_36px_rgba(22,33,28,0.25)] hover:-translate-y-0.5 hover:bg-jade-deep',
  secondary: 'border border-ink/15 bg-rice/80 text-ink shadow-sm hover:border-jade/50 hover:text-jade-deep',
  ghost: 'text-ink/70 hover:bg-ink/5 hover:text-ink',
  danger: 'bg-cinnabar text-white shadow-[0_16px_36px_rgba(228,83,53,0.24)] hover:-translate-y-0.5 hover:bg-[#c83e28]',
  quiet: 'bg-white/60 text-ink/75 hover:bg-white hover:text-ink'
}

const sizeClass: Record<string, string> = {
  sm: 'px-3 py-2 text-xs',
  md: 'px-4 py-2.5 text-sm',
  lg: 'px-6 py-3.5 text-base'
}

const classes = computed(() => [
  'focus-ring inline-flex items-center justify-center gap-2 rounded-full font-black tracking-wide transition disabled:pointer-events-none disabled:translate-y-0 disabled:opacity-50',
  variantClass[props.variant],
  sizeClass[props.size]
])
</script>

<template>
  <button :type="type" :disabled="disabled || loading" :class="classes">
    <span
      v-if="loading"
      class="size-4 animate-spin rounded-full border-2 border-current border-t-transparent"
      aria-hidden="true"
    />
    <slot />
  </button>
</template>
