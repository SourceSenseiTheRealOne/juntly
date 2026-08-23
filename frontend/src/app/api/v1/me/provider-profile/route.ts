import { auth } from "@clerk/nextjs/server";

import {
  getProviderProfile,
  replaceProviderProfile,
} from "@/shared/api/generated";
import type {
  ErrorResponse,
  ProviderProfileEnvelope,
  ReplaceProviderProfileRequest,
} from "@/shared/api/generated";

const requestIDHeader = "X-Request-ID";
export const runtime = "nodejs";

export async function GET(request: Request): Promise<Response> {
  const requestID = readRequestID(request.headers);
  const token = await sessionToken();
  if (!token)
    return errorResponse("UNAUTHORIZED", "Unauthorized", 401, requestID);
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(requestID);
  try {
    const upstream = await getProviderProfile({
      baseUrl: origin,
      headers: upstreamHeaders(token, requestID),
    });
    return profileUpstreamResponse(upstream, requestID);
  } catch {
    return unavailable(requestID);
  }
}

export async function PUT(request: Request): Promise<Response> {
  const requestID = readRequestID(request.headers);
  const token = await sessionToken();
  if (!token)
    return errorResponse("UNAUTHORIZED", "Unauthorized", 401, requestID);
  const body = await readReplacement(request);
  if (!body)
    return errorResponse("INVALID_REQUEST", "Invalid request", 400, requestID);
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(requestID);
  try {
    const upstream = await replaceProviderProfile({
      baseUrl: origin,
      body,
      headers: upstreamHeaders(token, requestID),
    });
    return profileUpstreamResponse(upstream, requestID);
  } catch {
    return unavailable(requestID);
  }
}

async function sessionToken(): Promise<string | null> {
  try {
    const state = await auth();
    return state.isAuthenticated ? await state.getToken() : null;
  } catch {
    return null;
  }
}

async function readReplacement(
  request: Request,
): Promise<ReplaceProviderProfileRequest | null> {
  try {
    const value: unknown = JSON.parse(await request.text());
    if (!isProfileFields(value, false)) return null;
    return value as ReplaceProviderProfileRequest;
  } catch {
    return null;
  }
}

function profileUpstreamResponse(
  upstream: {
    data?: ProviderProfileEnvelope;
    error?: unknown;
    response?: Response;
  },
  requestID: string,
): Response {
  if (
    upstream.response?.status === 403 &&
    isError(upstream.error, "FORBIDDEN", requestID) &&
    upstream.response.headers.get(requestIDHeader) === requestID
  ) {
    return Response.json(upstream.error, {
      status: 403,
      headers: { [requestIDHeader]: requestID },
    });
  }
  if (
    upstream.error ||
    !upstream.response?.ok ||
    upstream.response.headers.get(requestIDHeader) !== requestID ||
    !isEnvelope(upstream.data)
  )
    return unavailable(requestID);
  return Response.json(upstream.data, {
    status: 200,
    headers: { [requestIDHeader]: requestID },
  });
}

function isEnvelope(
  value: ProviderProfileEnvelope | undefined,
): value is ProviderProfileEnvelope {
  return (
    isExact(value, ["profile"]) &&
    (value.profile === null || isProfileFields(value.profile, true))
  );
}

function isProfileFields(
  value: unknown,
  timestamps: boolean,
): value is Record<string, unknown> {
  const base = [
    "bio",
    "displayName",
    "languageCodes",
    "maxTravelDistanceKm",
    "primaryLocalityId",
    "providerType",
    "receivesCustomer",
    "remoteServices",
    "serviceLocalityIds",
    "travelsToCustomer",
  ];
  const keys = timestamps ? [...base, "createdAt", "updatedAt"] : base;
  if (!isExact(value, keys)) return false;
  const v = value as Record<string, unknown>;
  if (
    typeof v.displayName !== "string" ||
    v.displayName.trim().length < 2 ||
    v.displayName.length > 100 ||
    !["individual", "professional", "business"].includes(
      String(v.providerType),
    ) ||
    typeof v.bio !== "string" ||
    v.bio.length > 1000 ||
    typeof v.primaryLocalityId !== "string" ||
    !isUUID(v.primaryLocalityId) ||
    !uuidArray(v.serviceLocalityIds, 1, 20) ||
    !Number.isInteger(v.maxTravelDistanceKm) ||
    Number(v.maxTravelDistanceKm) < 0 ||
    Number(v.maxTravelDistanceKm) > 200 ||
    typeof v.travelsToCustomer !== "boolean" ||
    typeof v.receivesCustomer !== "boolean" ||
    typeof v.remoteServices !== "boolean" ||
    !stringArray(v.languageCodes, 1, 10)
  )
    return false;
  if (!v.travelsToCustomer && !v.receivesCustomer && !v.remoteServices)
    return false;
  if (
    timestamps &&
    (typeof v.createdAt !== "string" ||
      Number.isNaN(Date.parse(v.createdAt)) ||
      typeof v.updatedAt !== "string" ||
      Number.isNaN(Date.parse(v.updatedAt)))
  )
    return false;
  return true;
}

function uuidArray(value: unknown, min: number, max: number): boolean {
  return (
    Array.isArray(value) &&
    value.length >= min &&
    value.length <= max &&
    new Set(value).size === value.length &&
    value.every((item) => typeof item === "string" && isUUID(item))
  );
}
function stringArray(value: unknown, min: number, max: number): boolean {
  return (
    Array.isArray(value) &&
    value.length >= min &&
    value.length <= max &&
    new Set(value).size === value.length &&
    value.every(
      (item) =>
        typeof item === "string" && item.length >= 2 && item.length <= 10,
    )
  );
}
function isError(
  value: unknown,
  code: string,
  requestID: string,
): value is ErrorResponse {
  return (
    isExact(value, ["error"]) &&
    isExact((value as { error: unknown }).error, [
      "code",
      "message",
      "requestId",
    ]) &&
    (value as ErrorResponse).error.code === code &&
    (value as ErrorResponse).error.requestId === requestID
  );
}
function isExact(
  value: unknown,
  expected: string[],
): value is Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value))
    return false;
  const keys = Object.keys(value).sort(),
    wanted = [...expected].sort();
  return (
    keys.length === wanted.length &&
    keys.every((key, index) => key === wanted[index])
  );
}
function isUUID(value: string): boolean {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(
    value,
  );
}
function upstreamHeaders(token: string, requestID: string): HeadersInit {
  return { Authorization: `Bearer ${token}`, [requestIDHeader]: requestID };
}
function readRequestID(headers: Headers): string {
  const value = headers.get(requestIDHeader);
  return value && /^[A-Za-z0-9._:-]{8,128}$/.test(value)
    ? value
    : `req_${crypto.randomUUID()}`;
}
function unavailable(requestID: string): Response {
  return errorResponse(
    "SERVICE_UNAVAILABLE",
    "Service unavailable",
    503,
    requestID,
  );
}
function errorResponse(
  code: ErrorResponse["error"]["code"],
  message: string,
  status: number,
  requestID: string,
): Response {
  return Response.json(
    { error: { code, message, requestId: requestID } } satisfies ErrorResponse,
    { status, headers: { [requestIDHeader]: requestID } },
  );
}
