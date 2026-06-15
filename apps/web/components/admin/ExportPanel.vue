<script setup lang="ts">
const props = defineProps<{
  start: string
  end: string
  status: string
  q: string
}>()

const adminApi = useAdminApi()
const toast = useToast()
const exporting = ref(false)

async function exportCsv() {
  exporting.value = true
  try {
    const blob = await adminApi.exportDonations({
      start: props.start,
      end: props.end,
      status: props.status,
      q: props.q
    })
    if (!import.meta.client) return

    const objectUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = `donations_${props.start || 'all'}_${props.end || 'all'}.csv`
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(objectUrl)
    toast.success('CSV 导出已开始')
  } catch {
    toast.error('CSV 导出失败，请稍后重试')
  } finally {
    exporting.value = false
  }
}
</script>

<template>
  <UiCard class="p-4" tone="plain">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <p class="admin-label">export</p>
        <p class="mt-1 text-sm text-ink/55">导出列包含订单号、昵称、金额、币种、状态、时间；默认不包含邮箱。</p>
      </div>
      <UiButton variant="secondary" :loading="exporting" @click="exportCsv">导出 CSV</UiButton>
    </div>
  </UiCard>
</template>
