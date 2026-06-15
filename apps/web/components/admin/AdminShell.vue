<script setup lang="ts">
import AdminLogin from '~/components/admin/AdminLogin.vue'
import AdminSidebar from '~/components/admin/AdminSidebar.vue'
import AdminTopbar from '~/components/admin/AdminTopbar.vue'
import DashboardOverview from '~/components/admin/DashboardOverview.vue'
import DonationRecordList from '~/components/admin/DonationRecordList.vue'
import PaymentMethodManager from '~/components/admin/PaymentMethodManager.vue'
import SettingsPanel from '~/components/admin/SettingsPanel.vue'
import TierManager from '~/components/admin/TierManager.vue'

const VALID_SECTIONS: Record<string, true> = {
  dashboard: true,
  tiers: true,
  'payment-methods': true,
  donations: true,
  settings: true
}

const SECTION_TITLES: Record<string, string> = {
  dashboard: '总览',
  tiers: '捐赠档位',
  'payment-methods': '支付方式',
  donations: '捐赠记录',
  settings: '设置'
}

const SECTION_COMPONENTS = {
  dashboard: DashboardOverview,
  tiers: TierManager,
  'payment-methods': PaymentMethodManager,
  donations: DonationRecordList,
  settings: SettingsPanel
}

const props = defineProps<{
  basePath: string
  initialSubPath: string
}>()

const adminApi = useAdminApi()
const checking = ref(true)
const overriddenBasePath = ref('')

const basePathForLinks = computed(() => (overriddenBasePath.value || props.basePath).replace(/\/$/, '') || props.basePath)
const normalizedSubPath = computed(() => {
  const path = props.initialSubPath || '/dashboard'
  return path === '/' ? '/dashboard' : path
})
const activeSection = computed(() => {
  const firstSegment = normalizedSubPath.value.replace(/^\//, '').split('/')[0] || 'dashboard'
  return VALID_SECTIONS[firstSegment] ? firstSegment : 'dashboard'
})
const isLoginPath = computed(() => normalizedSubPath.value.replace(/^\//, '').split('/')[0] === 'login')
const sectionTitle = computed(() => SECTION_TITLES[activeSection.value] || '总览')
const activeComponent = computed(() => SECTION_COMPONENTS[activeSection.value as keyof typeof SECTION_COMPONENTS] || DashboardOverview)

function pathFor(section: string) {
  return `${basePathForLinks.value}/${section}`.replace(/\/+/g, '/')
}

async function navigateSection(section: string) {
  await navigateTo(pathFor(section))
}

async function loadMe() {
  checking.value = true
  try {
    await adminApi.me()
  } catch {
    adminApi.admin.value = null
  } finally {
    checking.value = false
  }
}

async function onLoggedIn() {
  await navigateSection(adminApi.admin.value?.must_change_password ? 'settings' : 'dashboard')
}

async function logout() {
  await adminApi.logout()
  await navigateTo(pathFor('login'))
}

function onAdminPathUpdated(path: string) {
  overriddenBasePath.value = path
}

watch(
  () => [adminApi.admin.value, isLoginPath.value, normalizedSubPath.value] as const,
  async ([admin]) => {
    if (!admin) return
    if (isLoginPath.value || props.initialSubPath === '/') {
      await navigateSection('dashboard')
    }
  }
)

onMounted(loadMe)
</script>

<template>
  <div class="min-h-screen bg-[#efe4d1] px-3 py-4 text-ink sm:px-5">
    <div v-if="checking" class="grid min-h-[70vh] place-items-center text-ink/55">
      正在校验后台会话…
    </div>

    <AdminLogin v-else-if="!adminApi.admin.value || isLoginPath" @logged-in="onLoggedIn" />

    <div v-else class="mx-auto grid max-w-[96rem] gap-5 lg:grid-cols-[17rem_1fr]">
      <AdminSidebar :active="activeSection" @navigate="navigateSection" />
      <div class="min-w-0 space-y-5">
        <AdminTopbar :admin="adminApi.admin.value" :title="sectionTitle" :base-path="basePathForLinks" @logout="logout" />
        <div v-if="adminApi.admin.value?.must_change_password" class="rounded-[1.5rem] border border-cinnabar/20 bg-cinnabar/5 p-4 text-sm font-bold text-cinnabar">
          当前管理员必须修改初始密码后再继续使用。请前往设置页更新密码。
        </div>
        <component :is="activeComponent" @admin-path-updated="onAdminPathUpdated" />
      </div>
    </div>
  </div>
</template>
