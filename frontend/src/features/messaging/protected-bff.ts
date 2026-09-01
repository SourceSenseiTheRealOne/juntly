import { auth } from "@clerk/nextjs/server";
import type { ErrorResponse } from "@/shared/api/generated";

export const requestIDHeader = "X-Request-ID";

export async function sessionToken(): Promise<string | null> {
  try {
    const state = await auth();
    return state.isAuthenticated ? await state.getToken() : null;
  } catch {
    return null;
  }
}
export function requestID(headers: Headers): string {
  const value = headers.get(requestIDHeader);
  return value && /^[A-Za-z0-9._:-]{8,128}$/.test(value)
    ? value
    : `req_${crypto.randomUUID()}`;
}
export function exact(value: unknown, expected: string[]): value is Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) return false;
  const actual = Object.keys(value).sort();
  const wanted = [...expected].sort();
  return actual.length === wanted.length && actual.every((key, index) => key === wanted[index]);
}
export function uuid(value: unknown): value is string {
  return typeof value === "string" && /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(value);
}
export function dateTime(value: unknown): value is string {
  return typeof value === "string" && !Number.isNaN(Date.parse(value));
}
export function fail(code: ErrorResponse["error"]["code"], message: string, status: number, id: string): Response {
  return Response.json({ error: { code, message, requestId: id } } satisfies ErrorResponse, { status, headers: { [requestIDHeader]: id } });
}
export function unavailable(id: string): Response { return fail("SERVICE_UNAVAILABLE", "Service unavailable", 503, id); }
export function authorizedHeaders(token: string, id: string): Record<string, string> {
  return { Authorization: `Bearer ${token}`, [requestIDHeader]: id };
}
export function correlated(response: Response | undefined, id: string): boolean {
  return !!response && response.headers.get(requestIDHeader) === id;
}
