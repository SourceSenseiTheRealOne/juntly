import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { GET } from "./route";

describe("GET /api/v1/catalog/categories", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("forwards only locale and correlation through the generated client", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (request: Request) => {
        expect(request.url).toBe(
          "http://go-api:8080/api/v1/catalog/categories?locale=pt-PT",
        );
        expect(request.headers.get("Authorization")).toBeNull();
        expect(request.headers.get("X-Request-ID")).toBe("req_categories_ok");
        return Response.json(
          {
            categories: [
              {
                id: "aaaaaaaa-aaaa-daaa-0aaa-aaaaaaaaaaaa",
                parentId: null,
                slug: "home-repairs",
                name: "Reparações domésticas",
              },
            ],
          },
          { headers: { "X-Request-ID": "req_categories_ok" } },
        );
      }),
    );

    const response = await GET(
      new Request("http://localhost/api/v1/catalog/categories?locale=pt-PT", {
        headers: { "X-Request-ID": "req_categories_ok" },
      }),
    );

    expect(response.status).toBe(200);
    await expect(response.json()).resolves.toMatchObject({
      categories: [{ slug: "home-repairs" }],
    });
  });

  it.each([
    "http://localhost/api/v1/catalog/categories",
    "http://localhost/api/v1/catalog/categories?locale=fr",
    "http://localhost/api/v1/catalog/categories?locale=pt-PT&admin=true",
    "http://localhost/api/v1/catalog/categories?locale=pt-PT&locale=en",
  ])("rejects invalid query %s before upstream", async (url) => {
    const upstream = vi.fn();
    vi.stubGlobal("fetch", upstream);
    const response = await GET(
      new Request(url, { headers: { "X-Request-ID": "req_categories_bad" } }),
    );
    expect(response.status).toBe(400);
    expect(upstream).not.toHaveBeenCalled();
  });

  it("maps malformed or mismatched upstream output to safe 503", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        Response.json(
          { categories: [], internalUserId: "private" },
          { headers: { "X-Request-ID": "different" } },
        ),
      ),
    );
    const response = await GET(
      new Request("http://localhost/api/v1/catalog/categories?locale=en", {
        headers: { "X-Request-ID": "req_categories_fail" },
      }),
    );
    expect(response.status).toBe(503);
    const body = await response.json();
    expect(JSON.stringify(body)).not.toContain("go-api");
    expect(JSON.stringify(body)).not.toContain("internalUserId");
  });
});
