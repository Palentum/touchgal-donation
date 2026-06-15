<script setup lang="ts">
import type { PaymentMethod, PaymentMethodConfig, PaymentProviderType } from '~/types/api'

const adminApi = useAdminApi()
const toast = useToast()
const { centsToDecimalInput, parseDecimalToCents } = useMoney()

const methods = ref<PaymentMethod[]>([])
const loading = ref(false)
const saving = ref(false)
const uploadingId = ref<string | null>(null)
const editingId = ref<string | null>(null)

const paymentTypes: Array<{ value: PaymentProviderType; label: string }> = [
  { value: 'static_qr', label: '固定二维码' },
  { value: 'redirect_url', label: '跳转链接' },
  { value: 'mock_qr', label: '开发模拟二维码' },
  { value: 'wechat_native', label: '微信 Native（预留）' },
  { value: 'alipay_f2f', label: '支付宝当面付（预留）' },
  { value: 'stripe_checkout', label: 'Stripe Checkout（预留）' }
]

const form = reactive({
  code: '',
  name: '',
  type: 'mock_qr' as PaymentProviderType,
  provider: 'manual',
  icon_url: '',
  enabled: true,
  sort_order: 0,
  min_amount: '',
  max_amount: '',
  config: {
    qr_image_url: '',
    url_template: '',
    instructions: '',
    require_manual_confirm: true
  } as PaymentMethodConfig
})

const configPreview = computed(() => JSON.stringify(cleanConfig(), null, 2))

function cleanConfig() {
  const config: PaymentMethodConfig = {}
  if (form.type === 'static_qr') {
    config.qr_image_url = String(form.config.qr_image_url || '').trim()
    config.instructions = String(form.config.instructions || '').trim()
    config.require_manual_confirm = Boolean(form.config.require_manual_confirm)
  } else if (form.type === 'redirect_url') {
    config.url_template = String(form.config.url_template || '').trim()
    config.instructions = String(form.config.instructions || '').trim()
  } else {
    config.instructions = String(form.config.instructions || '').trim()
  }
  return config
}

function resetForm() {
  editingId.value = null
  form.code = ''
  form.name = ''
  form.type = 'mock_qr'
  form.provider = 'manual'
  form.icon_url = ''
  form.enabled = true
  form.sort_order = 0
  form.min_amount = ''
  form.max_amount = ''
  form.config = {
    qr_image_url: '',
    url_template: '',
    instructions: '',
    require_manual_confirm: true
  }
}

function editMethod(method: PaymentMethod) {
  const config = method.config_json || method.config || {}
  editingId.value = method.id
  form.code = method.code
  form.name = method.name
  form.type = method.type
  form.provider = method.provider || 'manual'
  form.icon_url = method.icon_url || ''
  form.enabled = method.enabled !== false
  form.sort_order = method.sort_order || 0
  form.min_amount = centsToDecimalInput(method.min_amount_cents || null)
  form.max_amount = centsToDecimalInput(method.max_amount_cents || null)
  form.config = {
    qr_image_url: String(config.qr_image_url || ''),
    url_template: String(config.url_template || ''),
    instructions: String(config.instructions || ''),
    require_manual_confirm: config.require_manual_confirm !== false
  }
}

async function load() {
  loading.value = true
  try {
    methods.value = await adminApi.listPaymentMethods()
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    await adminApi.savePaymentMethod(
      {
        code: form.code.trim(),
        name: form.name.trim(),
        type: form.type,
        provider: form.provider.trim() || 'manual',
        icon_url: form.icon_url.trim(),
        enabled: form.enabled,
        sort_order: Number(form.sort_order) || 0,
        min_amount_cents: form.min_amount ? parseDecimalToCents(form.min_amount) : null,
        max_amount_cents: form.max_amount ? parseDecimalToCents(form.max_amount) : null,
        config_json: cleanConfig()
      },
      editingId.value || undefined
    )
    toast.success(editingId.value ? '支付方式已更新' : '支付方式已创建')
    resetForm()
    await load()
  } finally {
    saving.value = false
  }
}

async function disableMethod(method: PaymentMethod) {
  await adminApi.deletePaymentMethod(method.id)
  toast.success('支付方式已停用')
  await load()
}

async function uploadQr(method: PaymentMethod, event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  uploadingId.value = method.id
  try {
    const response = await adminApi.uploadPaymentQr(method.id, file)
    toast.success('二维码已上传')
    if (editingId.value === method.id) {
      form.config.qr_image_url = response.qr_image_url || response.url || response.icon_url || form.config.qr_image_url
    }
    await load()
  } finally {
    uploadingId.value = null
    input.value = ''
  }
}

onMounted(load)
</script>

<template>
  <section class="grid gap-5 2xl:grid-cols-[0.9fr_1.1fr]">
    <UiCard class="p-5" tone="plain">
      <div class="mb-5 flex items-center justify-between gap-3">
        <div>
          <p class="admin-label">payment editor</p>
          <h2 class="mt-2 font-display text-3xl font-black text-ink">{{ editingId ? '编辑支付方式' : '新建支付方式' }}</h2>
        </div>
        <UiButton variant="ghost" @click="resetForm">清空</UiButton>
      </div>

      <form class="space-y-4" @submit.prevent="save">
        <div class="grid gap-3 sm:grid-cols-2">
          <label class="space-y-2 block">
            <span class="admin-label">编码</span>
            <input v-model="form.code" class="admin-field" placeholder="wechat" required>
          </label>
          <label class="space-y-2 block">
            <span class="admin-label">名称</span>
            <input v-model="form.name" class="admin-field" placeholder="微信支付" required>
          </label>
        </div>
        <div class="grid gap-3 sm:grid-cols-2">
          <label class="space-y-2 block">
            <span class="admin-label">类型</span>
            <select v-model="form.type" class="admin-field">
              <option v-for="type in paymentTypes" :key="type.value" :value="type.value">{{ type.label }}</option>
            </select>
          </label>
          <label class="space-y-2 block">
            <span class="admin-label">Provider</span>
            <input v-model="form.provider" class="admin-field">
          </label>
        </div>
        <label class="space-y-2 block">
          <span class="admin-label">图标 URL</span>
          <input v-model="form.icon_url" class="admin-field" placeholder="https://...">
        </label>
        <div class="grid gap-3 sm:grid-cols-3">
          <label class="space-y-2 block">
            <span class="admin-label">最小金额</span>
            <input v-model="form.min_amount" class="admin-field" inputmode="decimal" placeholder="可空">
          </label>
          <label class="space-y-2 block">
            <span class="admin-label">最大金额</span>
            <input v-model="form.max_amount" class="admin-field" inputmode="decimal" placeholder="可空">
          </label>
          <label class="space-y-2 block">
            <span class="admin-label">排序</span>
            <input v-model.number="form.sort_order" class="admin-field" type="number">
          </label>
        </div>

        <div class="rounded-[1.5rem] border border-ink/10 bg-ink/5 p-4">
          <p class="admin-label mb-3">配置</p>
          <div v-if="form.type === 'static_qr'" class="space-y-3">
            <input v-model="form.config.qr_image_url" class="admin-field" placeholder="二维码图片 URL，亦可在列表上传">
            <textarea v-model="form.config.instructions" class="admin-field min-h-24 rounded-3xl" placeholder="请扫码支付，备注订单号 {order_no}，支付后等待确认。" />
            <label class="flex items-center gap-3 text-sm font-bold text-ink/65">
              <input v-model="form.config.require_manual_confirm" type="checkbox" class="size-4 rounded border-ink/20 text-jade focus:ring-jade">
              需要人工确认
            </label>
          </div>
          <div v-else-if="form.type === 'redirect_url'" class="space-y-3">
            <input v-model="form.config.url_template" class="admin-field" placeholder="https://pay.example.com?amount={amount_decimal}&order={order_no}&return={success_url}">
            <textarea v-model="form.config.instructions" class="admin-field min-h-20 rounded-3xl" placeholder="将跳转至第三方支付页面。" />
          </div>
          <div v-else class="space-y-3">
            <textarea v-model="form.config.instructions" class="admin-field min-h-20 rounded-3xl" placeholder="说明文案" />
            <p v-if="form.type !== 'mock_qr'" class="rounded-2xl bg-saffron/20 p-3 text-xs font-bold leading-5 text-ink/65">
              该渠道首版只保存配置结构。真实商户密钥不得暴露给前端，启用前需完成后端适配器接入。
            </p>
          </div>
          <pre class="mt-4 overflow-x-auto rounded-2xl bg-ink p-4 text-xs text-rice/75">{{ configPreview }}</pre>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
          <label class="flex items-center gap-3 rounded-2xl bg-ink/5 px-4 py-3 text-sm font-bold text-ink/70">
            <input v-model="form.enabled" type="checkbox" class="size-4 rounded border-ink/20 text-jade focus:ring-jade">
            启用
          </label>
          <UiButton type="submit" :loading="saving">保存支付方式</UiButton>
        </div>
      </form>
    </UiCard>

    <UiCard class="p-5" tone="plain">
      <div class="mb-5 flex items-center justify-between gap-3">
        <h2 class="font-display text-3xl font-black text-ink">支付方式列表</h2>
        <span v-if="loading" class="text-sm text-ink/45">加载中…</span>
      </div>
      <div v-if="methods.length === 0" class="rounded-2xl border border-dashed border-ink/15 p-8 text-center text-sm text-ink/50">
        暂无支付方式。
      </div>
      <div v-else class="grid gap-3">
        <article v-for="method in methods" :key="method.id" class="rounded-[1.4rem] border border-ink/10 bg-white/65 p-4">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="font-black text-ink">{{ method.name }}</h3>
                <span class="rounded-full bg-ink/5 px-2 py-1 text-[0.65rem] font-black text-ink/60">{{ method.code }}</span>
                <span class="rounded-full bg-jade/10 px-2 py-1 text-[0.65rem] font-black text-jade-deep">{{ method.type }}</span>
                <span :class="method.enabled === false ? 'bg-ink/10 text-ink/45' : 'bg-cinnabar/10 text-cinnabar'" class="rounded-full px-2 py-1 text-[0.65rem] font-black">
                  {{ method.enabled === false ? '停用' : '启用' }}
                </span>
              </div>
              <p class="mt-2 max-w-xl text-sm leading-6 text-ink/58">
                {{ (method.config_json?.instructions || method.config?.instructions || '无说明') }}
              </p>
              <img
                v-if="method.type === 'static_qr' && (method.config_json?.qr_image_url || method.config?.qr_image_url)"
                :src="String(method.config_json?.qr_image_url || method.config?.qr_image_url)"
                alt="静态二维码"
                class="mt-3 size-28 rounded-2xl border border-ink/10 bg-white object-contain p-2"
              >
            </div>
            <div class="flex flex-wrap gap-2">
              <label v-if="method.type === 'static_qr'" class="focus-ring inline-flex cursor-pointer items-center justify-center rounded-full border border-ink/15 bg-rice/80 px-3 py-2 text-xs font-black text-ink shadow-sm hover:border-jade/50 hover:text-jade-deep">
                {{ uploadingId === method.id ? '上传中…' : '上传二维码' }}
                <input class="sr-only" type="file" accept="image/png,image/jpeg,image/webp" :disabled="uploadingId === method.id" @change="uploadQr(method, $event)">
              </label>
              <UiButton variant="secondary" size="sm" @click="editMethod(method)">编辑</UiButton>
              <UiButton variant="danger" size="sm" @click="disableMethod(method)">停用</UiButton>
            </div>
          </div>
        </article>
      </div>
    </UiCard>
  </section>
</template>
