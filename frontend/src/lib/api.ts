import type {
  ApiMessage,
  AuthUser,
  ChatMessage,
  Conversation,
  ItemDetails,
  ItemList,
  ItemPayload,
  PickupPoint,
  PickupPointPayload,
  Rental,
  RentalPayload,
  Review,
  ReviewPayload,
  SignUpResponse,
} from "../types/api";
import { normalizeRole } from "./utils";

const API_PREFIX = "/api";

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_PREFIX}${path}`, {
    credentials: "include",
    ...init,
  });

  const contentType = response.headers.get("content-type") ?? "";
  const hasJson = contentType.includes("application/json");
  const body = hasJson ? await response.json().catch(() => null) : null;

  if (!response.ok) {
    const message =
      typeof body?.error === "string"
        ? body.error
        : typeof body?.message === "string"
          ? body.message
          : `Request failed with status ${response.status}`;
    throw new ApiError(response.status, message);
  }

  return body as T;
}

function jsonRequest<T>(path: string, method: string, body?: unknown) {
  return request<T>(path, {
    method,
    headers: {
      "Content-Type": "application/json",
    },
    body: body ? JSON.stringify(body) : undefined,
  });
}

export const api = {
  signUp(payload: { email: string; password: string }) {
    return jsonRequest<SignUpResponse>("/auth/signup", "POST", payload);
  },
  login(payload: { email: string; password: string }) {
    return jsonRequest<ApiMessage>("/auth/login", "POST", payload);
  },
  verifyEmail(payload: { verification_id: string; code: string }) {
    return jsonRequest<void>("/auth/verify-email", "POST", payload);
  },
  logout() {
    return request<void>("/auth/logout", { method: "POST" });
  },
  async me(): Promise<AuthUser | null> {
    try {
      const response = await request<{ id: string; role: string }>("/auth/me");
      return {
        id: response.id,
        role: normalizeRole(response.role),
      };
    } catch {
      return null;
    }
  },
  getItems() {
    return request<ItemList[]>("/items");
  },
  getPickupPoints() {
    return request<PickupPoint[]>("/pickup-points");
  },
  getItem(itemId: string) {
    return request<ItemDetails>(`/items/${itemId}`);
  },
  getMyItems() {
    return request<ItemList[]>("/items/my");
  },
  createItem(payload: ItemPayload) {
    return jsonRequest<{ id: string }>("/items", "POST", payload);
  },
  updateItem(itemId: string, payload: Partial<ItemPayload>) {
    return jsonRequest<void>(`/items/${itemId}`, "PATCH", payload);
  },
  deleteItem(itemId: string) {
    return request<void>(`/items/${itemId}`, { method: "DELETE" });
  },
  uploadItemImage(itemId: string, file: File) {
    const formData = new FormData();
    formData.append("file", file);

    return request<ApiMessage>(`/items/${itemId}/images`, {
      method: "POST",
      body: formData,
    });
  },
  createRental(payload: RentalPayload) {
    return jsonRequest("/rentals", "POST", payload);
  },
  getMyRentals() {
    return request<Rental[]>("/rentals/my");
  },
  cancelRental(rentalId: string) {
    return request<void>(`/rental/${rentalId}/cancel`, { method: "PATCH" });
  },
  payRental(rentalId: string) {
    return request<void>(`/rental/${rentalId}/pay`, { method: "PATCH" });
  },
  getReviews(itemId: string) {
    return request<Review[]>(`/items/${itemId}/reviews`);
  },
  createReview(itemId: string, payload: ReviewPayload) {
    return jsonRequest(`/items/${itemId}/reviews`, "POST", payload);
  },
  updateReview(itemId: string, reviewId: string, payload: Partial<ReviewPayload>) {
    return jsonRequest<ApiMessage>(`/items/${itemId}/reviews/${reviewId}`, "PATCH", payload);
  },
  deleteReview(itemId: string, reviewId: string) {
    return request<ApiMessage>(`/items/${itemId}/reviews/${reviewId}`, { method: "DELETE" });
  },
  createPickupPoint(payload: PickupPointPayload) {
    return jsonRequest<PickupPoint>("/pickup-points", "POST", payload);
  },
  getConversations() {
    return request<Conversation[]>("/chats");
  },
  getMessages(userId: string) {
    return request<ChatMessage[]>(`/chats/${userId}`);
  },
  sendMessage(userId: string, content: string) {
    return jsonRequest<ChatMessage>(`/chats/${userId}`, "POST", { content });
  },
};
