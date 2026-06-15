<script setup lang="ts">
import type { RecentDonation } from '~/types/api'

withDefaults(
  defineProps<{
    items: RecentDonation[]
    timezone?: string
  }>(),
  {
    timezone: 'Asia/Shanghai'
  }
)

const { formatCents, formatDateTime } = useMoney()
</script>

<template>
  <section class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
    <div class="mb-6 flex items-end justify-between gap-4">
      <div>
        <p class="text-xs font-black uppercase tracking-[0.25em] text-cinnabar">last 30 days</p>
        <h2 class="mt-2 font-display text-4xl font-black tracking-[-0.04em] text-ink">最近公开捐赠</h2>
      </div>
      <p class="hidden text-sm text-ink/50 sm:block">只展示已确认且选择公开的记录。</p>
    </div>

    <div v-if="items.length === 0" class="rounded-[2rem] border border-dashed border-ink/20 bg-white/45 p-8 text-center text-ink/55">
      最近还没有公开捐赠，成为第一位支持者吧。
    </div>

    <div v-else>
      <UiTable class="hidden md:block">
        <thead class="bg-ink/5 text-xs uppercase tracking-[0.18em] text-ink/45">
          <tr>
            <th class="px-5 py-4 font-black">昵称</th>
            <th class="px-5 py-4 font-black">留言</th>
            <th class="px-5 py-4 font-black">金额</th>
            <th class="px-5 py-4 font-black">时间</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-ink/8">
          <tr v-for="(item, index) in items" :key="`${item.paid_at}-${index}`" class="hover:bg-jade/5">
            <td class="px-5 py-4 font-black text-ink">{{ item.nickname || '匿名捐赠者' }}</td>
            <td class="max-w-xl px-5 py-4 text-ink/65">{{ item.message || '—' }}</td>
            <td class="px-5 py-4 font-black text-jade-deep">{{ formatCents(item.amount_cents, item.currency) }}</td>
            <td class="px-5 py-4 text-ink/55">{{ formatDateTime(item.paid_at, timezone) }}</td>
          </tr>
        </tbody>
      </UiTable>

      <div class="grid gap-3 md:hidden">
        <article v-for="(item, index) in items" :key="`${item.paid_at}-${index}`" class="rounded-[1.5rem] border border-ink/10 bg-white/65 p-4 shadow-sm">
          <div class="flex items-start justify-between gap-4">
            <p class="font-black text-ink">{{ item.nickname || '匿名捐赠者' }}</p>
            <p class="font-display text-2xl font-black text-jade-deep">{{ formatCents(item.amount_cents, item.currency) }}</p>
          </div>
          <p class="mt-3 text-sm leading-6 text-ink/65">{{ item.message || '—' }}</p>
          <p class="mt-3 text-xs font-bold text-ink/45">{{ formatDateTime(item.paid_at, timezone) }}</p>
        </article>
      </div>
    </div>
  </section>
</template>
