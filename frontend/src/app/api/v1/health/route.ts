import { getHealth } from "@/shared/api/generated";
import type { ErrorResponse, HealthResponse } from "@/shared/api/generated";

const requestIDHeader = "X-Request-ID";
const unavailableMessage = "Service unavailable";

export const runtime = "nodejs";

export async function GET(request: Request): Promise<Response> {
  const requestID = readRequestID(request.headers);
  const apiOrigin = process.env.JUNTLY_API_ORIGIN;

  if (!apiOrigin) {
    return unavailableResponse(requestID);
  }

  try {
    const upstream = await getHealth({
      baseUrl: apiOrigin,
      headers: {
        [requestIDHeader]: requestID,
      },
    });

    if (
      upstream.error ||
      !upstream.response?.ok ||
      !isHealthResponse(upstream.data, requestID) ||
      upstream.response.headers.get(requestIDHeader) !== requestID
    ) {
      return unavailableResponse(requestID);
    }

    return Response.json(upstream.data, {
      status: 200,
      headers: {
        [requestIDHeader]: requestID,
      },
    });
  } catch {
    return unavailableResponse(requestID);
  }
}

function unavailableResponse(requestID: string): Response {
  const body: ErrorResponse = {
    error: {
      code: "SERVICE_UNAVAILABLE",
      message: unavailableMessage,
      requestId: requestID,
    },
  };

  return Response.json(body, {
    status: 503,
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

function isHealthResponse(
  value: HealthResponse | undefined,
  requestID: string,
): value is HealthResponse {
  return (
    value?.status === "ok" &&
    value.service === "juntly-api" &&
    typeof value.version === "string" &&
    value.version.length > 0 &&
    typeof value.checkedAt === "string" &&
    !Number.isNaN(Date.parse(value.checkedAt)) &&
    value.requestId === requestID
  );
}
