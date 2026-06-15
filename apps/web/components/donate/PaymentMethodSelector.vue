<script setup lang="ts">
import type { PaymentMethod } from '~/types/api'

const props = defineProps<{
  methods: PaymentMethod[]
  modelValue: string | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const visibleMethods = computed(() => props.methods.filter((method) => method.enabled !== false))

const typeLabel: Record<string, string> = {
  static_qr: '固定二维码',
  redirect_url: '跳转支付',
  mock_qr: '开发模拟',
  wechat_native: '微信 Native',
  alipay_f2f: '支付宝当面付',
  stripe_checkout: 'Stripe Checkout'
}
</script>

<template>
  <div class="grid gap-3 sm:grid-cols-2">
    <button
      v-for="method in visibleMethods"
      :key="method.id"
      type="button"
      class="focus-ring flex items-center gap-4 rounded-[1.4rem] border p-4 text-left transition hover:-translate-y-0.5"
      :class="method.id === modelValue ? 'border-cinnabar bg-cinnabar text-white shadow-[0_16px_40px_rgba(228,83,53,0.24)]' : 'border-ink/10 bg-white/65 text-ink hover:border-cinnabar/40 hover:bg-white/90'"
      @click="emit('update:modelValue', method.id)"
    >
      <img v-if="method.icon_url" :src="method.icon_url" :alt="method.name" class="size-10 rounded-xl object-cover">
      <span v-else class="grid size-10 place-items-center rounded-xl bg-ink/10 font-black">{{ method.name.slice(0, 1) }}</span>
      <span>
        <strong class="block text-sm">{{ method.name }}</strong>
        <span class="text-xs opacity-70">{{ typeLabel[method.type] || method.type }}</span>
      </span>
    </button>
    <p v-if="visibleMethods.length === 0" class="rounded-3xl border border-dashed border-ink/20 bg-white/50 p-5 text-sm text-ink/55">
      暂无可用支付方式，请稍后再试。
    </p>
  </div>
</template>
