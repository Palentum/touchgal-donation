<script setup lang="ts">
import type { DonationTier } from '~/types/api'

const props = defineProps<{
  tiers: DonationTier[]
  modelValue: string | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const { formatCents } = useMoney()
</script>

<template>
  <div class="grid gap-3 sm:grid-cols-3">
    <button
      v-for="tier in tiers"
      :key="tier.id"
      type="button"
      class="focus-ring group relative overflow-hidden rounded-[1.7rem] border p-4 text-left transition hover:-translate-y-1"
      :class="tier.id === modelValue ? 'border-jade bg-jade-deep text-rice shadow-[0_18px_45px_rgba(15,111,92,0.22)]' : 'border-ink/10 bg-white/65 text-ink hover:border-jade/40 hover:bg-white/90'"
      @click="emit('update:modelValue', tier.id)"
    >
      <span v-if="tier.is_default" class="absolute right-3 top-3 rounded-full bg-cinnabar px-2 py-1 text-[0.65rem] font-black uppercase tracking-widest text-white">
        推荐
      </span>
      <span class="block pr-12 text-sm font-black">{{ tier.name }}</span>
      <span class="mt-4 block font-display text-3xl font-black tracking-[-0.04em]">{{ formatCents(tier.amount_cents, tier.currency) }}</span>
      <span class="mt-3 block min-h-10 text-xs leading-5 opacity-70">{{ tier.description }}</span>
    </button>
  </div>
</template>
