import type { ItemDetails, ItemList, UserRole } from "../types/api";

export function formatMoney(value?: number) {
  if (!value) {
    return "Not set";
  }

  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(value);
}

export function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function normalizeRole(role: string | null | undefined): UserRole | null {
  if (role === "USER" || role === "ADMIN") {
    return role;
  }

  return null;
}

export function mergeItemSummary(details: ItemDetails, summary?: ItemList): ItemDetails {
  return {
    ...details,
    price_per_hour: details.price_per_hour || summary?.price_per_hour || 0,
    price_per_day: details.price_per_day || summary?.price_per_day || 0,
    pickup_city: details.pickup_city || summary?.pickup_city || "",
    is_platform_item: details.is_platform_item || summary?.is_platform_item || false,
  };
}

export function classNames(...values: Array<string | false | null | undefined>) {
  return values.filter(Boolean).join(" ");
}
