import { getPublicListing } from "@/shared/api/generated";
import type {
  ErrorResponse,
  PublicListingResponse,
} from "@/shared/api/generated";

const header = "X-Request-ID";
export const runtime = "nodejs";

type RouteContext = { params: Promise<{ listingId: string }> };

export async function GET(
  request: Request,
  context: RouteContext,
): Promise<Response> {
  const id = requestID(request.headers);
  const locale = localeQuery(new URL(request.url).searchParams);
  const { listingId } = await context.params;
  if (!locale || !uuid(listingId))
    return error("INVALID_REQUEST", "Invalid request", 400, id);
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(id);
  try {
    const upstream = await getPublicListing({
      baseUrl: origin,
      path: { listingId },
      query: { locale },
      headers: { [header]: id },
    });
    if (
      upstream.response?.status === 404 &&
      validError(upstream.error, "NOT_FOUND", id) &&
      upstream.response.headers.get(header) === id
    )
      return Response.json(upstream.error, {
        status: 404,
        headers: { [header]: id },
      });
    if (
      upstream.error ||
      !upstream.response?.ok ||
      upstream.response.headers.get(header) !== id ||
      !validListing(upstream.data)
    )
      return unavailable(id);
    return Response.json(upstream.data, {
      status: 200,
      headers: { [header]: id },
    });
  } catch {
    return unavailable(id);
  }
}

function localeQuery(values: URLSearchParams): "pt-PT" | "en" | "es" | null {
  if (![...values.keys()].every((key) => key === "locale")) return null;
  const entries = values.getAll("locale");
  if (entries.length !== 1) return null;
  return entries[0] === "pt-PT" || entries[0] === "en" || entries[0] === "es"
    ? entries[0]
    : null;
}

function validListing(
  value: PublicListingResponse | undefined,
): value is PublicListingResponse {
  return (
    !!value &&
    exact(value, [
      "id",
      "title",
      "description",
      "categoryId",
      "categorySlug",
      "categoryName",
      "primaryLocalityId",
      "localitySlug",
      "localityName",
      "priceType",
      "priceMinor",
      "currency",
      "travelsToCustomer",
      "receivesCustomer",
      "remoteServices",
      "providerDisplayName",
      "providerType",
      "promoted",
      "updatedAt",
    ]) &&
    uuid(value.id) &&
    uuid(value.categoryId) &&
    uuid(value.primaryLocalityId) &&
    typeof value.title === "string" &&
    value.title.length >= 2 &&
    value.title.length <= 140 &&
    typeof value.description === "string" &&
    value.description.length >= 20 &&
    value.description.length <= 4000 &&
    typeof value.categorySlug === "string" &&
    typeof value.categoryName === "string" &&
    typeof value.localitySlug === "string" &&
    typeof value.localityName === "string" &&
    ["fixed", "hourly", "daily", "quote", "negotiable"].includes(
      value.priceType,
    ) &&
    (value.priceMinor === null ||
      (Number.isInteger(value.priceMinor) && value.priceMinor > 0)) &&
    value.currency === "EUR" &&
    typeof value.travelsToCustomer === "boolean" &&
    typeof value.receivesCustomer === "boolean" &&
    typeof value.remoteServices === "boolean" &&
    typeof value.providerDisplayName === "string" &&
    value.providerDisplayName.length >= 2 &&
    ["individual", "professional", "business"].includes(value.providerType) &&
    typeof value.promoted === "boolean" &&
    typeof value.updatedAt === "string" &&
    !Number.isNaN(Date.parse(value.updatedAt))
  );
}

function validError(
  value: unknown,
  code: string,
  id: string,
): value is ErrorResponse {
  return (
    exact(value, ["error"]) &&
    exact((value as { error: unknown }).error, [
      "code",
      "message",
      "requestId",
    ]) &&
    (value as ErrorResponse).error.code === code &&
    (value as ErrorResponse).error.requestId === id
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
function uuid(value: string): boolean {
  return /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(value);
}
function requestID(headers: Headers): string {
  const value = headers.get(header);
  return value && /^[A-Za-z0-9._:-]{8,128}$/.test(value)
    ? value
    : `req_${crypto.randomUUID()}`;
}
function unavailable(id: string): Response {
  return error("SERVICE_UNAVAILABLE", "Service unavailable", 503, id);
}
function error(
  code: ErrorResponse["error"]["code"],
  message: string,
  status: number,
  id: string,
): Response {
  return Response.json(
    { error: { code, message, requestId: id } } satisfies ErrorResponse,
    { status, headers: { [header]: id } },
  );
}
