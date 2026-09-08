import type { ErrorResponse } from "@/shared/api/generated";
import { resolveSoleAdministratorSession } from "@/features/auth/sole-administrator";
import {
  correlated,
  correlatedError,
  fail,
  requestID,
  requestIDHeader,
  sessionToken,
} from "@/features/messaging/protected-bff";
import {
  maxPhotoBytes,
  mediaID,
  validPhotos,
  validUploadIntent,
  validUploadRequest,
} from "./media-contract";
import type { MediaScope } from "./media-contract";

type Action = "list" | "image" | "upload" | "finalize";
export type MediaParams = { listingId: string; mediaId?: string };
const errorCodes: Record<number, ErrorResponse["error"]["code"]> = {
  400: "INVALID_REQUEST",
  401: "UNAUTHORIZED",
  403: "FORBIDDEN",
  404: "NOT_FOUND",
  409: "CONFLICT",
  503: "SERVICE_UNAVAILABLE",
};
const messages: Record<number, string> = {
  400: "Invalid request",
  401: "Unauthorized",
  403: "Forbidden",
  404: "Not found",
  409: "Conflict",
  503: "Service unavailable",
};

export async function mediaProxy(
  request: Request,
  params: MediaParams,
  scope: MediaScope,
  action: Action,
): Promise<Response> {
  const id = requestID(request.headers);
  const safeHeaders = {
    "Cache-Control": "private, no-store",
    "X-Content-Type-Options": "nosniff",
    "Cross-Origin-Resource-Policy": "same-origin",
    [requestIDHeader]: id,
  };
  const failure = (status: number) => {
    const response = fail(
      errorCodes[status] ?? "SERVICE_UNAVAILABLE",
      messages[status] ?? "Service unavailable",
      status,
      id,
    );
    for (const [key, value] of Object.entries(safeHeaders))
      response.headers.set(key, value);
    return response;
  };
  try {
    if (
      !mediaID(params.listingId) ||
      ((action === "image" || action === "finalize") &&
        !mediaID(params.mediaId)) ||
      new URL(request.url).search
    )
      return failure(400);
    const mutation = action === "upload" || action === "finalize";
    if (
      (mutation && scope !== "me") ||
      request.method !== (mutation ? "POST" : "GET")
    )
      return failure(400);
    if (
      mutation &&
      ((request.headers.has("Origin") &&
        request.headers.get("Origin") !== new URL(request.url).origin) ||
        request.headers.get("Sec-Fetch-Site") === "cross-site")
    )
      return failure(403);
    let token: string | null = null;
    if (scope === "moderation") {
      const admin = await resolveSoleAdministratorSession();
      if (admin.status !== "authorized")
        return failure(
          admin.status === "unauthenticated"
            ? 401
            : admin.status === "forbidden"
              ? 403
              : 503,
        );
      token = admin.token;
    } else if (scope === "me") {
      token = await sessionToken();
      if (!token) return failure(401);
    }
    const origin = process.env.JUNTLY_API_ORIGIN;
    if (!origin || (action === "upload" && !process.env.JUNTLY_STORAGE_ORIGIN))
      return failure(503);
    let body: string | undefined;
    if (mutation) {
      let text: string;
      try {
        text = new TextDecoder().decode(
          await boundedBytes(request.body, 16384),
        );
      } catch {
        return failure(400);
      }
      if (action === "upload") {
        if (
          request.headers.get("Content-Type")?.split(";")[0] !==
          "application/json"
        )
          return failure(400);
        let value: unknown;
        try {
          value = JSON.parse(text);
        } catch {
          return failure(400);
        }
        if (!validUploadRequest(value)) return failure(400);
        body = JSON.stringify(value);
      } else if (text.trim() !== "" && text.trim() !== "{}")
        return failure(400);
    }
    const headers: Record<string, string> = { [requestIDHeader]: id };
    if (token) headers.Authorization = `Bearer ${token}`;
    if (body !== undefined) headers["Content-Type"] = "application/json";
    const suffix =
      action === "upload"
        ? "/upload-intents"
        : action === "image"
          ? `/${params.mediaId}`
          : action === "finalize"
            ? `/${params.mediaId}/finalize`
            : "";
    const upstream = await fetch(
      `${origin}/api/v1/${scope}/listings/${params.listingId}/media${suffix}`,
      {
        method: mutation ? "POST" : "GET",
        headers,
        body,
        cache: "no-store",
        redirect: "error",
        signal: AbortSignal.timeout(20000),
      },
    );
    if (!correlated(upstream, id)) return failure(503);
    if (action === "finalize" && upstream.status === 204)
      return new Response(null, { status: 204, headers: safeHeaders });
    if (action === "image" && upstream.status === 200) {
      if (upstream.headers.get("Content-Type") !== "image/png")
        return failure(503);
      const bytes = await boundedBytes(upstream.body, maxPhotoBytes);
      const signature = [137, 80, 78, 71, 13, 10, 26, 10];
      if (!signature.every((value, index) => bytes[index] === value))
        return failure(503);
      return new Response(bytes.buffer, {
        headers: {
          ...safeHeaders,
          "Content-Type": "image/png",
          "Content-Length": String(bytes.length),
        },
      });
    }
    const bytes = await boundedBytes(upstream.body, 16384);
    const value: unknown = JSON.parse(new TextDecoder().decode(bytes));
    if (
      upstream.status === 200 &&
      ((action === "list" && validPhotos(value)) ||
        (action === "upload" &&
          validUploadIntent(value, process.env.JUNTLY_STORAGE_ORIGIN)))
    )
      return Response.json(value, { headers: safeHeaders });
    const code = errorCodes[upstream.status];
    if (code && correlatedError(value, code, id))
      return failure(upstream.status);
    return failure(503);
  } catch {
    return failure(503);
  }
}

async function boundedBytes(
  stream: ReadableStream<Uint8Array> | null,
  limit: number,
): Promise<Uint8Array<ArrayBuffer>> {
  if (!stream) return new Uint8Array();
  const reader = stream.getReader();
  const chunks: Uint8Array[] = [];
  let length = 0;
  try {
    for (;;) {
      const { done, value } = await reader.read();
      if (done) break;
      length += value.length;
      if (length > limit) {
        await reader.cancel();
        throw new Error("response limit");
      }
      chunks.push(value);
    }
  } finally {
    reader.releaseLock();
  }
  const output = new Uint8Array(length);
  let offset = 0;
  for (const chunk of chunks) {
    output.set(chunk, offset);
    offset += chunk.length;
  }
  return output;
}
