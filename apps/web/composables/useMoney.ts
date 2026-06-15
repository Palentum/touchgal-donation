const moneyFormatters = new Map<string, Intl.NumberFormat>()
const dateFormatters = new Map<string, Intl.DateTimeFormat>()
const STATUS_LABELS: Record<string, string> = {
  created: '已创建',
  pending: '待确认',
  paid: '已完成',
  failed: '失败',
  cancelled: '已取消',
  expired: '已过期',
  refunded: '已退款'
}

function formatterFor(currency: string) {
  const key = (currency || 'CNY').toUpperCase()
  const cached = moneyFormatters.get(key)
  if (cached) return cached

  const formatter = new Intl.NumberFormat('zh-CN', {
    style: 'currency',
    currency: key,
    currencyDisplay: 'narrowSymbol',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })
  moneyFormatters.set(key, formatter)
  return formatter
}

function dateFormatterFor(timezone: string) {
  const cached = dateFormatters.get(timezone)
  if (cached) return cached

  const formatter = new Intl.DateTimeFormat('zh-CN', {
    timeZone: timezone,
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
  dateFormatters.set(timezone, formatter)
  return formatter
}

export function useMoney() {
  function formatCents(amountCents: number, currency = 'CNY') {
    return formatterFor(currency).format(amountCents / 100)
  }

  function centsToDecimalInput(amountCents?: number | null) {
    if (!amountCents) return ''
    const sign = amountCents < 0 ? '-' : ''
    const absolute = Math.abs(amountCents)
    const yuan = Math.floor(absolute / 100)
    const cents = String(absolute % 100).padStart(2, '0')
    return `${sign}${yuan}.${cents}`
  }

  function parseDecimalToCents(input: string) {
    const normalized = input.trim().replace(/[￥¥,\s]/g, '')
    if (!normalized) return 0
    if (!/^\d+(\.\d{0,2})?$/.test(normalized)) {
      throw new Error('金额必须是最多两位小数的正数')
    }

    const parts = normalized.split('.')
    const wholePart = parts[0] || '0'
    const decimalPart = parts[1] || ''
    const centsText = `${decimalPart}00`.slice(0, 2)
    const cents = BigInt(wholePart) * 100n + BigInt(centsText)
    if (cents > BigInt(Number.MAX_SAFE_INTEGER)) {
      throw new Error('金额超过可安全处理范围')
    }
    return Number(cents)
  }

  function formatDateTime(isoTime?: string | null, timezone = 'Asia/Shanghai') {
    if (!isoTime) return '—'
    return dateFormatterFor(timezone).format(new Date(isoTime))
  }

  function statusLabel(status: string) {
    return STATUS_LABELS[status] || status
  }

  return {
    formatCents,
    centsToDecimalInput,
    parseDecimalToCents,
    formatDateTime,
    statusLabel
  }
}
