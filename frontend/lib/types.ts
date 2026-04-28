export type WCStatus =
  | "pending"
  | "processing"
  | "on-hold"
  | "completed"
  | "cancelled"
  | "refunded"
  | "failed"
  | "shipped"
  | "in-transit"
  | "out-for-delivery";

export type ShipmentStatus =
  | "unknown"
  | "shipped"
  | "in-transit"
  | "out-for-delivery"
  | "delivered"
  | "delivery-failed"
  | "returned";

export type Health = "unknown" | "on_track" | "at_risk" | "breached";

export interface TimelineEvent {
  occurred_at: string;
  description: string;
  location?: string;
  type?: string;
}

export interface TrackingView {
  number: string;
  carrier: string;
  service?: string;
  service_code?: string;
  url?: string;
  status: ShipmentStatus;
  status_label: string;
  health: Health;
  health_label: string;
  last_event?: string;
  last_event_at?: string;
  estimated_delivery?: string;
  delivered_at?: string;
  idle_since?: string;
  risk_score: number;
  events?: TimelineEvent[];
}

export interface Address {
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  city: string;
  state: string;
  postcode: string;
}

export interface LineItem {
  id: number;
  name: string;
  quantity: number;
  total: string;
}

export interface OrderListItem {
  id: number;
  wc_order_id: number;
  status: string;
  status_label: string;
  customer_name: string;
  customer_city: string;
  customer_state: string;
  total: number;
  created_at: string;
  paid_at?: string;
  tracking: TrackingView;
}

export interface OrderDetail extends OrderListItem {
  email: string;
  phone: string;
  shipping_method?: string;
  line_items: LineItem[];
  shipping: Address;
  billing: Address;
}

export interface OrdersResponse {
  orders: OrderListItem[];
  total: number;
  count: number;
  limit: number;
  offset: number;
}

export interface FacetValue {
  value: string;
  count: number;
}

export interface Facets {
  carriers: FacetValue[];
  ufs: FacetValue[];
  statuses: FacetValue[];
  health: FacetValue[];
}

export interface PreviousPeriod {
  total_30d: number;
  delivered_30d: number;
  on_time_30d: number;
  on_time_rate: number;
  avg_delivery_days: number;
}

export interface Overview {
  total_30d: number;
  delivered_30d: number;
  on_time_30d: number;
  on_time_rate: number;
  at_risk: number;
  breached: number;
  avg_delivery_days: number;
  idle_alarms: number;
  previous_period?: PreviousPeriod;
}

export interface CarrierStats {
  carrier: string;
  total: number;
  breached: number;
  avg_delivery_days: number;
}

export interface OverviewResponse {
  overview: Overview;
  carriers: CarrierStats[];
}

export interface WooCommerceIntegration {
  url: string;
  consumer_key: string;
  consumer_secret: string;
  webhook_secret: string;
  enabled: boolean;
  configured: boolean;
}

export interface FrenetIntegration {
  api_token: string;
  panel_email: string;
  panel_password: string;
  enabled: boolean;
  configured: boolean;
}

export interface WAHAIntegration {
  url: string;
  api_key: string;
  enabled: boolean;
  configured: boolean;
}

export interface IntegrationsResponse {
  woocommerce: WooCommerceIntegration;
  frenet: FrenetIntegration;
  waha: WAHAIntegration;
}

export interface TestResult {
  ok: boolean;
  message?: string;
  error?: string;
}

export interface SyncSource {
  entity: string;
  last_synced_at: string | null;
  seconds_ago: number;
}

export interface SyncStatusResponse {
  sources: SyncSource[];
}

export interface StatusChange {
  id: number;
  order_id: number;
  from_status: string;
  to_status: string;
  source: string;
  note: string;
  actor: string;
  created_at: string;
}

export interface OrderNotification {
  id: number;
  order_id: number;
  channel: string;
  template: string;
  payload?: Record<string, unknown>;
  status: string;
  error?: string;
  sent_at: string;
}

export interface OrderHistoryResponse {
  changes: StatusChange[];
  notifications: OrderNotification[];
}
