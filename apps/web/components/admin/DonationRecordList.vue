<script setup lang="ts">
import AdminExportPanel from '~/components/admin/ExportPanel.vue'
import type { AdminDonationRecord, DonationStatus } from '~/types/api'

const adminApi = useAdminApi()
const toast = useToast()
const { formatCents, formatDateTime, statusLabel } = useMoney()

const today = new Date()
const end = ref(today.toISOString().slice(0, 10))
const startDate = new Date(today)
startDate.setUTCDate(startDate.getUTCDate() - 29)
const start = ref(startDate.toISOString().slice(0, 10))
const status = ref('')
const q = ref('')
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const rows = ref<AdminDonationRecord[]>([])
const total = ref(0)

const statusOptions: Array<{ value: DonationStatus | ''; label: string }> = [
  { value: '', label: '全部状态' },
  { value: 'created', label: '已创建' },
  { value: 'pending', label: '待确认' },
  { value: 'paid', label: '已完成' },
  { value: 'failed', label: '失败' },
  { value: 'cancelled', label: '已取消' },
  { value: 'expired', label: '已过期' },
  { value: 'refunded', label: '已退款' }
]

const canPrev = computed(() => page.value > 1)
const canNext = computed(() => rows.value.length === pageSize && page.value * pageSize < (total.value || Number.MAX_SAFE_INTEGER))

async function load() {
  loading.value = true
  try {
    const response = await adminApi.listDonations({
      start: start.value,
      end: end.value,
      status: status.value,
      q: q.value.trim(),
      page: page.value,
      page_size: pageSize
    })
    rows.value = response.items || []
    total.value = response.total || rows.value.length
  } finally {
    loading.value = false
  }
}

async function applyFilters() {
  page.value = 1
  await load()
}

async function updateStatus(record: AdminDonationRecord, nextStatus: Extract<DonationStatus, 'paid' | 'failed' | 'cancelled'>) {
  await adminApi.updateDonationStatus(record.id, nextStatus)
  toast.success(`订单已标记为${statusLabel(nextStatus)}`)
  await load()
}

async function go(delta: number) {
  page.value += delta
  await load()
}

onMounted(load)
</script>

<template>
  <section class="space-y-5">
    <UiCard class="p-5" tone="plain">
      <div class="grid gap-3 lg:grid-cols-[1fr_1fr_1fr_1.2fr_auto] lg:items-end">
        <label class="space-y-2 block">
          <span class="admin-label">开始日期</span>
          <input v-model="start" class="admin-field" type="date">
        </label>
        <label class="space-y-2 block">
          <span class="admin-label">结束日期</span>
          <input v-model="end" class="admin-field" type="date">
        </label>
        <label class="space-y-2 block">
          <span class="admin-label">状态</span>
          <select v-model="status" class="admin-field">
            <option v-for="option in statusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
          </select>
        </label>
        <label class="space-y-2 block">
          <span class="admin-label">关键词</span>
          <input v-model="q" class="admin-field" placeholder="订单号 / 昵称 / 邮箱">
        </label>
        <UiButton :loading="loading" @click="applyFilters">筛选</UiButton>
      </div>
    </UiCard>

    <AdminExportPanel :start="start" :end="end" :status="status" :q="q" />

    <UiCard class="p-5" tone="plain">
      <div class="mb-5 flex items-center justify-between gap-3">
        <div>
          <p class="admin-label">records</p>
          <h2 class="mt-2 font-display text-3xl font-black text-ink">捐赠记录</h2>
        </div>
        <span class="text-sm text-ink/45">第 {{ page }} 页 · {{ total }} 条</span>
      </div>

      <div v-if="rows.length === 0" class="rounded-2xl border border-dashed border-ink/15 p-8 text-center text-sm text-ink/50">
        当前筛选条件下暂无记录。
      </div>

      <UiTable v-else class="hidden xl:block">
        <thead class="bg-ink/5 text-xs uppercase tracking-[0.18em] text-ink/45">
          <tr>
            <th class="px-4 py-4 font-black">订单号</th>
            <th class="px-4 py-4 font-black">昵称 / 邮箱</th>
            <th class="px-4 py-4 font-black">金额</th>
            <th class="px-4 py-4 font-black">支付方式</th>
            <th class="px-4 py-4 font-black">状态</th>
            <th class="px-4 py-4 font-black">创建 / 支付</th>
            <th class="px-4 py-4 font-black">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-ink/8">
          <tr v-for="record in rows" :key="record.id" class="align-top hover:bg-jade/5">
            <td class="max-w-[13rem] break-all px-4 py-4 font-mono text-xs font-black text-ink">{{ record.order_no }}</td>
            <td class="px-4 py-4 text-sm">
              <p class="font-black text-ink">{{ record.nickname || '匿名捐赠者' }}</p>
              <p class="text-ink/50">{{ record.email || '无邮箱' }}</p>
            </td>
            <td class="px-4 py-4 font-black text-jade-deep">{{ formatCents(record.amount_cents, record.currency) }}</td>
            <td class="px-4 py-4 text-sm text-ink/60">{{ record.payment_method_name || record.payment_method?.name || record.payment_method_id || '—' }}</td>
            <td class="px-4 py-4"><span class="rounded-full bg-ink/5 px-2 py-1 text-xs font-black text-ink/65">{{ statusLabel(record.status) }}</span></td>
            <td class="px-4 py-4 text-xs leading-5 text-ink/55">
              <p>创建 {{ formatDateTime(record.created_at) }}</p>
              <p>支付 {{ formatDateTime(record.paid_at) }}</p>
            </td>
            <td class="px-4 py-4">
              <div v-if="record.status === 'pending' || record.status === 'created'" class="flex flex-wrap gap-2">
                <UiButton size="sm" @click="updateStatus(record, 'paid')">确认</UiButton>
                <UiButton size="sm" variant="secondary" @click="updateStatus(record, 'failed')">失败</UiButton>
                <UiButton size="sm" variant="danger" @click="updateStatus(record, 'cancelled')">取消</UiButton>
              </div>
              <span v-else class="text-xs text-ink/35">无需操作</span>
            </td>
          </tr>
        </tbody>
      </UiTable>

      <div v-if="rows.length > 0" class="grid gap-3 xl:hidden">
        <article v-for="record in rows" :key="record.id" class="rounded-[1.4rem] border border-ink/10 bg-white/65 p-4">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="break-all font-mono text-xs font-black text-ink">{{ record.order_no }}</p>
              <p class="mt-2 font-black text-ink">{{ record.nickname || '匿名捐赠者' }}</p>
              <p class="text-sm text-ink/50">{{ record.email || '无邮箱' }}</p>
            </div>
            <p class="font-display text-2xl font-black text-jade-deep">{{ formatCents(record.amount_cents, record.currency) }}</p>
          </div>
          <div class="mt-3 grid gap-2 text-sm text-ink/55">
            <p>状态：{{ statusLabel(record.status) }}</p>
            <p>支付方式：{{ record.payment_method_name || record.payment_method?.name || '—' }}</p>
            <p>创建：{{ formatDateTime(record.created_at) }}</p>
            <p>支付：{{ formatDateTime(record.paid_at) }}</p>
          </div>
          <div v-if="record.status === 'pending' || record.status === 'created'" class="mt-4 flex flex-wrap gap-2">
            <UiButton size="sm" @click="updateStatus(record, 'paid')">确认</UiButton>
            <UiButton size="sm" variant="secondary" @click="updateStatus(record, 'failed')">失败</UiButton>
            <UiButton size="sm" variant="danger" @click="updateStatus(record, 'cancelled')">取消</UiButton>
          </div>
        </article>
      </div>

      <div class="mt-5 flex items-center justify-end gap-2">
        <UiButton variant="secondary" :disabled="!canPrev" @click="go(-1)">上一页</UiButton>
        <UiButton variant="secondary" :disabled="!canNext" @click="go(1)">下一页</UiButton>
      </div>
    </UiCard>
  </section>
</template>
