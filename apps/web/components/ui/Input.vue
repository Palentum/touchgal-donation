<script lang="ts">
export default {
  inheritAttrs: false
}
</script>

<script setup lang="ts">
const model = defineModel<string | number | null>({ default: '' })

withDefaults(
  defineProps<{
    label?: string
    hint?: string
    error?: string
    textarea?: boolean
  }>(),
  {
    label: '',
    hint: '',
    error: '',
    textarea: false
  }
)
</script>

<template>
  <label class="block space-y-2">
    <span v-if="label" class="text-xs font-black uppercase tracking-[0.22em] text-ink/55">{{ label }}</span>
    <textarea
      v-if="textarea"
      v-model="model"
      class="focus-ring min-h-28 w-full resize-y rounded-3xl border border-ink/10 bg-white/75 px-4 py-3 text-sm text-ink placeholder:text-ink/35 shadow-inner shadow-black/5 outline-none transition focus:border-jade/70"
      v-bind="$attrs"
    />
    <input
      v-else
      v-model="model"
      class="focus-ring w-full rounded-full border border-ink/10 bg-white/75 px-4 py-3 text-sm text-ink placeholder:text-ink/35 shadow-inner shadow-black/5 outline-none transition focus:border-jade/70"
      v-bind="$attrs"
    >
    <span v-if="error" class="block text-xs font-bold text-cinnabar">{{ error }}</span>
    <span v-else-if="hint" class="block text-xs text-ink/45">{{ hint }}</span>
  </label>
</template>
