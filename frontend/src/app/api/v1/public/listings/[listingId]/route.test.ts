import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET } from "./route";

const listing = {
  id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  title: "Public plumbing",
  description: "Public plumbing listing with enough descriptive text.",
  categoryId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  categorySlug: "plumbing",
  categoryName: "Canalização",
  primaryLocalityId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
  localitySlug: "zebreira",
  localityName: "Zebreira",
  priceType: "fixed",
  priceMinor: 5000,
  currency: "EUR",
  travelsToCustomer: true,
  receivesCustomer: false,
  remoteServices: false,
  providerDisplayName: "Public provider",
  providerType: "professional",
  updatedAt: "2026-08-24T12:00:00Z",
};

describe("GET /api/v1/public/listings/[listingId]", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("returns a correlated closed public projection without authorization", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (request: Request) => {
        expect(request.headers.get("Authorization")).toBeNull();
        return Response.json(listing, {
          headers: { "X-Request-ID": "req_public_detail" },
        });
      }),
    );
    const response = await GET(
      new Request(
        "http://localhost/api/v1/public/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa?locale=pt-PT",
        { headers: { "X-Request-ID": "req_public_detail" } },
      ),
      { params: Promise.resolve({ listingId: listing.id }) },
    );
    expect(response.status).toBe(200);
    await expect(response.json()).resolves.toMatchObject({ id: listing.id });
  });

  it("preserves a correlated not found response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        Response.json(
          {
            error: {
              code: "NOT_FOUND",
              message: "Not found",
              requestId: "req_public_missing",
            },
          },
          { status: 404, headers: { "X-Request-ID": "req_public_missing" } },
        ),
      ),
    );
    const response = await GET(
      new Request(
        "http://localhost/api/v1/public/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa?locale=en",
        { headers: { "X-Request-ID": "req_public_missing" } },
      ),
      { params: Promise.resolve({ listingId: listing.id }) },
    );
    expect(response.status).toBe(404);
    await expect(response.json()).resolves.toMatchObject({
      error: { code: "NOT_FOUND" },
    });
  });
});
