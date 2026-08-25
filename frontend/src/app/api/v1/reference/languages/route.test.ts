import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET } from "./route";

describe("GET /api/v1/reference/languages", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("returns exact localized languages without authorization", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (request: Request) => {
        expect(request.url).toBe(
          "http://go-api:8080/api/v1/reference/languages?locale=en",
        );
        expect(request.headers.get("Authorization")).toBeNull();
        return Response.json(
          { languages: [{ code: "pt-PT", name: "Portuguese" }] },
          { headers: { "X-Request-ID": "req_languages_ok" } },
        );
      }),
    );
    const response = await GET(
      new Request("http://localhost/api/v1/reference/languages?locale=en", {
        headers: { "X-Request-ID": "req_languages_ok" },
      }),
    );
    expect(response.status).toBe(200);
    await expect(response.json()).resolves.toEqual({
      languages: [{ code: "pt-PT", name: "Portuguese" }],
    });
  });

  it("rejects unsupported or expanded queries before upstream", async () => {
    for (const url of [
      "http://localhost/api/v1/reference/languages?locale=fr",
      "http://localhost/api/v1/reference/languages?locale=en&admin=true",
    ]) {
      const upstream = vi.fn();
      vi.stubGlobal("fetch", upstream);
      const response = await GET(
        new Request(url, { headers: { "X-Request-ID": "req_languages_bad" } }),
      );
      expect(response.status).toBe(400);
      expect(upstream).not.toHaveBeenCalled();
    }
  });
});
