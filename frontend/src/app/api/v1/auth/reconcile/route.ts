import { auth } from "@clerk/nextjs/server";

import { reconcileInternalUser } from "@/shared/api/generated";
import type {
  ErrorResponse,
  InternalUserResponse,
} from "@/shared/api/generated";

const requestIDHeader = "X-Request-ID";
const unavailableMessage = "Service unavailable";
const unauthorizedMessage = "Unauthorized";

export const runtime = "nodejs";

export async function POST(request: Request): Promise<Response> {
  const requestID = readRequestID(request.headers);
  const { isAuthenticated, getToken } = await auth();

  if (!isAuthenticated) {
    return unauthorizedResponse(requestID);
  }

  const token = await getToken();
  if (!token) {
    return unauthorizedResponse(requestID);
  }

  const apiOrigin = process.env.JUNTLY_API_ORIGIN;
  if (!apiOrigin) {
    return unavailableResponse(requestID);
  }

  try {
    const upstream = await reconcileInternalUser({
      baseUrl: apiOrigin,
      headers: {
        Authorization: `Bearer ${token}`,
        [requestIDHeader]: requestID,
      },
    });

    if (
      upstream.error ||
      !upstream.response?.ok ||
      upstream.response.headers.get(requestIDHeader) !== requestID ||
      !isInternalUserResponse(upstream.data)
    ) {
      return unavailableResponse(requestID);
    }

    return Response.json(upstream.data, {
      status: httpStatusOK,
      headers: {
        [requestIDHeader]: requestID,
      },
    });
  } catch {
    return unavailableResponse(requestID);
  }
}

const httpStatusOK = 200;

function unauthorizedResponse(requestID: string): Response {
  return errorResponse("UNAUTHORIZED", unauthorizedMessage, 401, requestID);
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
    error: {
      code,
      message,
      requestId: requestID,
    },
  };

  return Response.json(body, {
    status,
    headers: {
      [requestIDHeader]: requestID,
    },
  });
}

function readRequestID(headers: Headers): string {
  const requestID = headers.get(requestIDHeader);

  if (requestID && isValidRequestID(requestID)) {
    return requestID;
  }

  return `req_${globalThis.crypto.randomUUID()}`;
}

function isValidRequestID(value: string): boolean {
  return /^[A-Za-z0-9._:-]{8,128}$/.test(value);
}

function isInternalUserResponse(
  value: InternalUserResponse | undefined,
): value is InternalUserResponse {
  return (
    typeof value?.id === "string" &&
    isUUID(value.id) &&
    typeof value.createdAt === "string" &&
    !Number.isNaN(Date.parse(value.createdAt))
  );
}

function isUUID(value: string): boolean {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(
    value,
  );
}
