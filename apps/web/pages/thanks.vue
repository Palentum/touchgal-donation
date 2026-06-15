<script setup lang="ts">
import type { DonationStatusResponse, DonationStatus } from '~/types/api'

const route = useRoute()
const api = useApi()
const { formatCents, formatDateTime, statusLabel } = useMoney()

useHead({ title: '感谢支持 · Support Us' })

const orderNo = computed(() => String(route.query.order_no || '').trim())
const status = ref<DonationStatusResponse | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const pollStartedAt = ref<number | null>(null)
let timer: ReturnType<typeof window.setInterval> | null = null

const terminalStatuses: Record<DonationStatus, true | undefined> = {
  paid: true,
  failed: true,
  cancelled: true,
  expired: true,
  refunded: true,
  created: undefined,
  pending: undefined
}

const statusTone = computed(() => {
  if (!status.value) return 'pending'
  if (status.value.status === 'paid') return 'paid'
  if (terminalStatuses[status.value.status]) return 'failed'
  return 'pending'
})

async function fetchStatus() {
  if (!orderNo.value) return
  loading.value = true
  errorMessage.value = ''
  try {
    status.value = await api.getDonationStatus(orderNo.value)
    if (status.value.status === 'paid' || terminalStatuses[status.value.status]) {
      stopPolling()
    }
  } catch {
    errorMessage.value = '暂时无法读取订单状态，请稍后重试。'
  } finally {
    loading.value = false
  }
}

function stopPolling() {
  if (timer !== null) {
    window.clearInterval(timer)
    timer = null
  }
}

function startPolling() {
  if (!import.meta.client || !orderNo.value) return
  pollStartedAt.value = Date.now()
  stopPolling()
  timer = window.setInterval(async () => {
    if (pollStartedAt.value && Date.now() - pollStartedAt.value > 120000) {
      stopPolling()
      return
    }
    await fetchStatus()
  }, 3000)
}

onMounted(async () => {
  await fetchStatus()
  if (!status.value || status.value.status === 'created' || status.value.status === 'pending') {
    startPolling()
  }
})

onBeforeUnmount(stopPolling)
</script>

<template>
  <main class="mx-auto grid min-h-screen max-w-4xl place-items-center px-4 py-12 sm:px-6">
    <UiCard class="w-full p-6 sm:p-10">
      <div v-if="!orderNo" class="text-center">
        <p class="font-display text-5xl font-black text-cinnabar">缺少订单号</p>
        <p class="mt-4 text-ink/60">请从支付弹窗进入感谢页。</p>
        <NuxtLink class="ink-link mt-6 inline-block font-black" to="/">返回首页</NuxtLink>
      </div>

      <div v-else class="text-center">
        <p class="text-xs font-black uppercase tracking-[0.28em] text-ink/45">order {{ orderNo }}</p>
        <div
          class="mx-auto mt-6 grid size-24 place-items-center rounded-full text-5xl"
          :class="statusTone === 'paid' ? 'bg-jade text-rice' : statusTone === 'failed' ? 'bg-cinnabar text-white' : 'bg-saffron/35 text-ink'"
        >
          <span v-if="statusTone === 'paid'">✓</span>
          <span v-else-if="statusTone === 'failed'">!</span>
          <span v-else>…</span>
        </div>

        <h1 class="mt-7 font-display text-5xl font-black tracking-[-0.05em] text-ink sm:text-6xl">
          <template v-if="status?.status === 'paid'">感谢你的支持</template>
          <template v-else-if="status && terminalStatuses[status.status]">订单未完成</template>
          <template v-else>支付结果待确认</template>
        </h1>

        <p class="mx-auto mt-5 max-w-2xl text-base leading-7 text-ink/65">
          <template v-if="status?.status === 'paid'">
            后端已确认收到 {{ formatCents(status.amount_cents, status.currency) }}；确认时间 {{ formatDateTime(status.paid_at) }}。
          </template>
          <template v-else-if="status && terminalStatuses[status.status]">
            当前状态为 {{ statusLabel(status.status) }}，你可以返回首页重新发起捐赠。
          </template>
          <template v-else>
            已创建订单，支付结果待确认。我们每 3 秒查询一次后端状态，最多等待 2 分钟；不会从前端直接更改订单状态。
          </template>
        </p>

        <p v-if="loading" class="mt-4 text-sm text-ink/45">正在查询后端状态…</p>
        <p v-if="errorMessage" class="mt-4 text-sm font-bold text-cinnabar">{{ errorMessage }}</p>

        <div class="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
          <UiButton variant="secondary" @click="fetchStatus">重新查询</UiButton>
          <NuxtLink to="/" class="focus-ring inline-flex items-center justify-center rounded-full bg-ink px-5 py-3 text-sm font-black text-rice transition hover:bg-jade-deep">
            返回首页
          </NuxtLink>
        </div>
      </div>
    </UiCard>
  </main>
</template>
