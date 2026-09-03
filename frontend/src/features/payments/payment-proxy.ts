import type { ErrorResponse } from "@/shared/api/generated";
import {
  authorizedHeaders,
  correlated,
  correlatedError,
  requestIDHeader,
  unavailable,
} from "@/features/messaging/protected-bff";

export async function forwardPayment(
  requestID: string,
  token: string,
  path: string,
  method: "GET" | "POST",
  body: string | undefined,
  valid: (value: unknown) => boolean,
  successStatus: number,
): Promise<Response> {
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(requestID);
  try {
    const headers = authorizedHeaders(token, requestID);
    if (body !== undefined) headers["Content-Type"] = "application/json";
    const upstream = await fetch(`${origin}${path}`, {
      method,
      headers,
      ...(body === undefined ? {} : { body }),
    });
    const value: unknown = await upstream.json();
    if (correlated(upstream, requestID) && upstream.ok && valid(value)) {
      return Response.json(value, {
        status: successStatus,
        headers: { [requestIDHeader]: requestID },
      });
    }
    const errors: Array<[number, ErrorResponse["error"]["code"]]> = [
      [400, "INVALID_REQUEST"],
      [401, "UNAUTHORIZED"],
      [403, "FORBIDDEN"],
      [404, "NOT_FOUND"],
      [409, "CONFLICT"],
      [503, "SERVICE_UNAVAILABLE"],
    ];
    for (const [status, code] of errors) {
      if (
        upstream.status === status &&
        correlated(upstream, requestID) &&
        correlatedError(value, code, requestID)
      ) {
        return Response.json(value, {
          status,
          headers: { [requestIDHeader]: requestID },
        });
      }
    }
    return unavailable(requestID);
  } catch {
    return unavailable(requestID);
  }
}
