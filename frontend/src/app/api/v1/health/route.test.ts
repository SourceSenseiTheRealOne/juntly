import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";

import { GET } from "./route";

describe("GET /api/v1/health BFF route", () => {
  beforeEach(() => {
    vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("forwards the correlation ID upstream and returns matching evidence", async () => {
    const upstreamFetch = vi.fn(async (request: Request) => {
      expect(request.url).toBe("http://go-api:8080/api/v1/health");
      expect(request.headers.get("X-Request-ID")).toBe("req_browser_123");

      return Response.json(
        {
          status: "ok",
          service: "juntly-api",
          version: "0.1.0",
          checkedAt: "2026-08-20T09:30:00Z",
          requestId: "req_browser_123",
        },
        {
          headers: {
            "X-Request-ID": "req_browser_123",
          },
        },
      );
    });
    vi.stubGlobal("fetch", upstreamFetch);

    const response = await GET(
      new Request("http://localhost/api/v1/health", {
        headers: {
          "X-Request-ID": "req_browser_123",
        },
      }),
    );

    expect(response.status).toBe(200);
    expect(response.headers.get("X-Request-ID")).toBe("req_browser_123");
    await expect(response.json()).resolves.toEqual({
      status: "ok",
      service: "juntly-api",
      version: "0.1.0",
      checkedAt: "2026-08-20T09:30:00Z",
      requestId: "req_browser_123",
    });
    expect(upstreamFetch).toHaveBeenCalledOnce();
  });

  it("maps unavailable upstream errors to a privacy-safe 503", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new Error(
          "connect ECONNREFUSED http://go-api:8080/api/v1/health",
        );
      }),
    );

    const response = await GET(
      new Request("http://localhost/api/v1/health", {
        headers: {
          "X-Request-ID": "req_browser_503",
        },
      }),
    );

    expect(response.status).toBe(503);
    expect(response.headers.get("X-Request-ID")).toBe("req_browser_503");
    const body = await response.json();
    expect(body).toEqual({
      error: {
        code: "SERVICE_UNAVAILABLE",
        message: "Service unavailable",
        requestId: "req_browser_503",
      },
    });
    expect(JSON.stringify(body)).not.toContain("go-api");
    expect(JSON.stringify(body)).not.toContain("ECONNREFUSED");
  });

  it("maps malformed upstream success to the same stable 503", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        Response.json(
          {
            status: "ok",
            service: "juntly-api",
            version: "0.1.0",
            checkedAt: "2026-08-20T09:30:00Z",
            requestId: "different_request",
          },
          {
            headers: {
              "X-Request-ID": "different_request",
            },
          },
        ),
      ),
    );

    const response = await GET(
      new Request("http://localhost/api/v1/health", {
        headers: {
          "X-Request-ID": "req_browser_malformed",
        },
      }),
    );

    expect(response.status).toBe(503);
    expect(response.headers.get("X-Request-ID")).toBe("req_browser_malformed");
    await expect(response.json()).resolves.toEqual({
      error: {
        code: "SERVICE_UNAVAILABLE",
        message: "Service unavailable",
        requestId: "req_browser_malformed",
      },
    });
  });
});
