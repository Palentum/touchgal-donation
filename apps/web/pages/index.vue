<script setup lang="ts">
import AmountTierSelector from '~/components/donate/AmountTierSelector.vue'
import DonationHero from '~/components/donate/DonationHero.vue'
import DonorInfoForm from '~/components/donate/DonorInfoForm.vue'
import PaymentMethodSelector from '~/components/donate/PaymentMethodSelector.vue'
import PaymentModal from '~/components/donate/PaymentModal.vue'
import RecentDonationsTable from '~/components/donate/RecentDonationsTable.vue'
import type { CreateDonationResponse, DonorInfoModel } from '~/types/api'

const runtimeConfig = useRuntimeConfig()
const api = useApi()
const toast = useToast()
const { formatCents } = useMoney()

useHead({
  title: `${runtimeConfig.public.siteName || 'Support Us'} · 捐赠支持`,
  meta: [
    {
      name: 'description',
      content: '选择捐赠档位，填写公开留言，并通过二维码或跳转链接完成支付。'
    }
  ]
})

const { data: publicConfig, pending: configPending, error: configError } = await useAsyncData('public-config', () => api.getPublicConfig())
const { data: recentDonations, refresh: refreshRecent } = await useAsyncData('recent-donations', () => api.getRecentDonations(30), {
  default: () => ({ items: [] })
})

const selectedTierId = ref<string | null>(null)
const selectedPaymentMethodId = ref<string | null>(null)
const donorInfo = ref<DonorInfoModel>({
  nickname: '',
  email: '',
  message: '',
  public_visible: true
})
const submitting = ref(false)
const activeDonation = ref<CreateDonationResponse | null>(null)
const modalOpen = ref(false)

const tiers = computed(() => publicConfig.value?.tiers.filter((tier) => tier.enabled !== false) || [])
const paymentMethods = computed(() => publicConfig.value?.payment_methods.filter((method) => method.enabled !== false) || [])
const selectedTier = computed(() => tiers.value.find((tier) => tier.id === selectedTierId.value) || null)
const recentItems = computed(() => recentDonations.value?.items || [])
const raisedCents = computed(() => recentItems.value.reduce((sum, item) => sum + item.amount_cents, 0))

watchEffect(() => {
  if (!selectedTierId.value && tiers.value.length > 0) {
    const defaultTier = tiers.value.find((tier) => tier.is_default) || tiers.value[0]
    if (defaultTier) selectedTierId.value = defaultTier.id
  }
  if (!selectedPaymentMethodId.value && paymentMethods.value.length > 0) {
    const defaultMethod = paymentMethods.value[0]
    if (defaultMethod) selectedPaymentMethodId.value = defaultMethod.id
  }
})

function clientRequestId() {
  if (import.meta.client && 'crypto' in window && 'randomUUID' in window.crypto) {
    return window.crypto.randomUUID()
  }
  return `donation-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

async function submitDonation() {
  if (!selectedTier.value) {
    toast.error('请选择捐赠档位')
    return
  }
  if (!selectedPaymentMethodId.value) {
    toast.error('请选择支付方式')
    return
  }

  submitting.value = true
  try {
    activeDonation.value = await api.createDonation({
      tier_id: selectedTier.value.id,
      amount_cents: selectedTier.value.amount_cents,
      currency: selectedTier.value.currency,
      payment_method_id: selectedPaymentMethodId.value,
      nickname: donorInfo.value.nickname.trim(),
      email: donorInfo.value.email.trim(),
      message: donorInfo.value.message.trim(),
      public_visible: donorInfo.value.public_visible,
      client_request_id: clientRequestId()
    })
    modalOpen.value = true
    await refreshRecent()
  } finally {
    submitting.value = false
  }
}

async function goToThanks(orderNo: string) {
  modalOpen.value = false
  await navigateTo({ path: '/thanks', query: { order_no: orderNo } })
}

function openRedirect(url: string) {
  if (!url) {
    toast.error('支付链接为空')
    return
  }
  if (import.meta.client) {
    window.location.href = url
  }
}
</script>

<template>
  <main>
    <DonationHero
      v-if="publicConfig"
      :site="publicConfig.site"
      :raised-cents="raisedCents"
    />

    <section class="mx-auto grid max-w-7xl gap-6 px-4 pb-8 sm:px-6 lg:grid-cols-[0.92fr_1.08fr] lg:px-8">
      <UiCard class="p-5 sm:p-7 lg:sticky lg:top-6 lg:self-start" tone="ink">
        <p class="text-xs font-black uppercase tracking-[0.28em] text-rice/45">donation ritual</p>
        <h2 class="mt-4 font-display text-4xl font-black tracking-[-0.04em] text-rice">三步完成支持</h2>
        <p class="mt-4 text-sm leading-7 text-rice/65">
          金额以整数分保存；感谢页只查询后端状态，不会由前端直接改写支付结果。
        </p>
        <ol class="mt-8 space-y-4 text-sm text-rice/70">
          <li class="flex gap-3"><span class="font-display text-2xl text-cinnabar">01</span><span>选择预设档位，默认项会自动选中。</span></li>
          <li class="flex gap-3"><span class="font-display text-2xl text-cinnabar">02</span><span>填写昵称、邮箱和留言；邮箱不会公开。</span></li>
          <li class="flex gap-3"><span class="font-display text-2xl text-cinnabar">03</span><span>扫码或跳转支付，再到感谢页等待确认。</span></li>
        </ol>
      </UiCard>

      <UiCard class="p-4 sm:p-6">
        <div v-if="configPending" class="grid min-h-[28rem] place-items-center text-ink/55">
          正在加载捐赠配置…
        </div>
        <div v-else-if="configError || !publicConfig" class="rounded-[1.5rem] border border-cinnabar/20 bg-cinnabar/5 p-5 text-sm text-cinnabar">
          捐赠配置加载失败，请稍后刷新。
        </div>
        <form v-else class="space-y-8" @submit.prevent="submitDonation">
          <section class="space-y-4">
            <div class="flex items-center gap-3">
              <span class="grid size-9 place-items-center rounded-full bg-ink text-sm font-black text-rice">1</span>
              <h3 class="font-display text-3xl font-black tracking-[-0.04em] text-ink">选择档位</h3>
            </div>
            <AmountTierSelector v-model="selectedTierId" :tiers="tiers" />
          </section>

          <section class="space-y-4">
            <div class="flex items-center gap-3">
              <span class="grid size-9 place-items-center rounded-full bg-ink text-sm font-black text-rice">2</span>
              <h3 class="font-display text-3xl font-black tracking-[-0.04em] text-ink">留下信息</h3>
            </div>
            <DonorInfoForm v-model="donorInfo" />
          </section>

          <section class="space-y-4">
            <div class="flex items-center gap-3">
              <span class="grid size-9 place-items-center rounded-full bg-ink text-sm font-black text-rice">3</span>
              <h3 class="font-display text-3xl font-black tracking-[-0.04em] text-ink">支付方式</h3>
            </div>
            <PaymentMethodSelector v-model="selectedPaymentMethodId" :methods="paymentMethods" />
          </section>

          <div class="flex flex-col gap-3 rounded-[1.8rem] bg-white/65 p-4 sm:flex-row sm:items-center sm:justify-between">
            <p class="text-sm text-ink/60">
              当前金额：
              <strong class="font-display text-2xl text-ink">{{ selectedTier ? formatCents(selectedTier.amount_cents, selectedTier.currency) : '—' }}</strong>
            </p>
            <UiButton size="lg" type="submit" :loading="submitting" :disabled="tiers.length === 0 || paymentMethods.length === 0">
              立即捐赠
            </UiButton>
          </div>
        </form>
      </UiCard>
    </section>

    <RecentDonationsTable
      :items="recentItems"
      :timezone="publicConfig?.site.timezone || 'Asia/Shanghai'"
    />

    <PaymentModal
      :open="modalOpen"
      :donation="activeDonation"
      @close="modalOpen = false"
      @completed="goToThanks"
      @redirect="openRedirect"
    />
  </main>
</template>
