import {
  listQuotationProposals,
  submitQuotationProposal,
  type SubmitQuotationProposal,
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
  validQuotationProposal,
  validQuotationProposals,
} from "@/features/quotations/quotation-bff";

export const runtime = "nodejs";
type Context = { params: Promise<{ requestId: string }> };

export async function GET(
  request: Request,
  context: Context,
): Promise<Response> {
  const id = requestID(request.headers);
  const token = await sessionToken();
  const { requestId } = await context.params;
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  if (!uuid(requestId))
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const upstream = await listQuotationProposals({
      baseUrl,
      path: { requestId },
      headers: authorizedHeaders(token, id),
    });
    if (
      upstream.error ||
      !upstream.response?.ok ||
      !correlated(upstream.response, id) ||
      !validQuotationProposals(upstream.data)
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

export async function POST(
  request: Request,
  context: Context,
): Promise<Response> {
  const id = requestID(request.headers);
  const token = await sessionToken();
  const { requestId } = await context.params;
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  if (!uuid(requestId))
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const body = await readProposal(request);
  if (!body) return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const upstream = await submitQuotationProposal({
      baseUrl,
      path: { requestId },
      body,
      headers: authorizedHeaders(token, id),
    });
    if (
      upstream.error ||
      upstream.response?.status !== 201 ||
      !correlated(upstream.response, id) ||
      !validQuotationProposal(upstream.data)
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

async function readProposal(
  request: Request,
): Promise<SubmitQuotationProposal | null> {
  let value: unknown;
  try {
    value = JSON.parse(await request.text());
  } catch {
    return null;
  }
  if (value === null || typeof value !== "object" || Array.isArray(value))
    return null;
  const candidate = value as Record<string, unknown>;
  const baseKeys = ["priceMinor", "message", "availableAt"];
  const optionalKeys = ["estimatedMinutes", "expiresAt"];
  if (
    !exact(
      candidate,
      Object.keys(candidate).filter(
        (key) => baseKeys.includes(key) || optionalKeys.includes(key),
      ),
    ) ||
    !baseKeys.every((key) => key in candidate) ||
    !Number.isInteger(candidate.priceMinor) ||
    Number(candidate.priceMinor) <= 0 ||
    typeof candidate.message !== "string" ||
    !dateTime(candidate.availableAt) ||
    ("estimatedMinutes" in candidate &&
      candidate.estimatedMinutes !== null &&
      !Number.isInteger(candidate.estimatedMinutes)) ||
    ("expiresAt" in candidate &&
      candidate.expiresAt !== null &&
      !dateTime(candidate.expiresAt))
  )
    return null;
  return {
    priceMinor: Number(candidate.priceMinor),
    message: candidate.message,
    availableAt: candidate.availableAt,
    ...("estimatedMinutes" in candidate
      ? { estimatedMinutes: candidate.estimatedMinutes as number | null }
      : {}),
    ...("expiresAt" in candidate
      ? { expiresAt: candidate.expiresAt as string | null }
      : {}),
  };
}
