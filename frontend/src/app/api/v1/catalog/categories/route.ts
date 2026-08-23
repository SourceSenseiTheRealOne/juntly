import { listServiceCategories } from "@/shared/api/generated";
import type { CategoriesResponse, ErrorResponse } from "@/shared/api/generated";

const requestIDHeader = "X-Request-ID";
export const runtime = "nodejs";

export async function GET(request: Request): Promise<Response> {
  const requestID = readRequestID(request.headers);
  const query = new URL(request.url).searchParams;
  const locales = query.getAll("locale");
  if (query.size !== 1 || locales.length !== 1 || !isLocale(locales[0])) {
    return errorResponse("INVALID_REQUEST", "Invalid request", 400, requestID);
  }
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(requestID);
  try {
    const upstream = await listServiceCategories({
      baseUrl: origin,
      query: { locale: locales[0] },
      headers: { [requestIDHeader]: requestID },
    });
    if (
      upstream.error ||
      !upstream.response?.ok ||
      upstream.response.headers.get(requestIDHeader) !== requestID ||
      !isCategoriesResponse(upstream.data)
    )
      return unavailable(requestID);
    return Response.json(upstream.data, {
      status: 200,
      headers: { [requestIDHeader]: requestID },
    });
  } catch {
    return unavailable(requestID);
  }
}

function isCategoriesResponse(
  value: CategoriesResponse | undefined,
): value is CategoriesResponse {
  return (
    isExact(value, ["categories"]) &&
    Array.isArray(value.categories) &&
    value.categories.every(
      (category) =>
        isExact(category, ["id", "name", "parentId", "slug"]) &&
        isUUID(category.id) &&
        (category.parentId === null || isUUID(category.parentId)) &&
        typeof category.slug === "string" &&
        category.slug.length > 0 &&
        typeof category.name === "string" &&
        category.name.length > 0,
    )
  );
}

function isExact(
  value: unknown,
  expected: string[],
): value is Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value))
    return false;
  const keys = Object.keys(value).sort();
  const wanted = [...expected].sort();
  return (
    keys.length === wanted.length &&
    keys.every((key, index) => key === wanted[index])
  );
}
function isLocale(value: string | undefined): value is "pt-PT" | "en" | "es" {
  return value === "pt-PT" || value === "en" || value === "es";
}
function isUUID(value: string): boolean {
  return /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(value);
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
