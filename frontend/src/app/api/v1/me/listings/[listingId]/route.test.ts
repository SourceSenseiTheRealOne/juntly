import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
const mocks = vi.hoisted(() => ({ auth: vi.fn() }));
vi.mock("@clerk/nextjs/server", () => ({ auth: mocks.auth }));
import { POST, PUT } from "./route";
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
describe("listing item BFF", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    mocks.auth.mockReset();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });
  it("preserves correlated conflict and hides upload internals", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async (req: Request) => {
        expect(req.headers.get("Authorization")).toBe("Bearer server-token");
        if (req.url.endsWith("/submit"))
          return Response.json(
            {
              error: {
                code: "CONFLICT",
                message: "Conflict",
                requestId: "req_listing_item",
              },
            },
            { status: 409, headers: { "X-Request-ID": "req_listing_item" } },
          );
        return Response.json(
          {
            mediaId: "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
            capability: {
              url: "https://upload.example.invalid/c",
              method: "PUT",
              headers: { "Content-Type": "image/webp" },
            },
          },
          { headers: { "X-Request-ID": "req_listing_item" } },
        );
      }),
    );
    const submit = await POST(
      new Request(
        "http://localhost/api/v1/me/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/submit",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-Request-ID": "req_listing_item",
          },
          body: `{"revision":1}`,
        },
      ),
    );
    expect(submit.status).toBe(409);
    const upload = await POST(
      new Request(
        "http://localhost/api/v1/me/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/media/upload-intents",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-Request-ID": "req_listing_item",
          },
          body: `{"ordinal":1,"contentType":"image/webp","byteSize":1024,"checksumSha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
        },
      ),
    );
    expect(upload.status).toBe(200);
    expect(JSON.stringify(await upload.json())).not.toContain(
      "objectReference",
    );
  });
  it("rejects unknown replace fields before upstream", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    const r = await PUT(
      new Request(
        "http://localhost/api/v1/me/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            "X-Request-ID": "req_listing_replace",
          },
          body: JSON.stringify({ ...listing, ownerId: "forbidden" }),
        },
      ),
    );
    expect(r.status).toBe(400);
    expect(fetch).not.toHaveBeenCalled();
  });
});
