import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET } from "./route";
describe("GET /api/v1/reference/localities", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });
  it("forwards paired radius and preserves attribution", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (request: Request) => {
        expect(request.url).toContain("locale=pt-PT");
        expect(request.url).toContain("radiusKm=25");
        return Response.json(
          {
            localities: [
              {
                id: "11111111-1111-4111-8111-111111111111",
                slug: "zebreira",
                name: "Zebreira",
                parishName: "Zebreira e Segura",
                municipalityName: "Idanha-a-Nova",
                districtName: "Castelo Branco",
                distanceMeters: 0,
              },
            ],
            attribution: {
              text: "© OpenStreetMap contributors",
              url: "https://www.openstreetmap.org/copyright",
            },
          },
          { headers: { "X-Request-ID": "req_localities_ok" } },
        );
      }),
    );
    const r = await GET(
      new Request(
        "http://localhost/api/v1/reference/localities?locale=pt-PT&nearLocalityId=11111111-1111-4111-8111-111111111111&radiusKm=25",
        { headers: { "X-Request-ID": "req_localities_ok" } },
      ),
    );
    expect(r.status).toBe(200);
    const b = await r.json();
    expect(b.attribution.text).toContain("OpenStreetMap");
    expect(JSON.stringify(b)).not.toContain("latitude");
  });
  it("rejects unpaired or unknown query", async () => {
    for (const url of [
      "http://localhost/api/v1/reference/localities?locale=en&radiusKm=10",
      "http://localhost/api/v1/reference/localities?locale=en&admin=true",
    ]) {
      const f = vi.fn();
      vi.stubGlobal("fetch", f);
      const r = await GET(
        new Request(url, { headers: { "X-Request-ID": "req_localities_bad" } }),
      );
      expect(r.status).toBe(400);
      expect(f).not.toHaveBeenCalled();
    }
  });
});
