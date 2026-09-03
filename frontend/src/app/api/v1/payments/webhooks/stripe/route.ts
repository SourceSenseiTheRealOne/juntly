import {
  exact,
  requestID,
  requestIDHeader,
  unavailable,
} from "@/features/messaging/protected-bff";

export const runtime = "nodejs";

export async function POST(request: Request): Promise<Response> {
  const id = requestID(request.headers);
  const origin = process.env.JUNTLY_API_ORIGIN;
  const signature = request.headers.get("Stripe-Signature");
  if (!origin || !signature) return unavailable(id);
  const body = await request.arrayBuffer();
  if (body.byteLength === 0 || body.byteLength > 256 * 1024)
    return unavailable(id);
  try {
    const upstream = await fetch(`${origin}/api/v1/payments/webhooks/stripe`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Stripe-Signature": signature,
        [requestIDHeader]: id,
      },
      body,
    });
    const value: unknown = await upstream.json();
    if (
      upstream.ok &&
      upstream.headers.get(requestIDHeader) === id &&
      exact(value, ["received"]) &&
      value.received === true
    ) {
      return Response.json(value, {
        status: 200,
        headers: { [requestIDHeader]: id },
      });
    }
    return unavailable(id);
  } catch {
    return unavailable(id);
  }
}
