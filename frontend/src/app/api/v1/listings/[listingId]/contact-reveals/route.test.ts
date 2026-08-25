import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({ auth: vi.fn() }));
vi.mock("@clerk/nextjs/server", () => ({ auth: mocks.auth }));
import { POST } from "./route";

describe("/api/v1/listings/[listingId]/contact-reveals BFF", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    mocks.auth.mockReset();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });
  it("reveals contact only after server-side token forwarding", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async (request: Request) => {
        expect(request.headers.get("Authorization")).toBe(
          "Bearer server-token",
        );
        await expect(request.json()).resolves.toEqual({ channel: "phone" });
        return Response.json(
          { channel: "phone", contact: "+12025550123" },
          { headers: { "X-Request-ID": "req_reveal" } },
        );
      }),
    );
    const response = await POST(
      new Request(
        "http://localhost/api/v1/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/contact-reveals",
        {
          method: "POST",
          headers: { "X-Request-ID": "req_reveal" },
          body: `{"channel":"phone"}`,
        },
      ),
      {
        params: Promise.resolve({
          listingId: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
        }),
      },
    );
    expect(response.status).toBe(200);
    await expect(response.json()).resolves.toEqual({
      channel: "phone",
      contact: "+12025550123",
    });
  });
  it("rejects invalid body before upstream", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const response = await POST(
      new Request(
        "http://localhost/api/v1/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/contact-reveals",
        { method: "POST", body: `{"channel":"phone","extra":true}` },
      ),
      {
        params: Promise.resolve({
          listingId: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
        }),
      },
    );
    expect(response.status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
