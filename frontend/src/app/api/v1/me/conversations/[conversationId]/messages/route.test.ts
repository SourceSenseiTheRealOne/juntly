import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({ auth: vi.fn() }));
vi.mock("@clerk/nextjs/server", () => ({ auth: mocks.auth }));

import { POST } from "./route";

describe("conversation messages BFF", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    mocks.auth.mockReset();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("sends a private message with the server-side session token", async () => {
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
        await expect(request.json()).resolves.toEqual({ body: "Boa tarde" });
        return Response.json(
          {
            id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
            conversationId: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
            senderId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
            body: "Boa tarde",
            createdAt: "2026-09-01T12:00:00Z",
          },
          { status: 201, headers: { "X-Request-ID": "req_messages_bff" } },
        );
      }),
    );

    const response = await POST(
      new Request("http://localhost/api/v1/me/conversations/id/messages", {
        method: "POST",
        headers: { "X-Request-ID": "req_messages_bff" },
        body: JSON.stringify({ body: "Boa tarde" }),
      }),
      {
        params: Promise.resolve({
          conversationId: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
        }),
      },
    );

    expect(response.status).toBe(201);
    await expect(response.json()).resolves.toMatchObject({ body: "Boa tarde" });
  });
});
