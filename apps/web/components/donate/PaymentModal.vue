<script setup lang="ts">
import type { CreateDonationResponse } from '~/types/api'

const props = defineProps<{
  open: boolean
  donation: CreateDonationResponse | null
}>()

const emit = defineEmits<{
  close: []
  completed: [orderNo: string]
  redirect: [url: string]
}>()

const config = useRuntimeConfig()
const toast = useToast()
const { formatCents, formatDateTime } = useMoney()

function apiURL(path: string) {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path

  const apiBase = String(config.public.apiBase || '').replace(/\/$/, '')
  if (!apiBase) return path

  try {
    const base = new URL(apiBase)
    if (path.startsWith('/api/')) {
      return `${base.origin}${path}`
    }
    if (path.startsWith('/')) {
      return `${base.origin}${path}`
    }
    return `${apiBase}/${path}`
  } catch {
    return path
  }
}

const action = computed(() => props.donation?.payment_action || null)
const isMockPayment = computed(() => action.value?.qr_content?.startsWith('mockpay://') || false)
const qrImageUrl = computed(() => {
  if (!props.donation || !action.value) return ''
  if (action.value.qr_image_url) return apiURL(action.value.qr_image_url)
  if (action.value.mode === 'qr_content') {
    return apiURL(`/api/v1/public/donations/${encodeURIComponent(props.donation.order_no)}/qr.png`)
  }
  return ''
})

async function copyOrderNo() {
  if (!props.donation || !import.meta.client) return
  await navigator.clipboard.writeText(props.donation.order_no)
  toast.success('订单号已复制')
}

async function simulateMockPayment() {
  if (!props.donation) return
  await $fetch('/payments/mock_qr/webhook', {
    baseURL: String(config.public.apiBase || ''),
    method: 'POST',
    body: { order_no: props.donation.order_no }
  })
  toast.success('模拟支付已确认')
  emit('completed', props.donation.order_no)
}
</script>

<template>
  <UiModal :open="open" title="完成支付" description="支付状态以后端确认为准，前端只负责展示支付动作并跳转查询。" @close="emit('close')">
    <div v-if="donation && action" class="space-y-6">
      <div class="grid gap-3 rounded-[1.6rem] border border-ink/10 bg-white/60 p-4 sm:grid-cols-2">
        <div>
          <p class="text-xs font-black uppercase tracking-[0.22em] text-ink/45">订单号</p>
          <button type="button" class="mt-1 break-all text-left font-mono text-sm font-black text-jade-deep underline decoration-jade/30 underline-offset-4" @click="copyOrderNo">
            {{ donation.order_no }}
          </button>
        </div>
        <div>
          <p class="text-xs font-black uppercase tracking-[0.22em] text-ink/45">金额</p>
          <p class="mt-1 font-display text-3xl font-black text-ink">{{ formatCents(donation.amount_cents, donation.currency) }}</p>
        </div>
      </div>

      <div v-if="action.mode === 'redirect'" class="rounded-[1.6rem] bg-ink p-5 text-rice">
        <p class="font-display text-3xl font-black">即将跳转</p>
        <p class="mt-3 text-sm leading-6 text-rice/70">{{ action.instructions || '请在第三方支付页面完成支付，完成后回到感谢页查看确认状态。' }}</p>
        <UiButton class="mt-5" variant="primary" @click="emit('redirect', action.redirect_url || '')">
          确认并打开支付页
        </UiButton>
      </div>

      <div v-else class="grid gap-5 md:grid-cols-[17rem_1fr]">
        <div class="rounded-[2rem] border border-ink/10 bg-white p-4 shadow-inner shadow-black/5">
          <img v-if="qrImageUrl" :src="qrImageUrl" alt="支付二维码" class="aspect-square w-full rounded-[1.3rem] object-contain">
          <div v-else class="grid aspect-square place-items-center rounded-[1.3rem] bg-ink/5 p-4 text-center text-sm text-ink/60">
            {{ action.qr_content || '二维码生成中，请稍后重试。' }}
          </div>
        </div>
        <div class="space-y-4">
          <div class="rounded-[1.6rem] border border-dashed border-jade/30 bg-jade/5 p-4 text-sm leading-6 text-ink/70">
            {{ action.instructions || '请扫码支付，并在支付备注中填写订单号。支付完成后点击下方按钮进入感谢页查询状态。' }}
          </div>
          <p v-if="action.expires_at" class="text-xs font-bold text-ink/45">
            支付动作有效期至 {{ formatDateTime(action.expires_at) }}
          </p>
        </div>
      </div>

      <div class="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
        <UiButton variant="secondary" @click="emit('close')">取消</UiButton>
        <UiButton v-if="isMockPayment" variant="secondary" @click="simulateMockPayment">
          模拟支付成功
        </UiButton>
        <UiButton @click="emit('completed', donation.order_no)">我已完成支付</UiButton>
      </div>
    </div>
  </UiModal>
</template>
