import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
const mocks = vi.hoisted(() => ({ auth: vi.fn() }));
vi.mock("@clerk/nextjs/server", () => ({ auth: mocks.auth }));
import { GET, POST } from "./route";
const listing = {
  id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  categoryId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  primaryLocalityId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
  title: "Listing test",
  description: "Synthetic moderator BFF response description.",
  priceType: "fixed",
  priceMinor: 5000,
  currency: "EUR",
  travelsToCustomer: true,
  receivesCustomer: false,
  remoteServices: false,
  state: "pending_review",
  revision: 1,
  createdAt: "2026-08-24T12:00:00Z",
  updatedAt: "2026-08-24T12:00:00Z",
};
describe("moderation listings BFF", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    mocks.auth.mockReset();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });
  it("uses server token and preserves forbidden", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async (req: Request) => {
        expect(req.headers.get("Authorization")).toBe("Bearer server-token");
        if (req.url.endsWith("/approve"))
          return Response.json(
            {
              error: {
                code: "FORBIDDEN",
                message: "Forbidden",
                requestId: "req_moderation",
              },
            },
            { status: 403, headers: { "X-Request-ID": "req_moderation" } },
          );
        return Response.json(
          { listings: [listing] },
          { headers: { "X-Request-ID": "req_moderation" } },
        );
      }),
    );
    const queue = await GET(
      new Request("http://localhost/api/v1/moderation/listings", {
        headers: { "X-Request-ID": "req_moderation" },
      }),
    );
    expect(queue.status).toBe(200);
    const approve = await POST(
      new Request(
        "http://localhost/api/v1/moderation/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/approve",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-Request-ID": "req_moderation",
          },
          body: `{"revision":1}`,
        },
      ),
    );
    expect(approve.status).toBe(403);
  });
  it("does not call upstream signed out", async () => {
    mocks.auth.mockResolvedValue({ isAuthenticated: false, getToken: vi.fn() });
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    const r = await GET(
      new Request("http://localhost/api/v1/moderation/listings", {
        headers: { "X-Request-ID": "req_moderation_out" },
      }),
    );
    expect(r.status).toBe(401);
    expect(fetch).not.toHaveBeenCalled();
  });
});
