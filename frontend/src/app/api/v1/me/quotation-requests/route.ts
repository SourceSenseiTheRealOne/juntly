import {
  createQuotationRequest,
  listMyQuotationRequests,
  type CreateQuotationRequest,
} from "@/shared/api/generated";
import {
  authorizedHeaders,
  correlated,
  dateTime,
  exact,
  fail,
  requestID,
  requestIDHeader,
  sessionToken,
  unavailable,
  uuid,
} from "@/features/messaging/protected-bff";
import {
  validQuotationRequest,
  validQuotationRequests,
} from "@/features/quotations/quotation-bff";

export const runtime = "nodejs";

export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers);
  const token = await sessionToken();
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const upstream = await listMyQuotationRequests({
      baseUrl,
      headers: authorizedHeaders(token, id),
    });
    if (
      upstream.error ||
      !upstream.response?.ok ||
      !correlated(upstream.response, id) ||
      !validQuotationRequests(upstream.data)
    )
      return unavailable(id);
    return Response.json(upstream.data, {
      status: 200,
      headers: { [requestIDHeader]: id },
    });
  } catch {
    return unavailable(id);
  }
}

export async function POST(request: Request): Promise<Response> {
  const id = requestID(request.headers);
  const token = await sessionToken();
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  const body = await readCreateRequest(request);
  if (!body) return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const upstream = await createQuotationRequest({
      baseUrl,
      body,
      headers: authorizedHeaders(token, id),
    });
    if (
      upstream.error ||
      upstream.response?.status !== 201 ||
      !correlated(upstream.response, id) ||
      !validQuotationRequest(upstream.data)
    )
      return unavailable(id);
    return Response.json(upstream.data, {
      status: 201,
      headers: { [requestIDHeader]: id },
    });
  } catch {
    return unavailable(id);
  }
}

async function readCreateRequest(
  request: Request,
): Promise<CreateQuotationRequest | null> {
  let value: unknown;
  try {
    value = JSON.parse(await request.text());
  } catch {
    return null;
  }
  const validKeys =
    exact(value, [
      "title",
      "description",
      "categoryId",
      "localityId",
      "budgetMinor",
      "proposalDeadline",
    ]) ||
    exact(value, [
      "title",
      "description",
      "categoryId",
      "localityId",
      "proposalDeadline",
    ]);
  if (!validKeys) return null;
  const candidate = value as Record<string, unknown>;
  if (
    typeof candidate.title !== "string" ||
    typeof candidate.description !== "string" ||
    !uuid(candidate.categoryId) ||
    !uuid(candidate.localityId) ||
    !dateTime(candidate.proposalDeadline) ||
    ("budgetMinor" in candidate &&
      candidate.budgetMinor !== null &&
      (!Number.isInteger(candidate.budgetMinor) ||
        Number(candidate.budgetMinor) <= 0))
  )
    return null;
  return {
    title: candidate.title,
    description: candidate.description,
    categoryId: candidate.categoryId,
    localityId: candidate.localityId,
    proposalDeadline: candidate.proposalDeadline,
    ...("budgetMinor" in candidate
      ? { budgetMinor: candidate.budgetMinor as number | null }
      : {}),
  };
}
