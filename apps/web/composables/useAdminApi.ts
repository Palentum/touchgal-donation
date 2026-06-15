import type {
  AdminDonationRecord,
  AdminPathResponse,
  AdminUser,
  ApiEnvelope,
  ApiErrorEnvelope,
  DonationStatus,
  DonationTier,
  ListResponse,
  LoginResponse,
  OverviewResponse,
  PaymentMethod,
  PaymentMethodConfig,
  PublicSiteConfig,
  SettingsResponse,
  UploadResponse
} from '~/types/api'

type AdminRequestOptions = {
  method?: 'GET' | 'POST' | 'PATCH' | 'DELETE'
  body?: Record<string, any> | BodyInit | null
  query?: Record<string, string | number | boolean | undefined>
  headers?: HeadersInit
  toastErrors?: boolean
}

type DonationFilters = {
  start?: string
  end?: string
  status?: string
  q?: string
  page?: number
  page_size?: number
}

const UNSAFE_METHOD: Record<string, true> = {
  POST: true,
  PATCH: true,
  DELETE: true
}

function unwrapEnvelope<T>(payload: ApiEnvelope<T> | T): T {
  if (payload && typeof payload === 'object' && 'data' in payload) {
    return (payload as ApiEnvelope<T>).data
  }
  return payload as T
}

function errorMessage(error: unknown) {
  const possibleData = (error as { data?: ApiErrorEnvelope })?.data
  return possibleData?.error?.message || (error as Error)?.message || '后台请求失败'
}

function itemsFrom<T>(payload: ListResponse<T> | T[]): T[] {
  return Array.isArray(payload) ? payload : payload.items || []
}

function normalizeSettings(payload: SettingsResponse | { site?: PublicSiteConfig; admin?: { base_path?: string } }) {
  return {
    site: payload.site as PublicSiteConfig,
    admin: {
      base_path: payload.admin?.base_path || ''
    }
  }
}

function appendQuery(params: URLSearchParams, key: string, value?: string | number) {
  if (value !== undefined && value !== '') {
    params.set(key, String(value))
  }
}

export function useAdminApi() {
  const config = useRuntimeConfig()
  const toast = useToast()
  const baseURL = import.meta.server && config.apiInternalBase ? String(config.apiInternalBase) : String(config.public.apiBase || '')
  const admin = useState<AdminUser | null>('admin.user', () => null)
  const csrfToken = useState<string | null>('admin.csrf', () => null)

  if (import.meta.client && csrfToken.value === null) {
    csrfToken.value = window.localStorage.getItem('admin.csrf')
  }

  function setCsrfToken(token: string | null) {
    csrfToken.value = token
    if (!import.meta.client) return

    if (token) {
      window.localStorage.setItem('admin.csrf', token)
    } else {
      window.localStorage.removeItem('admin.csrf')
    }
  }

  async function request<T>(path: string, options: AdminRequestOptions = {}) {
    const { toastErrors = true, headers, method = 'GET', ...fetchOptions } = options
    const requestHeaders = new Headers(headers)

    if (UNSAFE_METHOD[method] && csrfToken.value) {
      requestHeaders.set('X-CSRF-Token', csrfToken.value)
    }

    try {
      const response = await $fetch<ApiEnvelope<T> | T>(path, {
        baseURL,
        credentials: 'include',
        method,
        headers: requestHeaders,
        ...fetchOptions
      })
      return unwrapEnvelope<T>(response)
    } catch (error) {
      const status = (error as { status?: number; statusCode?: number })?.status || (error as { statusCode?: number })?.statusCode
      if (status === 401) {
        admin.value = null
      }
      if (toastErrors) {
        toast.error(errorMessage(error))
      }
      throw error
    }
  }

  async function login(username: string, password: string) {
    const response = await request<LoginResponse>('/admin/auth/login', {
      method: 'POST',
      body: { username, password }
    })
    admin.value = response.admin
    setCsrfToken(response.csrf_token)
    return response.admin
  }

  async function logout() {
    try {
      await request<void>('/admin/auth/logout', { method: 'POST', toastErrors: false })
    } finally {
      admin.value = null
      setCsrfToken(null)
    }
  }

  async function me() {
    const response = await request<AdminUser | { admin: AdminUser; csrf_token?: string }>('/admin/auth/me', {
      toastErrors: false
    })
    const current = 'admin' in response ? response.admin : response
    admin.value = current
    if ('csrf_token' in response && response.csrf_token) {
      setCsrfToken(response.csrf_token)
    }
    return current
  }

  async function changePassword(oldPassword: string, newPassword: string) {
    await request<void>('/admin/auth/change-password', {
      method: 'POST',
      body: { old_password: oldPassword, new_password: newPassword }
    })
    if (admin.value) {
      admin.value = { ...admin.value, must_change_password: false }
    }
  }

  async function listTiers() {
    return itemsFrom(await request<ListResponse<DonationTier> | DonationTier[]>('/admin/tiers'))
  }

  async function saveTier(payload: Partial<DonationTier>, id?: string) {
    if (id) {
      return request<DonationTier>(`/admin/tiers/${encodeURIComponent(id)}`, { method: 'PATCH', body: payload })
    }
    return request<DonationTier>('/admin/tiers', { method: 'POST', body: payload })
  }

  async function deleteTier(id: string) {
    await request<void>(`/admin/tiers/${encodeURIComponent(id)}`, { method: 'DELETE' })
  }

  async function listPaymentMethods() {
    return itemsFrom(await request<ListResponse<PaymentMethod> | PaymentMethod[]>('/admin/payment-methods'))
  }

  async function savePaymentMethod(payload: Partial<PaymentMethod> & { config_json?: PaymentMethodConfig }, id?: string) {
    if (id) {
      return request<PaymentMethod>(`/admin/payment-methods/${encodeURIComponent(id)}`, {
        method: 'PATCH',
        body: payload
      })
    }
    return request<PaymentMethod>('/admin/payment-methods', { method: 'POST', body: payload })
  }

  async function deletePaymentMethod(id: string) {
    await request<void>(`/admin/payment-methods/${encodeURIComponent(id)}`, { method: 'DELETE' })
  }

  async function uploadPaymentQr(id: string, file: File) {
    const form = new FormData()
    form.set('file', file)
    return request<UploadResponse>(`/admin/payment-methods/${encodeURIComponent(id)}/upload-qr`, {
      method: 'POST',
      body: form
    })
  }

  async function listDonations(filters: DonationFilters) {
    return request<ListResponse<AdminDonationRecord>>('/admin/donations', {
      query: {
        start: filters.start,
        end: filters.end,
        status: filters.status,
        q: filters.q,
        page: filters.page,
        page_size: filters.page_size
      }
    })
  }

  async function updateDonationStatus(id: string, status: Extract<DonationStatus, 'paid' | 'failed' | 'cancelled'>) {
    return request<AdminDonationRecord>(`/admin/donations/${encodeURIComponent(id)}/status`, {
      method: 'PATCH',
      body: { status }
    })
  }

  async function exportDonations(filters: DonationFilters) {
    const params = new URLSearchParams()
    appendQuery(params, 'start', filters.start)
    appendQuery(params, 'end', filters.end)
    appendQuery(params, 'status', filters.status)
    appendQuery(params, 'q', filters.q)
    params.set('format', 'csv')

    return $fetch<Blob>(`/admin/donations/export?${params.toString()}`, {
      baseURL,
      credentials: 'include',
      responseType: 'blob'
    })
  }

  async function overview(start: string, end: string) {
    return request<OverviewResponse>('/admin/overview', { query: { start, end } })
  }

  async function settings() {
    return normalizeSettings(await request<SettingsResponse>('/admin/settings'))
  }

  async function updateSiteSettings(site: Partial<PublicSiteConfig>) {
    return request<SettingsResponse>('/admin/settings/site', { method: 'PATCH', body: site })
  }

  async function updateAdminPath(basePath: string) {
    return request<AdminPathResponse>('/admin/settings/admin-path', {
      method: 'PATCH',
      body: { base_path: basePath }
    })
  }

  return {
    admin,
    csrfToken,
    setCsrfToken,
    request,
    login,
    logout,
    me,
    changePassword,
    listTiers,
    saveTier,
    deleteTier,
    listPaymentMethods,
    savePaymentMethod,
    deletePaymentMethod,
    uploadPaymentQr,
    listDonations,
    updateDonationStatus,
    exportDonations,
    overview,
    settings,
    updateSiteSettings,
    updateAdminPath
  }
}
