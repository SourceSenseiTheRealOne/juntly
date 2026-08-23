import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({ auth: vi.fn() }));
vi.mock("@clerk/nextjs/server", () => ({ auth: mocks.auth }));

import { GET, PUT } from "./route";

const profile = {
  displayName: "Prestador local",
  providerType: "individual" as const,
  bio: "Serviço de confiança.",
  primaryLocalityId: "11111111-1111-4111-8111-111111111111",
  serviceLocalityIds: ["11111111-1111-4111-8111-111111111111"],
  maxTravelDistanceKm: 25,
  travelsToCustomer: true,
  receivesCustomer: false,
  remoteServices: false,
  languageCodes: ["pt-PT"],
};

describe("/api/v1/me/provider-profile", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    mocks.auth.mockReset();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("returns 401 signed out without upstream", async () => {
    mocks.auth.mockResolvedValue({ isAuthenticated: false, getToken: vi.fn() });
    const upstream = vi.fn();
    vi.stubGlobal("fetch", upstream);
    const response = await GET(
      new Request("http://localhost/api/v1/me/provider-profile", {
        headers: { "X-Request-ID": "req_profile_signed_out" },
      }),
    );
    expect(response.status).toBe(401);
    expect(upstream).not.toHaveBeenCalled();
  });

  it("forwards server bearer and validates owner-only GET response", async () => {
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
            profile: {
              ...profile,
              createdAt: "2026-08-23T16:00:00Z",
              updatedAt: "2026-08-23T16:00:00Z",
            },
          },
          { headers: { "X-Request-ID": "req_profile_get" } },
        );
      }),
    );
    const response = await GET(
      new Request("http://localhost/api/v1/me/provider-profile", {
        headers: { "X-Request-ID": "req_profile_get" },
      }),
    );
    expect(response.status).toBe(200);
    await expect(response.json()).resolves.toMatchObject({
      profile: { providerType: "individual" },
    });
  });

  it("PUT rejects unknown or incomplete browser data before upstream", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    for (const body of [{}, { ...profile, admin: true }]) {
      const upstream = vi.fn();
      vi.stubGlobal("fetch", upstream);
      const response = await PUT(
        new Request("http://localhost/api/v1/me/provider-profile", {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            "X-Request-ID": "req_profile_invalid",
          },
          body: JSON.stringify(body),
        }),
      );
      expect(response.status).toBe(400);
      expect(upstream).not.toHaveBeenCalled();
    }
  });

  it("PUT forwards exact validated profile and preserves 403", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async (request: Request) => {
        await expect(request.json()).resolves.toEqual(profile);
        return Response.json(
          {
            error: {
              code: "FORBIDDEN",
              message: "Forbidden",
              requestId: "req_profile_forbidden",
            },
          },
          {
            status: 403,
            headers: { "X-Request-ID": "req_profile_forbidden" },
          },
        );
      }),
    );
    const response = await PUT(
      new Request("http://localhost/api/v1/me/provider-profile", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "X-Request-ID": "req_profile_forbidden",
        },
        body: JSON.stringify(profile),
      }),
    );
    expect(response.status).toBe(403);
    await expect(response.json()).resolves.toMatchObject({
      error: { code: "FORBIDDEN" },
    });
  });
});
