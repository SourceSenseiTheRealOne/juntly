import type { AdministrationDashboard } from "@/shared/api/generated";

export function validDashboard(
  value: unknown,
): value is AdministrationDashboard {
  return (
    exact(value, ["metrics", "queue"]) &&
    exact(value.metrics, [
      "users",
      "providers",
      "activeListings",
      "completedBookings",
      "publishedReviews",
      "openReports",
    ]) &&
    Object.values(value.metrics).every(Number.isInteger) &&
    exact(value.queue, ["reports", "reviews"]) &&
    Array.isArray(value.queue.reports) &&
    value.queue.reports.length <= 100 &&
    value.queue.reports.every(
      (item) =>
        exact(item, ["id", "conversationId", "reason", "createdAt"]) &&
        uuid(item.id) &&
        uuid(item.conversationId) &&
        typeof item.reason === "string" &&
        dateTime(item.createdAt),
    ) &&
    Array.isArray(value.queue.reviews) &&
    value.queue.reviews.length <= 100 &&
    value.queue.reviews.every(
      (item) =>
        exact(item, ["id", "rating", "body", "state", "createdAt"]) &&
        uuid(item.id) &&
        Number.isInteger(item.rating) &&
        typeof item.body === "string" &&
        ["published", "hidden"].includes(String(item.state)) &&
        dateTime(item.createdAt),
    )
  );
}
function exact(
  value: unknown,
  expected: string[],
): value is Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value))
    return false;
  const actual = Object.keys(value).sort(),
    wanted = [...expected].sort();
  return (
    actual.length === wanted.length &&
    actual.every((key, index) => key === wanted[index])
  );
}
function uuid(value: unknown): value is string {
  return (
    typeof value === "string" &&
    /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(value)
  );
}
function dateTime(value: unknown): value is string {
  return typeof value === "string" && !Number.isNaN(Date.parse(value));
}
