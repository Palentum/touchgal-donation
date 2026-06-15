<script setup lang="ts">
import type { DonationTier } from '~/types/api'

const adminApi = useAdminApi()
const toast = useToast()
const { formatCents, centsToDecimalInput, parseDecimalToCents } = useMoney()

const tiers = ref<DonationTier[]>([])
const loading = ref(false)
const saving = ref(false)
const editingId = ref<string | null>(null)
const form = reactive({
  name: '',
  amount: '',
  currency: 'CNY',
  description: '',
  sort_order: 0,
  enabled: true,
  is_default: false
})

function resetForm() {
  editingId.value = null
  form.name = ''
  form.amount = ''
  form.currency = 'CNY'
  form.description = ''
  form.sort_order = 0
  form.enabled = true
  form.is_default = false
}

function editTier(tier: DonationTier) {
  editingId.value = tier.id
  form.name = tier.name
  form.amount = centsToDecimalInput(tier.amount_cents)
  form.currency = tier.currency
  form.description = tier.description
  form.sort_order = tier.sort_order || 0
  form.enabled = tier.enabled !== false
  form.is_default = tier.is_default
}

async function load() {
  loading.value = true
  try {
    tiers.value = await adminApi.listTiers()
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    const amountCents = parseDecimalToCents(form.amount)
    await adminApi.saveTier(
      {
        name: form.name.trim(),
        amount_cents: amountCents,
        currency: form.currency.trim().toUpperCase() || 'CNY',
        description: form.description.trim(),
        sort_order: Number(form.sort_order) || 0,
        enabled: form.enabled,
        is_default: form.is_default
      },
      editingId.value || undefined
    )
    toast.success(editingId.value ? '档位已更新' : '档位已创建')
    resetForm()
    await load()
  } finally {
    saving.value = false
  }
}

async function disableTier(tier: DonationTier) {
  await adminApi.deleteTier(tier.id)
  toast.success('档位已停用')
  await load()
}

onMounted(load)
</script>

<template>
  <section class="grid gap-5 xl:grid-cols-[0.85fr_1.15fr]">
    <UiCard class="p-5" tone="plain">
      <div class="mb-5 flex items-center justify-between gap-3">
        <div>
          <p class="admin-label">tier editor</p>
          <h2 class="mt-2 font-display text-3xl font-black text-ink">{{ editingId ? '编辑档位' : '新建档位' }}</h2>
        </div>
        <UiButton variant="ghost" @click="resetForm">清空</UiButton>
      </div>
      <form class="space-y-4" @submit.prevent="save">
        <label class="space-y-2 block">
          <span class="admin-label">名称</span>
          <input v-model="form.name" class="admin-field" required maxlength="80">
        </label>
        <div class="grid gap-3 sm:grid-cols-2">
          <label class="space-y-2 block">
            <span class="admin-label">金额</span>
            <input v-model="form.amount" class="admin-field" placeholder="29.90" inputmode="decimal" required>
          </label>
          <label class="space-y-2 block">
            <span class="admin-label">币种</span>
            <input v-model="form.currency" class="admin-field" maxlength="3" required>
          </label>
        </div>
        <label class="space-y-2 block">
          <span class="admin-label">简介</span>
          <textarea v-model="form.description" class="admin-field min-h-24 rounded-3xl" maxlength="300" />
        </label>
        <label class="space-y-2 block">
          <span class="admin-label">排序</span>
          <input v-model.number="form.sort_order" class="admin-field" type="number">
        </label>
        <div class="grid gap-3 sm:grid-cols-2">
          <label class="flex items-center gap-3 rounded-2xl bg-ink/5 px-4 py-3 text-sm font-bold text-ink/70">
            <input v-model="form.enabled" type="checkbox" class="size-4 rounded border-ink/20 text-jade focus:ring-jade">
            启用
          </label>
          <label class="flex items-center gap-3 rounded-2xl bg-ink/5 px-4 py-3 text-sm font-bold text-ink/70">
            <input v-model="form.is_default" type="checkbox" class="size-4 rounded border-ink/20 text-jade focus:ring-jade">
            默认档位
          </label>
        </div>
        <UiButton class="w-full" type="submit" :loading="saving">保存档位</UiButton>
      </form>
    </UiCard>

    <UiCard class="p-5" tone="plain">
      <div class="mb-5 flex items-center justify-between gap-3">
        <h2 class="font-display text-3xl font-black text-ink">档位列表</h2>
        <span v-if="loading" class="text-sm text-ink/45">加载中…</span>
      </div>
      <div v-if="tiers.length === 0" class="rounded-2xl border border-dashed border-ink/15 p-8 text-center text-sm text-ink/50">
        暂无档位。
      </div>
      <div v-else class="grid gap-3">
        <article v-for="tier in tiers" :key="tier.id" class="rounded-[1.4rem] border border-ink/10 bg-white/65 p-4">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="font-black text-ink">{{ tier.name }}</h3>
                <span v-if="tier.is_default" class="rounded-full bg-cinnabar px-2 py-1 text-[0.65rem] font-black text-white">默认</span>
                <span :class="tier.enabled === false ? 'bg-ink/10 text-ink/45' : 'bg-jade/10 text-jade-deep'" class="rounded-full px-2 py-1 text-[0.65rem] font-black">
                  {{ tier.enabled === false ? '停用' : '启用' }}
                </span>
              </div>
              <p class="mt-2 font-display text-3xl font-black text-jade-deep">{{ formatCents(tier.amount_cents, tier.currency) }}</p>
              <p class="mt-1 text-sm text-ink/55">{{ tier.description || '无简介' }}</p>
            </div>
            <div class="flex gap-2">
              <UiButton variant="secondary" size="sm" @click="editTier(tier)">编辑</UiButton>
              <UiButton variant="danger" size="sm" @click="disableTier(tier)">停用</UiButton>
            </div>
          </div>
        </article>
      </div>
    </UiCard>
  </section>
</template>
