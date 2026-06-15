<script setup lang="ts">
import type { DonorInfoModel } from '~/types/api'

const props = defineProps<{
  modelValue: DonorInfoModel
}>()

const emit = defineEmits<{
  'update:modelValue': [value: DonorInfoModel]
}>()

function updateField<K extends keyof DonorInfoModel>(key: K, value: DonorInfoModel[K]) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}

function updateVisibility(event: Event) {
  updateField('public_visible', (event.target as HTMLInputElement).checked)
}
</script>

<template>
  <div class="grid gap-4">
    <UiInput
      :model-value="modelValue.nickname"
      label="昵称"
      placeholder="匿名捐赠者"
      maxlength="60"
      autocomplete="nickname"
      @update:model-value="updateField('nickname', String($event || ''))"
    />
    <UiInput
      :model-value="modelValue.email"
      label="邮箱"
      hint="可空，仅用于联系或凭证；不会在前台展示。"
      placeholder="you@example.com"
      type="email"
      autocomplete="email"
      @update:model-value="updateField('email', String($event || ''))"
    />
    <UiInput
      :model-value="modelValue.message"
      label="留言"
      textarea
      maxlength="300"
      placeholder="写一点鼓励，最多 300 字。"
      @update:model-value="updateField('message', String($event || ''))"
    />
    <label class="flex items-center justify-between gap-4 rounded-3xl border border-ink/10 bg-white/60 px-4 py-3 text-sm text-ink/70">
      <span>
        <strong class="block text-ink">公开显示这笔捐赠</strong>
        <span class="text-xs">前台只展示昵称、留言、金额和时间，不展示邮箱。</span>
      </span>
      <input
        class="size-5 rounded border-ink/20 text-jade focus:ring-jade"
        type="checkbox"
        :checked="modelValue.public_visible"
        @change="updateVisibility"
      >
    </label>
  </div>
</template>
