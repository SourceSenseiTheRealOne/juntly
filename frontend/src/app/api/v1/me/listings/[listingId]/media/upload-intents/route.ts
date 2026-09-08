import { mediaProxy } from "@/features/listing-media/media-proxy";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export async function POST(
  request: Request,
  context: { params: Promise<{ listingId: string }> },
): Promise<Response> {
  return mediaProxy(request, await context.params, "me", "upload");
}
