import { listLocalities } from "@/shared/api/generated";
import type { ErrorResponse, LocalitiesResponse } from "@/shared/api/generated";
const header = "X-Request-ID";
export const runtime = "nodejs";
export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    q = new URL(request.url).searchParams;
  if (!keys(q, ["locale", "nearLocalityId", "radiusKm"]))
    return error("INVALID_REQUEST", "Invalid request", 400, id);
  const locales = q.getAll("locale"),
    near = q.getAll("nearLocalityId"),
    radius = q.getAll("radiusKm");
  if (
    locales.length !== 1 ||
    !locale(locales[0]) ||
    near.length > 1 ||
    radius.length > 1 ||
    (near.length === 1) !== (radius.length === 1)
  )
    return error("INVALID_REQUEST", "Invalid request", 400, id);
  const query: {
    locale: "pt-PT" | "en" | "es";
    nearLocalityId?: string;
    radiusKm?: number;
  } = { locale: locales[0] };
  if (near.length) {
    if (
      !uuid(near[0]) ||
      !/^(?:[1-9]|[1-9][0-9]|1[0-9]{2}|200)$/.test(radius[0])
    )
      return error("INVALID_REQUEST", "Invalid request", 400, id);
    query.nearLocalityId = near[0];
    query.radiusKm = Number(radius[0]);
  }
  const origin = process.env.JUNTLY_API_ORIGIN;
  if (!origin) return unavailable(id);
  try {
    const up = await listLocalities({
      baseUrl: origin,
      query,
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
function valid(v: LocalitiesResponse | undefined): v is LocalitiesResponse {
  return (
    exact(v, ["attribution", "localities"]) &&
    exact(v.attribution, ["text", "url"]) &&
    v.attribution.text === "© OpenStreetMap contributors" &&
    v.attribution.url === "https://www.openstreetmap.org/copyright" &&
    Array.isArray(v.localities) &&
    v.localities.every((x) => {
      const expected = [
        "districtName",
        "id",
        "municipalityName",
        "name",
        "parishName",
        "slug",
        ...(x.distanceMeters === undefined ? [] : ["distanceMeters"]),
      ];
      return (
        exact(x, expected) &&
        uuid(x.id) &&
        typeof x.slug === "string" &&
        typeof x.name === "string" &&
        typeof x.parishName === "string" &&
        typeof x.municipalityName === "string" &&
        typeof x.districtName === "string" &&
        (x.distanceMeters === undefined ||
          (Number.isInteger(x.distanceMeters) && x.distanceMeters >= 0))
      );
    })
  );
}
function keys(q: URLSearchParams, allowed: string[]) {
  return [...q.keys()].every((k) => allowed.includes(k));
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
function uuid(v: string) {
  return /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(v);
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
