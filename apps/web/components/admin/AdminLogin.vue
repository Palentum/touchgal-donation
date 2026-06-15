<script setup lang="ts">
const emit = defineEmits<{
  loggedIn: []
}>()

const adminApi = useAdminApi()
const username = ref('')
const password = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    await adminApi.login(username.value.trim(), password.value)
    emit('loggedIn')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="grid min-h-screen place-items-center px-4 py-10">
    <UiCard class="w-full max-w-md overflow-hidden p-0">
      <div class="bg-ink p-7 text-rice">
        <p class="text-xs font-black uppercase tracking-[0.28em] text-rice/45">private entrance</p>
        <h1 class="mt-4 font-display text-5xl font-black tracking-[-0.05em]">后台登录</h1>
        <p class="mt-3 text-sm leading-6 text-rice/60">入口路径由后端设置解析；登录后使用 HttpOnly 会话 Cookie 与 CSRF token。</p>
      </div>
      <form class="space-y-4 p-6" @submit.prevent="submit">
        <UiInput v-model="username" label="用户名" autocomplete="username" required />
        <UiInput v-model="password" label="密码" type="password" autocomplete="current-password" required />
        <UiButton class="w-full" type="submit" size="lg" :loading="loading">登录</UiButton>
      </form>
    </UiCard>
  </main>
</template>
