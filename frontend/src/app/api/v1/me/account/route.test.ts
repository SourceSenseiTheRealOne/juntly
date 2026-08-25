import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({
  auth: vi.fn(),
}));

vi.mock("@clerk/nextjs/server", () => ({ auth: mocks.auth }));

import { GET, PUT } from "./route";

describe("/api/v1/me/account BFF route", () => {
  beforeEach(() => {
    vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080");
  });

  afterEach(() => {
    mocks.auth.mockReset();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it.each([
    ["GET", GET, undefined],
    ["PUT", PUT, { providerEnabled: true }],
  ])(
    "returns 401 for signed-out %s without calling upstream",
    async (_, handler, body) => {
      mocks.auth.mockResolvedValue({
        isAuthenticated: false,
        getToken: vi.fn(),
      });
      const upstreamFetch = vi.fn();
      vi.stubGlobal("fetch", upstreamFetch);

      const response = await handler(
        request(body, "req_account_signed_out", body ? "PUT" : "GET"),
      );

      expect(response.status).toBe(401);
      await expect(response.json()).resolves.toEqual({
        error: {
          code: "UNAUTHORIZED",
          message: "Unauthorized",
          requestId: "req_account_signed_out",
        },
      });
      expect(upstreamFetch).not.toHaveBeenCalled();
    },
  );

  it("returns 401 when the Clerk session has no current token", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue(null),
    });
    const upstreamFetch = vi.fn();
    vi.stubGlobal("fetch", upstreamFetch);

    const response = await GET(
      request(undefined, "req_account_no_token", "GET"),
    );

    expect(response.status).toBe(401);
    expect(upstreamFetch).not.toHaveBeenCalled();
  });

  it("GET forwards only the server token and correlation ID", async () => {
    const getToken = vi.fn().mockResolvedValue("server-obtained-token");
    mocks.auth.mockResolvedValue({ isAuthenticated: true, getToken });
    vi.stubGlobal(
      "fetch",
      vi.fn(async (upstream: Request) => {
        expect(upstream.url).toBe("http://go-api:8080/api/v1/me/account");
        expect(upstream.method).toBe("GET");
        expect(upstream.headers.get("Authorization")).toBe(
          "Bearer server-obtained-token",
        );
        expect(upstream.headers.get("X-Request-ID")).toBe("req_account_get");
        await expect(upstream.text()).resolves.toBe("");
        return accountResponse(false, "req_account_get");
      }),
    );

    const response = await GET(request(undefined, "req_account_get", "GET"));

    expect(response.status).toBe(200);
    expect(response.headers.get("X-Request-ID")).toBe("req_account_get");
    await expect(response.json()).resolves.toEqual({
      customerEnabled: true,
      providerEnabled: false,
      onboardingCompletedAt: "2026-08-23T12:05:00.123456Z",
    });
    expect(getToken).toHaveBeenCalledOnce();
  });

  it("PUT validates and forwards the exact provider capability body", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-obtained-token"),
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async (upstream: Request) => {
        expect(upstream.url).toBe("http://go-api:8080/api/v1/me/account");
        expect(upstream.method).toBe("PUT");
        expect(upstream.headers.get("Authorization")).toBe(
          "Bearer server-obtained-token",
        );
        expect(upstream.headers.get("X-Request-ID")).toBe("req_account_put");
        await expect(upstream.json()).resolves.toEqual({
          providerEnabled: true,
        });
        return accountResponse(true, "req_account_put");
      }),
    );

    const response = await PUT(
      request({ providerEnabled: true }, "req_account_put", "PUT"),
    );

    expect(response.status).toBe(200);
    await expect(response.json()).resolves.toMatchObject({
      customerEnabled: true,
      providerEnabled: true,
    });
  });

  it.each([
    ["empty object", `{}`],
    ["null", `null`],
    ["unknown property", `{"providerEnabled":true,"admin":true}`],
    ["non-boolean", `{"providerEnabled":"true"}`],
    ["null capability", `{"providerEnabled":null}`],
    ["trailing bytes", `{"providerEnabled":true} trailing`],
    ["second value", `{"providerEnabled":true}{}`],
  ])("PUT rejects %s before upstream", async (_, body) => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-obtained-token"),
    });
    const upstreamFetch = vi.fn();
    vi.stubGlobal("fetch", upstreamFetch);

    const response = await PUT(
      new Request("http://localhost/api/v1/me/account", {
        method: "PUT",
        headers: { "X-Request-ID": "req_account_invalid" },
        body,
      }),
    );

    expect(response.status).toBe(400);
    await expect(response.json()).resolves.toEqual({
      error: {
        code: "INVALID_REQUEST",
        message: "Invalid request",
        requestId: "req_account_invalid",
      },
    });
    expect(upstreamFetch).not.toHaveBeenCalled();
  });

  it.each([
    ["missing origin", "missing-origin"],
    ["network failure", "network"],
    ["malformed response", "malformed"],
    ["extra response property", "extra"],
    ["correlation mismatch", "mismatch"],
  ])("maps %s to a topology-safe 503", async (_, failure) => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-obtained-token"),
    });
    if (failure === "missing-origin") {
      vi.stubEnv("JUNTLY_API_ORIGIN", "");
    }
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        if (failure === "network") {
          throw new Error("connect ECONNREFUSED http://go-api:8080");
        }
        if (failure === "malformed") {
          return Response.json(
            { customerEnabled: false, providerEnabled: "yes" },
            { headers: { "X-Request-ID": "req_account_failure" } },
          );
        }
        if (failure === "extra") {
          return Response.json(
            {
              customerEnabled: true,
              providerEnabled: false,
              onboardingCompletedAt: "2026-08-23T12:05:00Z",
              internalUserId: "not-public",
            },
            { headers: { "X-Request-ID": "req_account_failure" } },
          );
        }
        return accountResponse(false, "req_different_correlation");
      }),
    );

    const response = await GET(
      request(undefined, "req_account_failure", "GET"),
    );

    expect(response.status).toBe(503);
    const body = await response.json();
    expect(body).toEqual({
      error: {
        code: "SERVICE_UNAVAILABLE",
        message: "Service unavailable",
        requestId: "req_account_failure",
      },
    });
    expect(JSON.stringify(body)).not.toContain("go-api");
    expect(JSON.stringify(body)).not.toContain("ECONNREFUSED");
    expect(JSON.stringify(body)).not.toContain("internalUserId");
  });
});

function request(
  body: { providerEnabled: boolean } | undefined,
  requestID: string,
  method: "GET" | "PUT",
): Request {
  return new Request("http://localhost/api/v1/me/account", {
    method,
    headers: { "X-Request-ID": requestID },
    body: body ? JSON.stringify(body) : undefined,
  });
}

function accountResponse(
  providerEnabled: boolean,
  requestID: string,
): Response {
  return Response.json(
    {
      customerEnabled: true,
      providerEnabled,
      onboardingCompletedAt: "2026-08-23T12:05:00.123456Z",
    },
    { headers: { "X-Request-ID": requestID } },
  );
}
