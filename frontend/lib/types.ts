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

export interface TrackingEvent {
  EventDateTime: string;
  EventLocation: string;
  EventDescription: string;
  EventType: string;
}

export interface TrackingInfo {
  number: string;
  carrier: string;
  service_code?: string;
  url?: string;
  status: ShipmentStatus;
  status_label: string;
  events?: TrackingEvent[];
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
  status: string;
  status_label: string;
  customer_name: string;
  customer_city: string;
  customer_state: string;
  total: string;
  created_at: string;
  paid_at?: string;
  tracking: TrackingInfo;
}

export interface OrderDetail extends OrderListItem {
  email: string;
  phone: string;
  line_items: LineItem[];
  shipping: Address;
  billing: Address;
  shipping_method?: string;
}

export interface OrdersResponse {
  orders: OrderListItem[];
  page: number;
  count: number;
}
