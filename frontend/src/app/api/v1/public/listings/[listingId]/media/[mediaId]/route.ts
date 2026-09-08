import { mediaProxy } from "@/features/listing-media/media-proxy";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export async function GET(request: Request, context: { params: Promise<{ listingId: string; mediaId: string }> }): Promise<Response> {
  return mediaProxy(request, await context.params, "public", "image");
}
