import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET } from "./route";

describe("GET /api/v1/discovery/listings", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("forwards only bounded public query and validates the closed projection", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (request: Request) => {
        expect(request.headers.get("Authorization")).toBeNull();
        expect(request.url).toContain("locale=pt-PT");
        expect(request.url).toContain("q=plumbing");
        return Response.json(
          {
            listings: [
              {
                id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
                title: "Public plumbing",
                description:
                  "Public plumbing listing with enough descriptive text.",
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
                promoted: false,
                updatedAt: "2026-08-24T12:00:00Z",
              },
            ],
          },
          { headers: { "X-Request-ID": "req_public_discovery" } },
        );
      }),
    );
    const response = await GET(
      new Request(
        "http://localhost/api/v1/discovery/listings?locale=pt-PT&q=plumbing",
        { headers: { "X-Request-ID": "req_public_discovery" } },
      ),
    );
    expect(response.status).toBe(200);
    const body = await response.json();
    expect(body.listings).toHaveLength(1);
    expect(JSON.stringify(body)).not.toMatch(
      /internalUserId|clerkSubject|objectReference|latitude|longitude|bio|reason/,
    );
  });

  it("rejects unauthorized query keys before upstream", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const response = await GET(
      new Request(
        "http://localhost/api/v1/discovery/listings?locale=pt-PT&admin=true",
        { headers: { "X-Request-ID": "req_public_bad" } },
      ),
    );
    expect(response.status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
