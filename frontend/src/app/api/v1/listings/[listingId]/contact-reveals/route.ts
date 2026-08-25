import { auth } from "@clerk/nextjs/server";
import { revealListingContact } from "@/shared/api/generated";
import type {
  ContactRevealRequest,
  ContactRevealResponse,
  ErrorResponse,
} from "@/shared/api/generated";

const header = "X-Request-ID";
export const runtime = "nodejs";
type RouteContext = { params: Promise<{ listingId: string }> };

export async function POST(
  request: Request,
  context: RouteContext,
): Promise<Response> {
  const id = requestID(request.headers);
  const token = await tokenForSession();
  const { listingId } = await context.params;
  const body = await readBody(request);
  if (!token) return error("UNAUTHORIZED", "Unauthorized", 401, id);
  if (!uuid(listingId) || !body)
    return error("INVALID_REQUEST", "Invalid request", 400, id);
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(id);
  try {
    const upstream = await revealListingContact({
      baseUrl: origin,
      path: { listingId },
      body,
      headers: { Authorization: `Bearer ${token}`, [header]: id },
    });
    if (
      upstream.response?.status === 403 &&
      validError(upstream.error, "FORBIDDEN", id) &&
      upstream.response.headers.get(header) === id
    )
      return Response.json(upstream.error, {
        status: 403,
        headers: { [header]: id },
      });
    if (
      upstream.error ||
      !upstream.response?.ok ||
      upstream.response.headers.get(header) !== id ||
      !validReveal(upstream.data)
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

async function tokenForSession(): Promise<string | null> {
  try {
    const state = await auth();
    return state.isAuthenticated ? await state.getToken() : null;
  } catch {
    return null;
  }
}
async function readBody(
  request: Request,
): Promise<ContactRevealRequest | null> {
  try {
    const value: unknown = JSON.parse(await request.text());
    if (
      !exact(value, ["channel"]) ||
      !isChannel((value as Record<string, unknown>).channel)
    )
      return null;
    return { channel: (value as { channel: "phone" | "whatsapp" }).channel };
  } catch {
    return null;
  }
}
function validReveal(
  value: ContactRevealResponse | undefined,
): value is ContactRevealResponse {
  return (
    !!value &&
    exact(value, ["channel", "contact"]) &&
    isChannel(value.channel) &&
    typeof value.contact === "string" &&
    /^\+[1-9][0-9]{7,14}$/.test(value.contact)
  );
}
function validError(
  value: unknown,
  code: string,
  id: string,
): value is ErrorResponse {
  return (
    exact(value, ["error"]) &&
    exact((value as { error: unknown }).error, [
      "code",
      "message",
      "requestId",
    ]) &&
    (value as ErrorResponse).error.code === code &&
    (value as ErrorResponse).error.requestId === id
  );
}
function isChannel(value: unknown): value is "phone" | "whatsapp" {
  return value === "phone" || value === "whatsapp";
}
function exact(
  value: unknown,
  expected: string[],
): value is Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value))
    return false;
  const actual = Object.keys(value).sort(),
    wanted = [...expected].sort();
  return (
    actual.length === wanted.length &&
    actual.every((key, index) => key === wanted[index])
  );
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
