<script setup lang="ts">
import type { OverviewResponse } from '~/types/api'

const adminApi = useAdminApi()
const { formatCents } = useMoney()
const toast = useToast()

const today = new Date()
const end = ref(today.toISOString().slice(0, 10))
const startDate = new Date(today)
startDate.setUTCDate(startDate.getUTCDate() - 29)
const start = ref(startDate.toISOString().slice(0, 10))
const loading = ref(false)
const overview = ref<OverviewResponse | null>(null)

const maxDailyAmount = computed(() => Math.max(1, ...((overview.value?.daily || []).map((point) => point.paid_amount_cents))))

async function load() {
  loading.value = true
  try {
    overview.value = await adminApi.overview(start.value, end.value)
  } finally {
    loading.value = false
  }
}

async function reloadWithToast() {
  await load()
  toast.success('总览已刷新')
}

onMounted(load)
</script>

<template>
  <section class="space-y-5">
    <UiCard class="p-5" tone="plain">
      <div class="grid gap-3 sm:grid-cols-[1fr_1fr_auto] sm:items-end">
        <label class="space-y-2">
          <span class="admin-label">开始日期</span>
          <input v-model="start" class="admin-field" type="date">
        </label>
        <label class="space-y-2">
          <span class="admin-label">结束日期</span>
          <input v-model="end" class="admin-field" type="date">
        </label>
        <UiButton :loading="loading" @click="reloadWithToast">刷新</UiButton>
      </div>
    </UiCard>

    <div class="grid gap-4 md:grid-cols-4">
      <UiCard class="p-5" tone="plain">
        <p class="admin-label">已完成捐赠金额</p>
        <p class="mt-3 font-display text-4xl font-black text-jade-deep">{{ formatCents(overview?.total_paid_amount_cents || 0) }}</p>
      </UiCard>
      <UiCard class="p-5" tone="plain">
        <p class="admin-label">总捐赠笔数</p>
        <p class="mt-3 font-display text-4xl font-black text-ink">{{ overview?.total_order_count || 0 }}</p>
      </UiCard>
      <UiCard class="p-5" tone="plain">
        <p class="admin-label">支付完成笔数</p>
        <p class="mt-3 font-display text-4xl font-black text-ink">{{ overview?.paid_order_count || 0 }}</p>
      </UiCard>
      <UiCard class="p-5" tone="plain">
        <p class="admin-label">待确认笔数</p>
        <p class="mt-3 font-display text-4xl font-black text-cinnabar">{{ overview?.pending_order_count || 0 }}</p>
      </UiCard>
    </div>

    <UiCard class="p-5" tone="plain">
      <div class="mb-5 flex items-center justify-between gap-3">
        <div>
          <p class="admin-label">每日趋势</p>
          <h2 class="mt-2 font-display text-3xl font-black text-ink">已完成捐赠金额</h2>
        </div>
        <span v-if="loading" class="text-sm text-ink/45">加载中…</span>
      </div>
      <div v-if="!overview || overview.daily.length === 0" class="rounded-2xl border border-dashed border-ink/15 p-8 text-center text-sm text-ink/50">
        当前日期范围暂无数据。
      </div>
      <div v-else class="flex h-72 items-end gap-2 overflow-x-auto rounded-[1.4rem] bg-ink/5 p-4">
        <div v-for="point in overview.daily" :key="point.date" class="flex min-w-12 flex-1 flex-col items-center justify-end gap-2">
          <div class="w-full rounded-t-xl bg-gradient-to-t from-jade to-cinnabar" :style="{ height: `${Math.max(4, Math.round((point.paid_amount_cents * 100) / maxDailyAmount))}%` }" />
          <span class="text-[0.65rem] font-bold text-ink/45">{{ point.date.slice(5) }}</span>
        </div>
      </div>
    </UiCard>
  </section>
</template>
