import { listSpokenLanguages } from "@/shared/api/generated";
import type { ErrorResponse, LanguagesResponse } from "@/shared/api/generated";
const header = "X-Request-ID";
export const runtime = "nodejs";
export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    q = new URL(request.url).searchParams,
    values = q.getAll("locale");
  if (q.size !== 1 || values.length !== 1 || !locale(values[0]))
    return error("INVALID_REQUEST", "Invalid request", 400, id);
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(id);
  try {
    const up = await listSpokenLanguages({
      baseUrl: origin,
      query: { locale: values[0] },
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
function valid(
  value: LanguagesResponse | undefined,
): value is LanguagesResponse {
  return (
    exact(value, ["languages"]) &&
    Array.isArray(value.languages) &&
    value.languages.every(
      (x) =>
        exact(x, ["code", "name"]) &&
        typeof x.code === "string" &&
        x.code.length >= 2 &&
        typeof x.name === "string" &&
        x.name.length > 0,
    )
  );
}
function exact(v: unknown, e: string[]): v is Record<string, unknown> {
  if (v === null || typeof v !== "object" || Array.isArray(v)) return false;
  const k = Object.keys(v).sort(),
    w = [...e].sort();
  return k.length === w.length && k.every((x, i) => x === w[i]);
}
function locale(v: string | undefined): v is "pt-PT" | "en" | "es" {
  return v === "pt-PT" || v === "en" || v === "es";
}
function requestID(h: Headers) {
  const v = h.get(header);
  return v && /^[A-Za-z0-9._:-]{8,128}$/.test(v)
    ? v
    : `req_${crypto.randomUUID()}`;
}
function unavailable(id: string) {
  return error("SERVICE_UNAVAILABLE", "Service unavailable", 503, id);
}
function error(
  code: ErrorResponse["error"]["code"],
  message: string,
  status: number,
  id: string,
) {
  return Response.json(
    { error: { code, message, requestId: id } } satisfies ErrorResponse,
    { status, headers: { [header]: id } },
  );
}
