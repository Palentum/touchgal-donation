<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    open: boolean
    title?: string
    description?: string
    panelClass?: string
  }>(),
  {
    title: '',
    description: '',
    panelClass: ''
  }
)

const emit = defineEmits<{
  close: []
}>()

watch(
  () => props.open,
  (open) => {
    if (!import.meta.client) return
    document.body.style.overflow = open ? 'hidden' : ''
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  if (import.meta.client) {
    document.body.style.overflow = ''
  }
})
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div v-if="open" class="fixed inset-0 z-40 flex items-end justify-center bg-ink/45 p-3 backdrop-blur-sm sm:items-center" @click.self="emit('close')">
        <Transition
          appear
          enter-active-class="duration-300 ease-out"
          enter-from-class="translate-y-8 scale-95 opacity-0"
          enter-to-class="translate-y-0 scale-100 opacity-100"
          leave-active-class="duration-150 ease-in"
          leave-from-class="translate-y-0 scale-100 opacity-100"
          leave-to-class="translate-y-8 scale-95 opacity-0"
        >
          <section
            class="max-h-[92vh] w-full max-w-2xl overflow-y-auto rounded-[2rem] border border-white/40 bg-rice p-5 shadow-[0_30px_90px_rgba(0,0,0,0.32)] sm:p-7"
            :class="panelClass"
            role="dialog"
            aria-modal="true"
            :aria-label="title"
          >
            <header v-if="title || description" class="mb-5 flex items-start justify-between gap-4">
              <div>
                <p v-if="title" class="font-display text-3xl font-black text-ink">{{ title }}</p>
                <p v-if="description" class="mt-2 text-sm leading-6 text-ink/60">{{ description }}</p>
              </div>
              <button class="focus-ring rounded-full border border-ink/10 bg-white/70 px-3 py-1.5 text-sm font-black text-ink/60 hover:text-ink" type="button" @click="emit('close')">
                关闭
              </button>
            </header>
            <slot />
          </section>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
