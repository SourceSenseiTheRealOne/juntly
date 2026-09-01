import type {
  ProviderRating,
  Review,
  ReviewsResponse,
} from "@/shared/api/generated";
import { dateTime, exact, uuid } from "@/features/messaging/protected-bff";
export function validReview(v: unknown): v is Review {
  return (
    exact(v, [
      "id",
      "bookingId",
      "customerId",
      "providerId",
      "rating",
      "body",
      "providerResponse",
      "verifiedBooking",
      "state",
      "createdAt",
      "updatedAt",
    ]) &&
    uuid(v.id) &&
    uuid(v.bookingId) &&
    uuid(v.customerId) &&
    uuid(v.providerId) &&
    Number.isInteger(v.rating) &&
    Number(v.rating) >= 1 &&
    Number(v.rating) <= 5 &&
    typeof v.body === "string" &&
    typeof v.providerResponse === "string" &&
    v.verifiedBooking === true &&
    ["published", "hidden"].includes(String(v.state)) &&
    dateTime(v.createdAt) &&
    dateTime(v.updatedAt)
  );
}
export function validReviews(v: unknown): v is ReviewsResponse {
  return (
    exact(v, ["reviews"]) &&
    Array.isArray(v.reviews) &&
    v.reviews.every(validReview)
  );
}
export function validRating(v: unknown): v is ProviderRating {
  return (
    exact(v, ["providerId", "averageRating", "reviewCount"]) &&
    uuid(v.providerId) &&
    typeof v.averageRating === "number" &&
    v.averageRating >= 0 &&
    v.averageRating <= 5 &&
    Number.isInteger(v.reviewCount) &&
    Number(v.reviewCount) >= 0
  );
}
