<script setup lang="ts">
import type { AdminUser } from '~/types/api'

withDefaults(
  defineProps<{
    admin: AdminUser | null
    title: string
    basePath: string
  }>(),
  {
    admin: null
  }
)

const emit = defineEmits<{
  logout: []
}>()
</script>

<template>
  <header class="flex flex-col gap-4 rounded-[2rem] border border-ink/10 bg-white/65 p-4 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between">
    <div>
      <p class="text-xs font-black uppercase tracking-[0.24em] text-ink/40">{{ basePath }}</p>
      <h1 class="font-display text-4xl font-black tracking-[-0.05em] text-ink">{{ title }}</h1>
    </div>
    <div class="flex items-center gap-3">
      <div class="rounded-2xl bg-ink/5 px-4 py-2 text-sm">
        <span class="block font-black text-ink">{{ admin?.username }}</span>
        <span class="text-xs text-ink/45">{{ admin?.role }}</span>
      </div>
      <UiButton variant="secondary" @click="emit('logout')">退出</UiButton>
    </div>
  </header>
</template>
