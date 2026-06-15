import type {
  AdminRouteResponse,
  ApiEnvelope,
  ApiErrorEnvelope,
  CreateDonationRequest,
  CreateDonationResponse,
  DonationStatusResponse,
  PublicConfigResponse,
  RecentDonationsResponse
} from '~/types/api'

type ApiRequestOptions = {
  method?: 'GET' | 'POST' | 'PATCH' | 'DELETE'
  body?: Record<string, any> | BodyInit | null
  query?: Record<string, string | number | boolean | undefined>
  headers?: HeadersInit
  credentials?: RequestCredentials
  toastErrors?: boolean
}

function unwrapEnvelope<T>(payload: ApiEnvelope<T> | T): T {
  if (payload && typeof payload === 'object' && 'data' in payload) {
    return (payload as ApiEnvelope<T>).data
  }
  return payload as T
}

function errorMessage(error: unknown) {
  const possibleData = (error as { data?: ApiErrorEnvelope })?.data
  return possibleData?.error?.message || (error as Error)?.message || '请求失败，请稍后重试'
}

export function useApi() {
  const config = useRuntimeConfig()
  const toast = useToast()
  const baseURL = import.meta.server && config.apiInternalBase ? String(config.apiInternalBase) : String(config.public.apiBase || '')

  async function request<T>(path: string, options: ApiRequestOptions = {}) {
    const { toastErrors = true, ...fetchOptions } = options

    try {
      const response = await $fetch<ApiEnvelope<T> | T>(path, {
        baseURL,
        ...fetchOptions
      })
      return unwrapEnvelope<T>(response)
    } catch (error) {
      if (toastErrors) {
        toast.error(errorMessage(error))
      }
      throw error
    }
  }

  return {
    request,
    getPublicConfig: () => request<PublicConfigResponse>('/public/config'),
    getRecentDonations: (days = 30) =>
      request<RecentDonationsResponse>('/public/donations/recent', { query: { days } }),
    createDonation: (payload: CreateDonationRequest) =>
      request<CreateDonationResponse>('/public/donations', { method: 'POST', body: payload }),
    getDonationStatus: (orderNo: string) =>
      request<DonationStatusResponse>(`/public/donations/${encodeURIComponent(orderNo)}/status`, {
        toastErrors: false
      }),
    resolveRoute: (path: string) =>
      request<AdminRouteResponse>('/public/resolve-route', {
        query: { path },
        toastErrors: false
      })
  }
}
