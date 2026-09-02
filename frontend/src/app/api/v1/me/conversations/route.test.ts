import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({ auth: vi.fn() }));
vi.mock("@clerk/nextjs/server", () => ({ auth: mocks.auth }));

import { POST } from "./route";

describe("start conversation BFF", () => {
  beforeEach(() => {
    vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080");
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
  });
  afterEach(() => {
    mocks.auth.mockReset();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("preserves upstream forbidden instead of masking it as unavailable", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        Response.json(
          {
            error: {
              code: "FORBIDDEN",
              message: "Forbidden",
              requestId: "req_conversation_owner",
            },
          },
          {
            status: 403,
            headers: { "X-Request-ID": "req_conversation_owner" },
          },
        ),
      ),
    );

    const response = await POST(
      new Request("http://localhost/api/v1/me/conversations", {
        method: "POST",
        headers: { "X-Request-ID": "req_conversation_owner" },
        body: JSON.stringify({
          listingId: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
        }),
      }),
    );

    expect(response.status).toBe(403);
    await expect(response.json()).resolves.toMatchObject({
      error: { code: "FORBIDDEN", requestId: "req_conversation_owner" },
    });
  });
});
