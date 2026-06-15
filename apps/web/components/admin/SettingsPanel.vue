<script setup lang="ts">
const emit = defineEmits<{
  adminPathUpdated: [path: string]
}>()

const adminApi = useAdminApi()
const toast = useToast()
const { centsToDecimalInput, parseDecimalToCents } = useMoney()

const loading = ref(false)
const savingSite = ref(false)
const savingPath = ref(false)
const changingPassword = ref(false)
const newEntry = ref('')
const siteForm = reactive({
  name: '',
  hero_title: '',
  hero_subtitle: '',
  currency: 'CNY',
  timezone: 'Asia/Shanghai',
  goal_amount: '',
  show_goal: false
})
const adminPath = ref('')
const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

async function load() {
  loading.value = true
  try {
    const settings = await adminApi.settings()
    siteForm.name = settings.site.name || ''
    siteForm.hero_title = settings.site.hero_title || ''
    siteForm.hero_subtitle = settings.site.hero_subtitle || ''
    siteForm.currency = settings.site.currency || 'CNY'
    siteForm.timezone = settings.site.timezone || 'Asia/Shanghai'
    siteForm.goal_amount = centsToDecimalInput(settings.site.goal_cents || 0)
    siteForm.show_goal = Boolean(settings.site.show_goal)
    adminPath.value = settings.admin.base_path || ''
  } finally {
    loading.value = false
  }
}

async function saveSite() {
  savingSite.value = true
  try {
    await adminApi.updateSiteSettings({
      name: siteForm.name.trim(),
      hero_title: siteForm.hero_title.trim(),
      hero_subtitle: siteForm.hero_subtitle.trim(),
      currency: siteForm.currency.trim().toUpperCase() || 'CNY',
      timezone: siteForm.timezone.trim() || 'Asia/Shanghai',
      goal_cents: siteForm.goal_amount ? parseDecimalToCents(siteForm.goal_amount) : 0,
      show_goal: siteForm.show_goal
    })
    toast.success('站点设置已保存')
  } finally {
    savingSite.value = false
  }
}

async function saveAdminPath() {
  savingPath.value = true
  try {
    const response = await adminApi.updateAdminPath(adminPath.value.trim())
    const nextPath = response.base_path || adminPath.value.trim()
    newEntry.value = nextPath
    emit('adminPathUpdated', nextPath)
    toast.success('后台入口已更新', '请使用新地址访问后台。')
  } finally {
    savingPath.value = false
  }
}

async function changePassword() {
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    toast.error('两次输入的新密码不一致')
    return
  }
  changingPassword.value = true
  try {
    await adminApi.changePassword(passwordForm.old_password, passwordForm.new_password)
    passwordForm.old_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
    toast.success('密码已更新')
  } finally {
    changingPassword.value = false
  }
}

onMounted(load)
</script>

<template>
  <section class="grid gap-5 xl:grid-cols-[1.15fr_0.85fr]">
    <UiCard class="p-5" tone="plain">
      <div class="mb-5 flex items-center justify-between gap-3">
        <div>
          <p class="admin-label">site</p>
          <h2 class="mt-2 font-display text-3xl font-black text-ink">站点设置</h2>
        </div>
        <span v-if="loading" class="text-sm text-ink/45">加载中…</span>
      </div>
      <form class="space-y-4" @submit.prevent="saveSite">
        <div class="grid gap-3 sm:grid-cols-2">
          <label class="space-y-2 block">
            <span class="admin-label">站点名</span>
            <input v-model="siteForm.name" class="admin-field" required>
          </label>
          <label class="space-y-2 block">
            <span class="admin-label">币种</span>
            <input v-model="siteForm.currency" class="admin-field" maxlength="3" required>
          </label>
        </div>
        <label class="space-y-2 block">
          <span class="admin-label">Hero 标题</span>
          <input v-model="siteForm.hero_title" class="admin-field" required>
        </label>
        <label class="space-y-2 block">
          <span class="admin-label">Hero 副标题</span>
          <textarea v-model="siteForm.hero_subtitle" class="admin-field min-h-24 rounded-3xl" required />
        </label>
        <div class="grid gap-3 sm:grid-cols-2">
          <label class="space-y-2 block">
            <span class="admin-label">时区</span>
            <input v-model="siteForm.timezone" class="admin-field" placeholder="Asia/Shanghai">
          </label>
          <label class="space-y-2 block">
            <span class="admin-label">目标金额</span>
            <input v-model="siteForm.goal_amount" class="admin-field" inputmode="decimal" placeholder="0.00">
          </label>
        </div>
        <label class="flex items-center gap-3 rounded-2xl bg-ink/5 px-4 py-3 text-sm font-bold text-ink/70">
          <input v-model="siteForm.show_goal" type="checkbox" class="size-4 rounded border-ink/20 text-jade focus:ring-jade">
          首页展示目标进度
        </label>
        <UiButton type="submit" :loading="savingSite">保存站点设置</UiButton>
      </form>
    </UiCard>

    <div class="space-y-5">
      <UiCard class="p-5" tone="plain">
        <p class="admin-label">admin path</p>
        <h2 class="mt-2 font-display text-3xl font-black text-ink">后台入口路径</h2>
        <form class="mt-5 space-y-4" @submit.prevent="saveAdminPath">
          <label class="space-y-2 block">
            <span class="admin-label">路径</span>
            <input v-model="adminPath" class="admin-field" placeholder="/support-console-9c2e" required>
          </label>
          <p class="text-xs leading-5 text-ink/50">必须以 / 开头，不能是 /、/api、/thanks、/assets、/_nuxt，长度 6 到 80，只允许字母、数字、-、_、/。</p>
          <UiButton type="submit" :loading="savingPath">更新入口</UiButton>
        </form>
        <div v-if="newEntry" class="mt-4 rounded-2xl border border-jade/20 bg-jade/5 p-4 text-sm text-jade-deep">
          新入口：<NuxtLink class="font-black underline underline-offset-4" :to="newEntry">{{ newEntry }}</NuxtLink>
        </div>
      </UiCard>

      <UiCard class="p-5" tone="plain">
        <p class="admin-label">security</p>
        <h2 class="mt-2 font-display text-3xl font-black text-ink">修改密码</h2>
        <form class="mt-5 space-y-4" @submit.prevent="changePassword">
          <input v-model="passwordForm.old_password" class="admin-field" type="password" placeholder="旧密码" autocomplete="current-password" required>
          <input v-model="passwordForm.new_password" class="admin-field" type="password" placeholder="新密码，至少 12 位" autocomplete="new-password" minlength="12" required>
          <input v-model="passwordForm.confirm_password" class="admin-field" type="password" placeholder="再次输入新密码" autocomplete="new-password" minlength="12" required>
          <UiButton type="submit" variant="secondary" :loading="changingPassword">更新密码</UiButton>
        </form>
      </UiCard>
    </div>
  </section>
</template>
