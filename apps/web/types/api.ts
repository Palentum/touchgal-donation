export type PaymentProviderType =
  | 'static_qr'
  | 'redirect_url'
  | 'mock_qr'
  | 'wechat_native'
  | 'alipay_f2f'
  | 'stripe_checkout'

export type DonationStatus = 'created' | 'pending' | 'paid' | 'failed' | 'cancelled' | 'expired' | 'refunded'

export interface ApiEnvelope<T> {
  data: T
  request_id?: string
}

export interface ApiErrorEnvelope {
  error?: {
    code?: string
    message?: string
    details?: unknown
  }
  request_id?: string
}

export interface PublicSiteConfig {
  name: string
  hero_title: string
  hero_subtitle: string
  currency: string
  timezone: string
  goal_cents: number
  show_goal: boolean
}

export interface DonationTier {
  id: string
  name: string
  amount_cents: number
  currency: string
  description: string
  sort_order?: number
  enabled?: boolean
  is_default: boolean
  created_at?: string
  updated_at?: string
}

export interface PaymentMethodConfig {
  qr_image_url?: string
  url_template?: string
  instructions?: string
  require_manual_confirm?: boolean
  [key: string]: unknown
}

export interface PaymentMethod {
  id: string
  code: string
  name: string
  type: PaymentProviderType
  provider?: string
  icon_url?: string
  config_json?: PaymentMethodConfig
  config?: PaymentMethodConfig
  enabled?: boolean
  sort_order?: number
  min_amount_cents?: number | null
  max_amount_cents?: number | null
  created_at?: string
  updated_at?: string
}

export interface PublicConfigResponse {
  site: PublicSiteConfig
  tiers: DonationTier[]
  payment_methods: PaymentMethod[]
}

export interface RecentDonation {
  nickname: string
  message: string
  amount_cents: number
  currency: string
  paid_at: string
}

export interface DonorInfoModel {
  nickname: string
  email: string
  message: string
  public_visible: boolean
}

export interface RecentDonationsResponse {
  items: RecentDonation[]
}

export interface CreateDonationRequest {
  tier_id: string | null
  amount_cents: number
  currency: string
  payment_method_id: string
  nickname: string
  email: string
  message: string
  public_visible: boolean
  client_request_id: string
}

export interface PaymentAction {
  mode: 'qr_image' | 'qr_content' | 'redirect'
  qr_image_url?: string
  qr_content?: string
  redirect_url?: string
  instructions?: string
  expires_at?: string
}

export interface CreateDonationResponse {
  order_no: string
  status: DonationStatus
  amount_cents: number
  currency: string
  payment_action: PaymentAction
  thanks_url?: string
}

export interface DonationStatusResponse {
  order_no: string
  status: DonationStatus
  amount_cents: number
  currency: string
  paid_at?: string | null
}

export interface AdminRouteResponse {
  kind: 'admin'
  base_path: string
  sub_path: string
}

export interface AdminUser {
  id: string
  username: string
  role: string
  must_change_password: boolean
}

export interface LoginResponse {
  admin: AdminUser
  csrf_token: string
}

export interface OverviewRange {
  start: string
  end: string
}

export interface OverviewDailyPoint {
  date: string
  paid_amount_cents: number
  paid_count: number
  order_count: number
}

export interface OverviewResponse {
  range: OverviewRange
  total_paid_amount_cents: number
  total_order_count: number
  paid_order_count: number
  pending_order_count: number
  failed_order_count: number
  daily: OverviewDailyPoint[]
}

export interface AdminDonationRecord {
  id: string
  order_no: string
  tier_id?: string | null
  payment_method_id?: string | null
  payment_method_name?: string | null
  payment_method?: PaymentMethod | null
  nickname: string
  email?: string
  message?: string
  amount_cents: number
  currency: string
  status: DonationStatus
  public_visible: boolean
  created_at: string
  paid_at?: string | null
}

export interface ListResponse<T> {
  items: T[]
  total?: number
  page?: number
  page_size?: number
}

export interface SettingsResponse {
  site: PublicSiteConfig
  admin: {
    base_path: string
  }
}

export interface AdminPathResponse {
  base_path: string
}

export interface UploadResponse {
  url?: string
  qr_image_url?: string
  icon_url?: string
}
