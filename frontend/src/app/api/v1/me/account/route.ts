import { auth } from "@clerk/nextjs/server";

import {
  getAccountCapabilities,
  updateAccountCapabilities,
} from "@/shared/api/generated";
import type {
  AccountCapabilitiesResponse,
  ErrorResponse,
  UpdateAccountCapabilitiesRequest,
} from "@/shared/api/generated";

const requestIDHeader = "X-Request-ID";
const unauthorizedMessage = "Unauthorized";
const invalidRequestMessage = "Invalid request";
const unavailableMessage = "Service unavailable";

export const runtime = "nodejs";

export async function GET(request: Request): Promise<Response> {
  const requestID = readRequestID(request.headers);
  const token = await currentSessionToken();
  if (!token) {
    return unauthorizedResponse(requestID);
  }

  const apiOrigin = process.env.JUNTLY_API_ORIGIN;
  if (!apiOrigin) {
    return unavailableResponse(requestID);
  }

  try {
    const upstream = await getAccountCapabilities({
      baseUrl: apiOrigin,
      headers: upstreamHeaders(token, requestID),
    });
    return accountResponse(upstream, requestID);
  } catch {
    return unavailableResponse(requestID);
  }
}

export async function PUT(request: Request): Promise<Response> {
  const requestID = readRequestID(request.headers);
  const token = await currentSessionToken();
  if (!token) {
    return unauthorizedResponse(requestID);
  }

  const body = await readUpdateRequest(request);
  if (!body) {
    return invalidRequestResponse(requestID);
  }

  const apiOrigin = process.env.JUNTLY_API_ORIGIN;
  if (!apiOrigin) {
    return unavailableResponse(requestID);
  }

  try {
    const upstream = await updateAccountCapabilities({
      baseUrl: apiOrigin,
      body,
      headers: upstreamHeaders(token, requestID),
    });
    return accountResponse(upstream, requestID);
  } catch {
    return unavailableResponse(requestID);
  }
}

async function currentSessionToken(): Promise<string | null> {
  try {
    const { isAuthenticated, getToken } = await auth();
    if (!isAuthenticated) {
      return null;
    }
    return await getToken();
  } catch {
    return null;
  }
}

async function readUpdateRequest(
  request: Request,
): Promise<UpdateAccountCapabilitiesRequest | null> {
  try {
    const value: unknown = JSON.parse(await request.text());
    if (!isExactObject(value, ["providerEnabled"])) {
      return null;
    }
    if (typeof value.providerEnabled !== "boolean") {
      return null;
    }
    return { providerEnabled: value.providerEnabled };
  } catch {
    return null;
  }
}

function upstreamHeaders(token: string, requestID: string): HeadersInit {
  return {
    Authorization: `Bearer ${token}`,
    [requestIDHeader]: requestID,
  };
}

function accountResponse(
  upstream: {
    data?: AccountCapabilitiesResponse;
    error?: unknown;
    response?: Response;
  },
  requestID: string,
): Response {
  if (
    upstream.error ||
    !upstream.response?.ok ||
    upstream.response.headers.get(requestIDHeader) !== requestID ||
    !isAccountCapabilitiesResponse(upstream.data)
  ) {
    return unavailableResponse(requestID);
  }

  return Response.json(upstream.data, {
    status: 200,
    headers: { [requestIDHeader]: requestID },
  });
}

function isAccountCapabilitiesResponse(
  value: AccountCapabilitiesResponse | undefined,
): value is AccountCapabilitiesResponse {
  return (
    isExactObject(value, [
      "customerEnabled",
      "onboardingCompletedAt",
      "providerEnabled",
    ]) &&
    value.customerEnabled === true &&
    typeof value.providerEnabled === "boolean" &&
    typeof value.onboardingCompletedAt === "string" &&
    !Number.isNaN(Date.parse(value.onboardingCompletedAt))
  );
}

function isExactObject(
  value: unknown,
  expectedKeys: readonly string[],
): value is Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    return false;
  }
  const keys = Object.keys(value).sort();
  const sortedExpectedKeys = [...expectedKeys].sort();
  return (
    keys.length === sortedExpectedKeys.length &&
    keys.every((key, index) => key === sortedExpectedKeys[index])
  );
}

function unauthorizedResponse(requestID: string): Response {
  return errorResponse("UNAUTHORIZED", unauthorizedMessage, 401, requestID);
}

function invalidRequestResponse(requestID: string): Response {
  return errorResponse(
    "INVALID_REQUEST",
    invalidRequestMessage,
    400,
    requestID,
  );
}

function unavailableResponse(requestID: string): Response {
  return errorResponse(
    "SERVICE_UNAVAILABLE",
    unavailableMessage,
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
  const body: ErrorResponse = {
    error: { code, message, requestId: requestID },
  };
  return Response.json(body, {
    status,
    headers: { [requestIDHeader]: requestID },
  });
}

function readRequestID(headers: Headers): string {
  const requestID = headers.get(requestIDHeader);
  if (requestID && /^[A-Za-z0-9._:-]{8,128}$/.test(requestID)) {
    return requestID;
  }
  return `req_${globalThis.crypto.randomUUID()}`;
}
