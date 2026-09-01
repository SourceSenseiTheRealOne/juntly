import type { Booking, BookingsResponse } from "@/shared/api/generated";
import { dateTime, exact, uuid } from "@/features/messaging/protected-bff";
export function validBooking(value: unknown): value is Booking {
  return (
    exact(value, [
      "id",
      "customerId",
      "providerId",
      "sourceType",
      "sourceId",
      "state",
      "revision",
      "scheduledAt",
      "privateLocation",
      "agreedPriceMinor",
      "currency",
      "createdAt",
      "updatedAt",
    ]) &&
    uuid(value.id) &&
    uuid(value.customerId) &&
    uuid(value.providerId) &&
    ["proposal", "listing", "direct"].includes(String(value.sourceType)) &&
    (value.sourceId === null || uuid(value.sourceId)) &&
    [
      "draft",
      "pending_provider_confirmation",
      "confirmed",
      "scheduled",
      "in_progress",
      "completed",
      "cancelled",
      "disputed",
      "refunded",
    ].includes(String(value.state)) &&
    Number.isInteger(value.revision) &&
    Number(value.revision) > 0 &&
    dateTime(value.scheduledAt) &&
    typeof value.privateLocation === "string" &&
    value.privateLocation.length >= 5 &&
    Number.isInteger(value.agreedPriceMinor) &&
    Number(value.agreedPriceMinor) > 0 &&
    value.currency === "EUR" &&
    dateTime(value.createdAt) &&
    dateTime(value.updatedAt)
  );
}
export function validBookings(value: unknown): value is BookingsResponse {
  return (
    exact(value, ["bookings"]) &&
    Array.isArray(value.bookings) &&
    value.bookings.every(validBooking)
  );
}
