import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
const mocks = vi.hoisted(() => ({ auth: vi.fn(), currentUser: vi.fn() }));
vi.mock("@clerk/nextjs/server", () => ({
  auth: mocks.auth,
  currentUser: mocks.currentUser,
}));
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
    mocks.currentUser.mockReset();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });
  it("uses server token and preserves forbidden", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      userId: "user_admin",
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    mocks.currentUser.mockResolvedValue({
      id: "user_admin",
      emailAddresses: [
        {
          emailAddress: "source.sensei1205@gmail.com",
          verification: { status: "verified" },
        },
      ],
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
    mocks.auth.mockResolvedValue({
      isAuthenticated: false,
      userId: null,
      getToken: vi.fn(),
    });
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

  it("forbids every other authenticated account before the upstream API", async () => {
    const getToken = vi.fn().mockResolvedValue("must-not-be-used");
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      userId: "user_other",
      getToken,
    });
    mocks.currentUser.mockResolvedValue({
      id: "user_other",
      emailAddresses: [
        {
          emailAddress: "someone@example.com",
          verification: { status: "verified" },
        },
      ],
    });
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);

    const response = await GET(
      new Request("http://localhost/api/v1/moderation/listings", {
        headers: { "X-Request-ID": "req_moderation_other" },
      }),
    );

    expect(response.status).toBe(403);
    expect(getToken).not.toHaveBeenCalled();
    expect(fetch).not.toHaveBeenCalled();
  });

  it("forbids an unverified copy of the administrator email", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      userId: "user_unverified",
      getToken: vi.fn(),
    });
    mocks.currentUser.mockResolvedValue({
      id: "user_unverified",
      emailAddresses: [
        {
          emailAddress: "source.sensei1205@gmail.com",
          verification: { status: "unverified" },
        },
      ],
    });
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);

    const response = await GET(
      new Request("http://localhost/api/v1/moderation/listings", {
        headers: { "X-Request-ID": "req_moderation_unverified" },
      }),
    );

    expect(response.status).toBe(403);
    expect(fetch).not.toHaveBeenCalled();
  });
});
