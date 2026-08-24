import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({ auth: vi.fn() }));
vi.mock("@clerk/nextjs/server", () => ({ auth: mocks.auth }));
import { GET, POST } from "./route";

const listing = {
  id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  categoryId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  primaryLocalityId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
  title: "Listing test",
  description: "Synthetic listing BFF response description.",
  priceType: "fixed",
  priceMinor: 5000,
  currency: "EUR",
  travelsToCustomer: true,
  receivesCustomer: false,
  remoteServices: false,
  state: "draft",
  revision: 1,
  createdAt: "2026-08-24T12:00:00Z",
  updatedAt: "2026-08-24T12:00:00Z",
};
const create = {
  categoryId: listing.categoryId,
  primaryLocalityId: listing.primaryLocalityId,
  title: listing.title,
  description: listing.description,
  priceType: "fixed",
  priceMinor: 5000,
  currency: "EUR",
  travelsToCustomer: true,
  receivesCustomer: false,
  remoteServices: false,
};

describe("/api/v1/me/listings", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    mocks.auth.mockReset();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });
  it("does not call upstream signed out", async () => {
    mocks.auth.mockResolvedValue({ isAuthenticated: false, getToken: vi.fn() });
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    const r = await GET(
      new Request("http://localhost/api/v1/me/listings", {
        headers: { "X-Request-ID": "req_listings_out" },
      }),
    );
    expect(r.status).toBe(401);
    expect(fetch).not.toHaveBeenCalled();
  });
  it("forwards server bearer and validates collection", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async (req: Request) => {
        expect(req.headers.get("Authorization")).toBe("Bearer server-token");
        return Response.json(
          { listings: [listing] },
          { headers: { "X-Request-ID": "req_listings_get" } },
        );
      }),
    );
    const r = await GET(
      new Request("http://localhost/api/v1/me/listings", {
        headers: { "X-Request-ID": "req_listings_get" },
      }),
    );
    expect(r.status).toBe(200);
    await expect(r.json()).resolves.toEqual({ listings: [listing] });
  });
  it("rejects expanded create before upstream", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    const r = await POST(
      new Request("http://localhost/api/v1/me/listings", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Request-ID": "req_listings_bad",
        },
        body: JSON.stringify({ ...create, ownerId: "forbidden" }),
      }),
    );
    expect(r.status).toBe(400);
    expect(fetch).not.toHaveBeenCalled();
  });
});
