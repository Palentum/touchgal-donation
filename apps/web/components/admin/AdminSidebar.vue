<script setup lang="ts">
const props = defineProps<{
  active: string
}>()

const emit = defineEmits<{
  navigate: [section: string]
}>()

const items = [
  { key: 'dashboard', label: '总览', mark: '概' },
  { key: 'tiers', label: '档位', mark: '档' },
  { key: 'payment-methods', label: '支付方式', mark: '付' },
  { key: 'donations', label: '捐赠记录', mark: '捐' },
  { key: 'settings', label: '设置', mark: '设' }
]
</script>

<template>
  <aside class="rounded-[2rem] border border-ink/10 bg-rice/75 p-3 shadow-[0_20px_70px_rgba(22,33,28,0.10)] backdrop-blur lg:sticky lg:top-5 lg:h-[calc(100vh-2.5rem)]">
    <div class="px-3 py-4">
      <p class="text-xs font-black uppercase tracking-[0.25em] text-cinnabar">admin</p>
      <p class="mt-2 font-display text-3xl font-black tracking-[-0.05em] text-ink">捐赠控制台</p>
    </div>
    <nav class="mt-3 grid gap-2 sm:grid-cols-5 lg:grid-cols-1">
      <button
        v-for="item in items"
        :key="item.key"
        type="button"
        class="focus-ring flex items-center gap-3 rounded-2xl px-3 py-3 text-left text-sm font-black transition"
        :class="props.active === item.key ? 'bg-ink text-rice shadow-lg shadow-ink/15' : 'text-ink/62 hover:bg-white/70 hover:text-ink'"
        @click="emit('navigate', item.key)"
      >
        <span class="grid size-9 place-items-center rounded-xl bg-white/15 text-xs">{{ item.mark }}</span>
        <span>{{ item.label }}</span>
      </button>
    </nav>
  </aside>
</template>
