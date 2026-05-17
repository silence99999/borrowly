export type UserRole = "USER" | "ADMIN";

export interface AuthUser {
  id: string;
  role: UserRole | null;
}

export interface SignUpResponse {
  verification_id: string;
  message: string;
}

export interface ApiMessage {
  message: string;
}

export interface PickupPoint {
  id: string;
  address: string;
  city: string;
  is_active?: boolean;
}

export interface ItemList {
  id: string;
  title: string;
  category: string;
  price_per_hour: number;
  price_per_day: number;
  pickup_city: string;
  is_platform_item: boolean;
}

export interface ItemDetails {
  id: string;
  title: string;
  description: string;
  category: string;
  price_per_hour: number;
  price_per_day: number;
  pickup_city: string;
  is_platform_item: boolean;
  pickup_point?: PickupPoint;
  PickupPoint?: PickupPoint;
  owner?: {
    id: string;
    email?: string;
    created_at?: string;
  };
  Owner?: {
    id: string;
    email?: string;
    created_at?: string;
  };
}

export interface Rental {
  id: string;
  status: string;
  start_at: string;
  end_at: string;
  total_price: number;
  item: {
    id: string;
    title: string;
    category: string;
  };
  pickup_point: {
    city: string;
    address: string;
  };
}

export interface Review {
  id: string;
  item_id: string;
  reviewed_user_id: string;
  rating: number;
  comment: string;
  created_at: string;
}

export interface ItemPayload {
  pickup_point_id: string;
  title: string;
  description: string;
  category: string;
  price_per_hour: number;
  price_per_day: number;
}

export interface ReviewPayload {
  rating: number;
  comment: string;
}

export interface RentalPayload {
  item_id: string;
  start_at: string;
  end_at: string;
}

export interface PickupPointPayload {
  address: string;
  city: string;
  is_active: boolean;
}

export interface ChatMessage {
  id: string;
  sender_id: string;
  sender_email: string;
  receiver_id: string;
  content: string;
  is_read: boolean;
  created_at: string;
}

export interface Conversation {
  user_id: string;
  email: string;
  last_message: string;
  updated_at: string;
}
