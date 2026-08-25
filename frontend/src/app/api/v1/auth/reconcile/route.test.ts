import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({
  auth: vi.fn(),
}));

vi.mock("@clerk/nextjs/server", () => ({ auth: mocks.auth }));

import { POST } from "./route";

describe("POST /api/v1/auth/reconcile BFF route", () => {
  beforeEach(() => {
    vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080");
  });

  afterEach(() => {
    mocks.auth.mockReset();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("returns 401 without an authenticated Clerk session", async () => {
    mocks.auth.mockResolvedValue({ isAuthenticated: false, getToken: vi.fn() });
    const upstreamFetch = vi.fn();
    vi.stubGlobal("fetch", upstreamFetch);

    const response = await POST(
      new Request("http://localhost/api/v1/auth/reconcile", {
        method: "POST",
        headers: { "X-Request-ID": "req_reconcile_signed_out" },
      }),
    );

    expect(response.status).toBe(401);
    await expect(response.json()).resolves.toEqual({
      error: {
        code: "UNAUTHORIZED",
        message: "Unauthorized",
        requestId: "req_reconcile_signed_out",
      },
    });
    expect(upstreamFetch).not.toHaveBeenCalled();
  });

  it("returns 401 when an authenticated session has no current token", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue(null),
    });
    const upstreamFetch = vi.fn();
    vi.stubGlobal("fetch", upstreamFetch);

    const response = await POST(
      new Request("http://localhost/api/v1/auth/reconcile", {
        method: "POST",
        headers: { "X-Request-ID": "req_reconcile_no_token" },
      }),
    );

    expect(response.status).toBe(401);
    expect(upstreamFetch).not.toHaveBeenCalled();
  });

  it("forwards only a server-obtained bearer token and correlation ID", async () => {
    const getToken = vi.fn().mockResolvedValue("server-obtained-token");
    mocks.auth.mockResolvedValue({ isAuthenticated: true, getToken });
    const upstreamFetch = vi.fn(async (request: Request) => {
      expect(request.url).toBe("http://go-api:8080/api/v1/auth/reconcile");
      expect(request.method).toBe("POST");
      expect(request.headers.get("Authorization")).toBe(
        "Bearer server-obtained-token",
      );
      expect(request.headers.get("X-Request-ID")).toBe("req_reconcile_success");
      await expect(request.text()).resolves.toBe("");

      return Response.json(
        {
          id: "7b7b7d7e-38f9-4f0c-8a10-0fce9cf6f82b",
          createdAt: "2026-08-21T12:00:00.123456Z",
        },
        { headers: { "X-Request-ID": "req_reconcile_success" } },
      );
    });
    vi.stubGlobal("fetch", upstreamFetch);

    const response = await POST(
      new Request("http://localhost/api/v1/auth/reconcile", {
        method: "POST",
        headers: { "X-Request-ID": "req_reconcile_success" },
      }),
    );

    expect(response.status).toBe(200);
    expect(response.headers.get("X-Request-ID")).toBe("req_reconcile_success");
    await expect(response.json()).resolves.toEqual({
      id: "7b7b7d7e-38f9-4f0c-8a10-0fce9cf6f82b",
      createdAt: "2026-08-21T12:00:00.123456Z",
    });
    expect(getToken).toHaveBeenCalledOnce();
  });

  it("maps upstream failure to a topology-safe 503", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-obtained-token"),
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new Error(
          "connect ECONNREFUSED http://go-api:8080/api/v1/auth/reconcile",
        );
      }),
    );

    const response = await POST(
      new Request("http://localhost/api/v1/auth/reconcile", {
        method: "POST",
        headers: { "X-Request-ID": "req_reconcile_unavailable" },
      }),
    );

    expect(response.status).toBe(503);
    const body = await response.json();
    expect(body).toEqual({
      error: {
        code: "SERVICE_UNAVAILABLE",
        message: "Service unavailable",
        requestId: "req_reconcile_unavailable",
      },
    });
    expect(JSON.stringify(body)).not.toContain("go-api");
    expect(JSON.stringify(body)).not.toContain("ECONNREFUSED");
  });

  it("maps malformed upstream response to a topology-safe 503", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      getToken: vi.fn().mockResolvedValue("server-obtained-token"),
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        Response.json(
          { id: "not-a-uuid", createdAt: "not-a-timestamp" },
          { headers: { "X-Request-ID": "req_reconcile_malformed" } },
        ),
      ),
    );

    const response = await POST(
      new Request("http://localhost/api/v1/auth/reconcile", {
        method: "POST",
        headers: { "X-Request-ID": "req_reconcile_malformed" },
      }),
    );

    expect(response.status).toBe(503);
    await expect(response.json()).resolves.toEqual({
      error: {
        code: "SERVICE_UNAVAILABLE",
        message: "Service unavailable",
        requestId: "req_reconcile_malformed",
      },
    });
  });
});
