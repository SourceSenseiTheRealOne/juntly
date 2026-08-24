import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({ auth: vi.fn() }));
vi.mock("@clerk/nextjs/server", () => ({ auth: mocks.auth }));

import { GET, PUT } from "./route";

describe("/api/v1/me/contact-channels BFF", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    mocks.auth.mockReset();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("gets status-only channels with a server token", async () => {
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
        return Response.json(
          {
            channels: [
              {
                channel: "phone",
                configured: true,
                enabled: true,
                revealConsent: true,
              },
            ],
          },
          { headers: { "X-Request-ID": "req_contact_channels" } },
        );
      }),
    );
    const response = await GET(
      new Request("http://localhost/api/v1/me/contact-channels", {
        headers: { "X-Request-ID": "req_contact_channels" },
      }),
    );
    expect(response.status).toBe(200);
    const body = await response.json();
    expect(body.channels).toHaveLength(1);
    expect(JSON.stringify(body)).not.toMatch(
      /contact|ciphertext|nonce|keyVersion/,
    );
  });

  it("forwards a strict encrypted-channel update but never echoes contact", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async (request: Request) => {
        await expect(request.json()).resolves.toEqual({
          channel: "whatsapp",
          contact: "+12025550123",
          enabled: true,
          revealConsent: true,
        });
        return Response.json(
          {
            channel: "whatsapp",
            configured: true,
            enabled: true,
            revealConsent: true,
          },
          { headers: { "X-Request-ID": "req_contact_put" } },
        );
      }),
    );
    const response = await PUT(
      new Request("http://localhost/api/v1/me/contact-channels", {
        method: "PUT",
        headers: { "X-Request-ID": "req_contact_put" },
        body: JSON.stringify({
          channel: "whatsapp",
          contact: "+12025550123",
          enabled: true,
          revealConsent: true,
        }),
      }),
    );
    expect(response.status).toBe(200);
    expect(JSON.stringify(await response.json())).not.toContain("contact");
  });

  it("rejects unknown owner channel payload before upstream", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const response = await PUT(
      new Request("http://localhost/api/v1/me/contact-channels", {
        method: "PUT",
        body: `{"channel":"phone","contact":"+12025550123","enabled":true,"revealConsent":true,"admin":true}`,
      }),
    );
    expect(response.status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
