<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

definePageMeta({
  layout: false
})

const route = useRoute()
const api = useApi()
const AdminShell = defineAsyncComponent(() => import('~/components/admin/AdminShell.vue'))

const asyncKey = computed(() => `resolve-route:${route.path}`)
const { data: resolved, pending, error } = await useAsyncData(asyncKey.value, () => api.resolveRoute(route.path), {
  watch: [() => route.path]
})

watchEffect(() => {
  if (error.value) {
    showError({ statusCode: 404, statusMessage: '页面不存在' })
  }
})
</script>

<template>
  <main v-if="pending" class="grid min-h-screen place-items-center px-4 text-ink/55">
    正在确认入口…
  </main>
  <component
    :is="AdminShell"
    v-else-if="resolved?.kind === 'admin'"
    :base-path="resolved.base_path"
    :initial-sub-path="resolved.sub_path"
  />
</template>
