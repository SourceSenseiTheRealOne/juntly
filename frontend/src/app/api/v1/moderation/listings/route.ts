import { resolveSoleAdministratorSession } from "@/features/auth/sole-administrator";
import {
  approveListing,
  listPendingModerationListings,
  rejectListing,
} from "@/shared/api/generated";
import type {
  ErrorResponse,
  ListingResponse,
  ListingsResponse,
  RejectListingRequest,
  RevisionRequest,
} from "@/shared/api/generated";
const header = "X-Request-ID";
export const runtime = "nodejs";
export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    session = await resolveSoleAdministratorSession();
  if (session.status !== "authorized") return accessError(session.status, id);
  return upstream(
    id,
    session.token,
    () =>
      listPendingModerationListings({
        baseUrl: origin(),
        headers: headers(session.token, id),
      }),
    validListings,
  );
}
export async function POST(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    session = await resolveSoleAdministratorSession(),
    listingID = idFrom(request);
  if (session.status !== "authorized") return accessError(session.status, id);
  if (!listingID) return error("INVALID_REQUEST", "Invalid request", 400, id);
  const action = new URL(request.url).pathname.split("/").pop(),
    body = await json(request);
  if (action === "approve" && revision(body))
    return upstream(
      id,
      session.token,
      () =>
        approveListing({
          baseUrl: origin(),
          path: { listingId: listingID },
          body: body as RevisionRequest,
          headers: headers(session.token, id),
        }),
      validListing,
    );
  if (action === "reject" && reject(body))
    return upstream(
      id,
      session.token,
      () =>
        rejectListing({
          baseUrl: origin(),
          path: { listingId: listingID },
          body: body as RejectListingRequest,
          headers: headers(session.token, id),
        }),
      validListing,
    );
  return error("INVALID_REQUEST", "Invalid request", 400, id);
}
async function upstream<T>(
  id: string,
  token: string,
  call: () => Promise<{ data?: T; error?: unknown; response?: Response }>,
  valid: (v: T | undefined) => boolean,
) {
  if (!origin()) return unavailable(id);
  try {
    const up = await call();
    if (
      up.response?.status === 403 &&
      validError(up.error, "FORBIDDEN", id) &&
      up.response.headers.get(header) === id
    )
      return Response.json(up.error, {
        status: 403,
        headers: { [header]: id },
      });
    if (
      up.response?.status === 409 &&
      validError(up.error, "CONFLICT", id) &&
      up.response.headers.get(header) === id
    )
      return Response.json(up.error, {
        status: 409,
        headers: { [header]: id },
      });
    if (
      up.error ||
      !up.response?.ok ||
      up.response.headers.get(header) !== id ||
      !valid(up.data)
    )
      return unavailable(id);
    return Response.json(up.data, { status: 200, headers: { [header]: id } });
  } catch {
    return unavailable(id);
  }
}
function accessError(
  status: "unauthenticated" | "forbidden" | "unavailable",
  id: string,
) {
  if (status === "forbidden")
    return error("FORBIDDEN", "Forbidden", 403, id);
  if (status === "unavailable") return unavailable(id);
  return error("UNAUTHORIZED", "Unauthorized", 401, id);
}
async function json(r: Request) {
  try {
    return JSON.parse(await r.text());
  } catch {
    return null;
  }
}
function idFrom(r: Request) {
  const p = new URL(r.url).pathname.split("/"),
    i = p.indexOf("listings"),
    v = i >= 0 ? p[i + 1] : "";
  return uuid(v) ? v : null;
}
function origin() {
  return process.env.JUNTLY_API_ORIGIN;
}
function revision(v: unknown) {
  return (
    exact(v, ["revision"]) &&
    Number.isInteger((v as Record<string, unknown>).revision) &&
    Number((v as Record<string, unknown>).revision) > 0
  );
}
function reject(v: unknown) {
  if (!exact(v, ["revision", "reason"])) return false;
  const reason = v.reason;
  return (
    revision({ revision: v.revision }) &&
    typeof reason === "string" &&
    reason.trim().length > 0 &&
    reason.length <= 500
  );
}
function validListings(v: ListingsResponse | undefined) {
  return (
    !!v &&
    exact(v, ["listings"]) &&
    Array.isArray(v.listings) &&
    v.listings.every(validListing)
  );
}
function validListing(v: ListingResponse | undefined) {
  return (
    !!v &&
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
    v.currency === "EUR"
  );
}
function validError(v: unknown, c: string, id: string): v is ErrorResponse {
  return (
    exact(v, ["error"]) &&
    exact((v as { error: unknown }).error, ["code", "message", "requestId"]) &&
    (v as ErrorResponse).error.code === c &&
    (v as ErrorResponse).error.requestId === id
  );
}
function exact(v: unknown, k: string[]): v is Record<string, unknown> {
  if (v === null || typeof v !== "object" || Array.isArray(v)) return false;
  const a = Object.keys(v).sort(),
    b = [...k].sort();
  return a.length === b.length && a.every((x, i) => x === b[i]);
}
function uuid(v: unknown): v is string {
  return (
    typeof v === "string" &&
    /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(v)
  );
}
function headers(t: string, id: string): HeadersInit {
  return { Authorization: `Bearer ${t}`, [header]: id };
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
  c: ErrorResponse["error"]["code"],
  m: string,
  s: number,
  id: string,
) {
  return Response.json(
    { error: { code: c, message: m, requestId: id } } satisfies ErrorResponse,
    { status: s, headers: { [header]: id } },
  );
}
