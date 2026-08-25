import { auth } from "@clerk/nextjs/server";
import {
  getContactChannelStatuses,
  replaceContactChannel,
} from "@/shared/api/generated";
import type {
  ContactChannelStatus,
  ErrorResponse,
  ReplaceContactChannelRequest,
} from "@/shared/api/generated";

const header = "X-Request-ID";
export const runtime = "nodejs";

export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers);
  const token = await tokenForSession();
  if (!token) return error("UNAUTHORIZED", "Unauthorized", 401, id);
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(id);
  try {
    const upstream = await getContactChannelStatuses({
      baseUrl: origin,
      headers: headers(token, id),
    });
    if (
      upstream.error ||
      !upstream.response?.ok ||
      upstream.response.headers.get(header) !== id ||
      !validStatuses(upstream.data)
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

export async function PUT(request: Request): Promise<Response> {
  const id = requestID(request.headers);
  const token = await tokenForSession();
  if (!token) return error("UNAUTHORIZED", "Unauthorized", 401, id);
  const body = await readBody(request);
  if (!body) return error("INVALID_REQUEST", "Invalid request", 400, id);
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(id);
  try {
    const upstream = await replaceContactChannel({
      baseUrl: origin,
      body,
      headers: headers(token, id),
    });
    if (
      upstream.error ||
      !upstream.response?.ok ||
      upstream.response.headers.get(header) !== id ||
      !validStatus(upstream.data)
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
): Promise<ReplaceContactChannelRequest | null> {
  try {
    const value: unknown = JSON.parse(await request.text());
    if (!exact(value, ["channel", "contact", "enabled", "revealConsent"]))
      return null;
    const body = value as Record<string, unknown>;
    if (
      !isChannel(body.channel) ||
      typeof body.contact !== "string" ||
      !/^\+[1-9][0-9]{7,14}$/.test(body.contact) ||
      typeof body.enabled !== "boolean" ||
      typeof body.revealConsent !== "boolean"
    )
      return null;
    return {
      channel: body.channel,
      contact: body.contact,
      enabled: body.enabled,
      revealConsent: body.revealConsent,
    };
  } catch {
    return null;
  }
}
function validStatuses(
  value: unknown,
): value is { channels: ContactChannelStatus[] } {
  return (
    exact(value, ["channels"]) &&
    Array.isArray(value.channels) &&
    value.channels.every(validStatus)
  );
}
function validStatus(value: unknown): value is ContactChannelStatus {
  return (
    exact(value, ["channel", "configured", "enabled", "revealConsent"]) &&
    isChannel((value as ContactChannelStatus).channel) &&
    typeof (value as ContactChannelStatus).configured === "boolean" &&
    typeof (value as ContactChannelStatus).enabled === "boolean" &&
    typeof (value as ContactChannelStatus).revealConsent === "boolean"
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
function headers(token: string, id: string): HeadersInit {
  return { Authorization: `Bearer ${token}`, [header]: id };
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
