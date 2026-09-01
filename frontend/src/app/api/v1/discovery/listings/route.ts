import { searchPublicListings } from "@/shared/api/generated";
import type {
  ErrorResponse,
  PublicListingsResponse,
  PublicListingResponse,
  SearchPublicListingsData,
} from "@/shared/api/generated";

const header = "X-Request-ID";
export const runtime = "nodejs";

export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers);
  const query = readQuery(new URL(request.url).searchParams);
  if (!query) return error("INVALID_REQUEST", "Invalid request", 400, id);
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(id);
  try {
    const upstream = await searchPublicListings({
      baseUrl: origin,
      query,
      headers: { [header]: id },
    });
    if (
      upstream.error ||
      !upstream.response?.ok ||
      upstream.response.headers.get(header) !== id ||
      !validListings(upstream.data)
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

function readQuery(
  values: URLSearchParams,
): SearchPublicListingsData["query"] | null {
  const allowed = [
    "locale",
    "categoryId",
    "q",
    "nearLocalityId",
    "radiusKm",
    "priceType",
    "serviceMode",
  ];
  if (![...values.keys()].every((key) => allowed.includes(key))) return null;
  const locale = one(values, "locale");
  if (!isLocale(locale)) return null;
  const categoryId = optional(values, "categoryId");
  const query = optional(values, "q");
  const nearLocalityId = optional(values, "nearLocalityId");
  const radiusKm = optional(values, "radiusKm");
  const priceType = optional(values, "priceType");
  const serviceMode = optional(values, "serviceMode");
  if (
    categoryId === undefined ||
    query === undefined ||
    nearLocalityId === undefined ||
    radiusKm === undefined ||
    priceType === undefined ||
    serviceMode === undefined ||
    (nearLocalityId === null) !== (radiusKm === null)
  )
    return null;
  const result: SearchPublicListingsData["query"] = { locale };
  if (categoryId !== null) {
    if (!uuid(categoryId)) return null;
    result.categoryId = categoryId;
  }
  if (query !== null) {
    const normalized = query.trim().replace(/\s+/g, " ");
    if (normalized.length < 2 || normalized.length > 80) return null;
    result.q = normalized;
  }
  if (nearLocalityId !== null && radiusKm !== null) {
    if (
      !uuid(nearLocalityId) ||
      !/^(?:[1-9]|[1-9][0-9]|1[0-9]{2}|200)$/.test(radiusKm)
    )
      return null;
    result.nearLocalityId = nearLocalityId;
    result.radiusKm = Number(radiusKm);
  }
  if (priceType !== null) {
    if (!isPriceType(priceType)) return null;
    result.priceType = priceType;
  }
  if (serviceMode !== null) {
    if (!isServiceMode(serviceMode)) return null;
    result.serviceMode = serviceMode;
  }
  return result;
}

function one(values: URLSearchParams, key: string): string | null {
  const entries = values.getAll(key);
  return entries.length === 1 && entries[0] ? entries[0] : null;
}

function optional(
  values: URLSearchParams,
  key: string,
): string | null | undefined {
  if (!values.has(key)) return null;
  return one(values, key) ?? undefined;
}

function validListings(
  value: PublicListingsResponse | undefined,
): value is PublicListingsResponse {
  return (
    exact(value, ["listings"]) &&
    Array.isArray(value.listings) &&
    value.listings.every(validListing)
  );
}

function validListing(value: PublicListingResponse): boolean {
  return (
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
    isPriceType(value.priceType) &&
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

function exact(
  value: unknown,
  expected: string[],
): value is Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value))
    return false;
  const actual = Object.keys(value).sort();
  const wanted = [...expected].sort();
  return (
    actual.length === wanted.length &&
    actual.every((key, index) => key === wanted[index])
  );
}
function isLocale(value: string | null): value is "pt-PT" | "en" | "es" {
  return value === "pt-PT" || value === "en" || value === "es";
}
function isPriceType(
  value: string,
): value is "fixed" | "hourly" | "daily" | "quote" | "negotiable" {
  return ["fixed", "hourly", "daily", "quote", "negotiable"].includes(value);
}
function isServiceMode(
  value: string,
): value is "travels_to_customer" | "receives_customer" | "remote_services" {
  return [
    "travels_to_customer",
    "receives_customer",
    "remote_services",
  ].includes(value);
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
