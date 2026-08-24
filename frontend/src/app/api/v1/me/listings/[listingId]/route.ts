import { auth } from "@clerk/nextjs/server";
import {
  archiveListing,
  createListingMediaUploadIntent,
  pauseListing,
  replaceMyDraftListing,
  submitListingForReview,
} from "@/shared/api/generated";
import type {
  ArchiveListingRequest,
  CreateListingMediaUploadIntentData,
  ErrorResponse,
  ListingResponse,
  ReplaceDraftListingRequest,
  RevisionRequest,
  UploadIntentResponse,
} from "@/shared/api/generated";
const header = "X-Request-ID";
export const runtime = "nodejs";
export async function PUT(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    token = await tokenForSession(),
    listingID = idFrom(request);
  if (!token) return error("UNAUTHORIZED", "Unauthorized", 401, id);
  const body = await json(request);
  if (!listingID || !validReplace(body))
    return error("INVALID_REQUEST", "Invalid request", 400, id);
  return upstream(
    id,
    token,
    () =>
      replaceMyDraftListing({
        baseUrl: origin(),
        path: { listingId: listingID },
        body: body as ReplaceDraftListingRequest,
        headers: headers(token, id),
      }),
    validListing,
  );
}
export async function POST(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    token = await tokenForSession(),
    listingID = idFrom(request);
  if (!token) return error("UNAUTHORIZED", "Unauthorized", 401, id);
  if (!listingID) return error("INVALID_REQUEST", "Invalid request", 400, id);
  const action = new URL(request.url).pathname.split("/").pop(),
    body = await json(request);
  if (action === "submit" && validRevision(body))
    return upstream(
      id,
      token,
      () =>
        submitListingForReview({
          baseUrl: origin(),
          path: { listingId: listingID },
          body: body as RevisionRequest,
          headers: headers(token, id),
        }),
      validListing,
    );
  if (action === "pause" && validRevision(body))
    return upstream(
      id,
      token,
      () =>
        pauseListing({
          baseUrl: origin(),
          path: { listingId: listingID },
          body: body as RevisionRequest,
          headers: headers(token, id),
        }),
      validListing,
    );
  if (action === "archive" && validArchive(body))
    return upstream(
      id,
      token,
      () =>
        archiveListing({
          baseUrl: origin(),
          path: { listingId: listingID },
          body: body as ArchiveListingRequest,
          headers: headers(token, id),
        }),
      validListing,
    );
  if (action === "upload-intents" && validUpload(body))
    return upstream(
      id,
      token,
      () =>
        createListingMediaUploadIntent({
          baseUrl: origin(),
          path: { listingId: listingID },
          body: body as CreateListingMediaUploadIntentData["body"],
          headers: headers(token, id),
        }),
      validIntent,
    );
  return error("INVALID_REQUEST", "Invalid request", 400, id);
}
async function upstream<T>(
  id: string,
  token: string,
  call: () => Promise<{ data?: T; error?: unknown; response?: Response }>,
  valid: (v: T | undefined) => boolean,
): Promise<Response> {
  if (!origin()) return unavailable(id);
  try {
    const up = await call();
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
async function tokenForSession(): Promise<string | null> {
  try {
    const state = await auth();
    return state.isAuthenticated ? await state.getToken() : null;
  } catch {
    return null;
  }
}
async function json(request: Request): Promise<unknown> {
  try {
    return JSON.parse(await request.text());
  } catch {
    return null;
  }
}
function idFrom(request: Request) {
  const parts = new URL(request.url).pathname.split("/");
  const index = parts.indexOf("listings");
  const value = index >= 0 ? parts[index + 1] : "";
  return uuid(value) ? value : null;
}
function origin() {
  return process.env.JUNTLY_API_ORIGIN;
}
function validReplace(v: unknown) {
  return (
    exact(v, [
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
      "revision",
    ]) &&
    validCreate(v) &&
    validRevision(v)
  );
}
function validCreate(v: unknown) {
  if (
    !exact(v, [
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
    ]) &&
    !exact(v, [
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
      "revision",
    ])
  )
    return false;
  const x = v as Record<string, unknown>;
  return (
    uuid(x.categoryId) &&
    uuid(x.primaryLocalityId) &&
    typeof x.title === "string" &&
    x.title.trim().length >= 2 &&
    typeof x.description === "string" &&
    x.description.trim().length >= 20 &&
    ["fixed", "hourly", "daily", "quote", "negotiable"].includes(
      String(x.priceType),
    ) &&
    x.currency === "EUR" &&
    typeof x.travelsToCustomer === "boolean" &&
    typeof x.receivesCustomer === "boolean" &&
    typeof x.remoteServices === "boolean" &&
    (x.priceMinor === null ||
      (Number.isInteger(x.priceMinor) && Number(x.priceMinor) > 0))
  );
}
function validRevision(v: unknown) {
  return (
    typeof v === "object" &&
    v !== null &&
    !Array.isArray(v) &&
    Number.isInteger((v as Record<string, unknown>).revision) &&
    Number((v as Record<string, unknown>).revision) > 0
  );
}
function validArchive(v: unknown) {
  return (
    exact(v, ["revision", "state"]) &&
    validRevision(v) &&
    ["draft", "rejected", "active", "paused"].includes(
      String((v as Record<string, unknown>).state),
    )
  );
}
function validUpload(v: unknown) {
  return (
    exact(v, ["ordinal", "contentType", "byteSize", "checksumSha256"]) &&
    Number.isInteger((v as Record<string, unknown>).ordinal) &&
    Number((v as Record<string, unknown>).ordinal) >= 1 &&
    Number((v as Record<string, unknown>).ordinal) <= 10 &&
    ["image/jpeg", "image/png", "image/webp"].includes(
      String((v as Record<string, unknown>).contentType),
    ) &&
    Number.isInteger((v as Record<string, unknown>).byteSize) &&
    Number((v as Record<string, unknown>).byteSize) > 0 &&
    Number((v as Record<string, unknown>).byteSize) <= 10485760 &&
    typeof (v as Record<string, unknown>).checksumSha256 === "string" &&
    /^[0-9a-f]{64}$/.test(String((v as Record<string, unknown>).checksumSha256))
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
function validIntent(v: UploadIntentResponse | undefined) {
  return (
    !!v &&
    exact(v, ["mediaId", "capability"]) &&
    uuid(v.mediaId) &&
    exact(v.capability, ["url", "method", "headers"]) &&
    typeof v.capability.url === "string" &&
    v.capability.method === "PUT"
  );
}
function validError(v: unknown, code: string, id: string): v is ErrorResponse {
  return (
    exact(v, ["error"]) &&
    exact((v as { error: unknown }).error, ["code", "message", "requestId"]) &&
    (v as ErrorResponse).error.code === code &&
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
function headers(t: string, i: string): HeadersInit {
  return { Authorization: `Bearer ${t}`, [header]: i };
}
function requestID(h: Headers) {
  const v = h.get(header);
  return v && /^[A-Za-z0-9._:-]{8,128}$/.test(v)
    ? v
    : `req_${crypto.randomUUID()}`;
}
function unavailable(i: string) {
  return error("SERVICE_UNAVAILABLE", "Service unavailable", 503, i);
}
function error(
  c: ErrorResponse["error"]["code"],
  m: string,
  s: number,
  i: string,
) {
  return Response.json(
    { error: { code: c, message: m, requestId: i } } satisfies ErrorResponse,
    { status: s, headers: { [header]: i } },
  );
}
