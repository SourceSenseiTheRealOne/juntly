import { auth } from "@clerk/nextjs/server";
import { createListing, listMyListings } from "@/shared/api/generated";
import type {
  CreateListingRequest,
  ErrorResponse,
  ListingResponse,
  ListingsResponse,
} from "@/shared/api/generated";

const header = "X-Request-ID";
export const runtime = "nodejs";
export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    token = await tokenForSession();
  if (!token) return error("UNAUTHORIZED", "Unauthorized", 401, id);
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(id);
  try {
    const up = await listMyListings({
      baseUrl: origin,
      headers: headers(token, id),
    });
    if (
      up.error ||
      !up.response?.ok ||
      up.response.headers.get(header) !== id ||
      !validListings(up.data)
    )
      return unavailable(id);
    return Response.json(up.data, { status: 200, headers: { [header]: id } });
  } catch {
    return unavailable(id);
  }
}
export async function POST(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    token = await tokenForSession();
  if (!token) return error("UNAUTHORIZED", "Unauthorized", 401, id);
  const body = await readCreate(request);
  if (!body) return error("INVALID_REQUEST", "Invalid request", 400, id);
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(id);
  try {
    const up = await createListing({
      baseUrl: origin,
      body,
      headers: headers(token, id),
    });
    if (
      up.error ||
      !up.response?.ok ||
      up.response.headers.get(header) !== id ||
      !validListing(up.data)
    )
      return unavailable(id);
    return Response.json(up.data, { status: 200, headers: { [header]: id } });
  } catch {
    return unavailable(id);
  }
}
async function tokenForSession(): Promise<string | null> {
  try {
    const state = await auth();
    return state.isAuthenticated ? await state.getToken() : null;
  } catch {
    return null;
  }
}
async function readCreate(
  request: Request,
): Promise<CreateListingRequest | null> {
  try {
    const value: unknown = JSON.parse(await request.text());
    return validCreate(value) ? (value as CreateListingRequest) : null;
  } catch {
    return null;
  }
}
function validCreate(value: unknown): value is Record<string, unknown> {
  if (
    !exact(value, [
      "categoryId",
      "primaryLocalityId",
      "title",
      "description",
      "priceType",
      "priceMinor",
      "currency",
      "travelsToCustomer",
      "receivesCustomer",
      "remoteServices",
    ])
  )
    return false;
  const v = value as Record<string, unknown>;
  return (
    uuid(v.categoryId) &&
    uuid(v.primaryLocalityId) &&
    typeof v.title === "string" &&
    v.title.trim().length >= 2 &&
    v.title.length <= 140 &&
    typeof v.description === "string" &&
    v.description.trim().length >= 20 &&
    v.description.length <= 4000 &&
    ["fixed", "hourly", "daily", "quote", "negotiable"].includes(
      String(v.priceType),
    ) &&
    v.currency === "EUR" &&
    typeof v.travelsToCustomer === "boolean" &&
    typeof v.receivesCustomer === "boolean" &&
    typeof v.remoteServices === "boolean" &&
    (v.priceMinor === null ||
      (Number.isInteger(v.priceMinor) && Number(v.priceMinor) > 0))
  );
}
function validListings(v: ListingsResponse | undefined): v is ListingsResponse {
  return (
    exact(v, ["listings"]) &&
    Array.isArray(v.listings) &&
    v.listings.every(validListing)
  );
}
function validListing(v: ListingResponse | undefined): v is ListingResponse {
  return (
    exact(v, [
      "id",
      "categoryId",
      "primaryLocalityId",
      "title",
      "description",
      "priceType",
      "priceMinor",
      "currency",
      "travelsToCustomer",
      "receivesCustomer",
      "remoteServices",
      "state",
      "revision",
      "createdAt",
      "updatedAt",
    ]) &&
    uuid(v.id) &&
    uuid(v.categoryId) &&
    uuid(v.primaryLocalityId) &&
    typeof v.title === "string" &&
    typeof v.description === "string" &&
    ["fixed", "hourly", "daily", "quote", "negotiable"].includes(
      String(v.priceType),
    ) &&
    (v.priceMinor === null ||
      (Number.isInteger(v.priceMinor) && v.priceMinor > 0)) &&
    v.currency === "EUR" &&
    typeof v.travelsToCustomer === "boolean" &&
    typeof v.receivesCustomer === "boolean" &&
    typeof v.remoteServices === "boolean" &&
    [
      "draft",
      "pending_review",
      "active",
      "rejected",
      "paused",
      "archived",
    ].includes(String(v.state)) &&
    Number.isInteger(v.revision) &&
    v.revision > 0 &&
    typeof v.createdAt === "string" &&
    !Number.isNaN(Date.parse(v.createdAt)) &&
    typeof v.updatedAt === "string" &&
    !Number.isNaN(Date.parse(v.updatedAt))
  );
}
function exact(
  value: unknown,
  keys: string[],
): value is Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value))
    return false;
  const a = Object.keys(value).sort(),
    b = [...keys].sort();
  return a.length === b.length && a.every((key, i) => key === b[i]);
}
function uuid(v: unknown): v is string {
  return (
    typeof v === "string" &&
    /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(v)
  );
}
function headers(token: string, id: string): HeadersInit {
  return { Authorization: `Bearer ${token}`, [header]: id };
}
function requestID(h: Headers) {
  const v = h.get(header);
  return v && /^[A-Za-z0-9._:-]{8,128}$/.test(v)
    ? v
    : `req_${crypto.randomUUID()}`;
}
function unavailable(id: string) {
  return error("SERVICE_UNAVAILABLE", "Service unavailable", 503, id);
}
function error(
  code: ErrorResponse["error"]["code"],
  message: string,
  status: number,
  id: string,
) {
  return Response.json(
    { error: { code, message, requestId: id } } satisfies ErrorResponse,
    { status, headers: { [header]: id } },
  );
}
