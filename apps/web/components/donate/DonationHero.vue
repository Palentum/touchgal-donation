<script setup lang="ts">
import type { PublicSiteConfig } from '~/types/api'

const props = withDefaults(
  defineProps<{
    site: PublicSiteConfig
    raisedCents?: number
  }>(),
  {
    raisedCents: 0
  }
)

const { formatCents } = useMoney()

const progressPercent = computed(() => {
  if (!props.site.show_goal || props.site.goal_cents <= 0) return 0
  return Math.min(100, Math.round((props.raisedCents * 100) / props.site.goal_cents))
})
</script>

<template>
  <section class="relative mx-auto grid max-w-7xl gap-10 px-4 pb-10 pt-8 sm:px-6 lg:grid-cols-[1.05fr_0.95fr] lg:px-8 lg:pb-16 lg:pt-14">
    <div class="relative z-10">
      <div class="mb-6 inline-flex items-center gap-3 rounded-full border border-ink/10 bg-rice/70 px-4 py-2 text-xs font-black uppercase tracking-[0.24em] text-ink/60 shadow-sm backdrop-blur">
        <span class="size-2 rounded-full bg-cinnabar" />
        {{ site.name }}
      </div>
      <h1 class="max-w-4xl font-display text-5xl font-black leading-[0.95] tracking-[-0.05em] text-ink sm:text-7xl lg:text-8xl">
        {{ site.hero_title }}
      </h1>
      <p class="mt-7 max-w-2xl text-lg leading-8 text-ink/68 sm:text-xl">
        {{ site.hero_subtitle }}
      </p>
      <div v-if="site.show_goal && site.goal_cents > 0" class="mt-9 max-w-xl rounded-[2rem] border border-ink/10 bg-white/55 p-4 shadow-inner shadow-white/60 backdrop-blur">
        <div class="flex items-center justify-between gap-4 text-sm font-black text-ink/70">
          <span>本期目标</span>
          <span>{{ formatCents(raisedCents, site.currency) }} / {{ formatCents(site.goal_cents, site.currency) }}</span>
        </div>
        <div class="mt-3 h-3 overflow-hidden rounded-full bg-ink/10">
          <div class="h-full rounded-full bg-gradient-to-r from-jade to-cinnabar transition-all duration-700" :style="{ width: `${progressPercent}%` }" />
        </div>
      </div>
    </div>

    <div class="relative hidden min-h-[24rem] lg:block">
      <div class="absolute inset-x-10 top-8 rotate-[-5deg] rounded-[2.5rem] border border-ink/10 bg-rice/80 p-8 shadow-[0_36px_90px_rgba(22,33,28,0.14)]">
        <p class="font-display text-6xl font-black text-cinnabar">谢谢</p>
        <p class="mt-8 text-sm uppercase tracking-[0.28em] text-ink/45">support ledger</p>
        <div class="mt-5 space-y-3">
          <div class="h-4 w-4/5 rounded-full bg-ink/10" />
          <div class="h-4 w-2/3 rounded-full bg-ink/10" />
          <div class="h-4 w-5/6 rounded-full bg-ink/10" />
        </div>
      </div>
      <div class="absolute bottom-4 right-6 rotate-[7deg] rounded-[2rem] bg-ink p-7 text-rice shadow-[0_30px_70px_rgba(22,33,28,0.25)]">
        <p class="text-xs uppercase tracking-[0.28em] text-rice/50">funds kept in cents</p>
        <p class="mt-4 font-display text-5xl font-black">100%</p>
        <p class="mt-1 text-sm text-rice/70">透明、克制、可追踪。</p>
      </div>
    </div>
  </section>
</template>
