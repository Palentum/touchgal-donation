<script setup lang="ts">
const { toasts, remove } = useToast()

const toneClass: Record<string, string> = {
  success: 'border-jade/30 bg-[#edf8f2] text-jade-deep',
  error: 'border-cinnabar/30 bg-[#fff0ec] text-[#8f2b1e]',
  info: 'border-ink/15 bg-rice text-ink',
  warning: 'border-saffron/50 bg-[#fff7dd] text-[#815819]'
}
</script>

<template>
  <div class="pointer-events-none fixed right-3 top-3 z-50 flex w-[calc(100vw-1.5rem)] max-w-sm flex-col gap-3 sm:right-5 sm:top-5">
    <TransitionGroup
      enter-active-class="duration-200 ease-out"
      enter-from-class="translate-x-6 opacity-0"
      enter-to-class="translate-x-0 opacity-100"
      leave-active-class="duration-150 ease-in"
      leave-from-class="translate-x-0 opacity-100"
      leave-to-class="translate-x-6 opacity-0"
    >
      <article
        v-for="toast in toasts"
        :key="toast.id"
        class="pointer-events-auto rounded-3xl border p-4 shadow-[0_18px_45px_rgba(22,33,28,0.16)] backdrop-blur"
        :class="toneClass[toast.tone]"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-sm font-black">{{ toast.title }}</p>
            <p v-if="toast.message" class="mt-1 text-xs leading-5 opacity-75">{{ toast.message }}</p>
          </div>
          <button type="button" class="rounded-full px-2 text-lg leading-none opacity-60 hover:opacity-100" @click="remove(toast.id)">
            ×
          </button>
        </div>
      </article>
    </TransitionGroup>
  </div>
</template>
