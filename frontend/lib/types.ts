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
  count: number;
}
