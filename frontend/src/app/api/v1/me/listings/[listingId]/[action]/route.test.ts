import { afterEach, expect, it, vi } from "vitest";
const mocks = vi.hoisted(() => ({ auth: vi.fn() }));
vi.mock("@clerk/nextjs/server", () => ({ auth: mocks.auth }));
import { POST } from "./route";
afterEach(() => {
  mocks.auth.mockReset();
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
});
it("forwards nested submit action to the owner handler", async () => {
  vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080");
  mocks.auth.mockResolvedValue({
    isAuthenticated: true,
    getToken: vi.fn().mockResolvedValue("server-token"),
  });
  vi.stubGlobal(
    "fetch",
    vi.fn(async () =>
      Response.json(
        {
          id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
          categoryId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
          primaryLocalityId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
          title: "Listing test",
          description: "Synthetic nested route listing description.",
          priceType: "fixed",
          priceMinor: 5000,
          currency: "EUR",
          travelsToCustomer: true,
          receivesCustomer: false,
          remoteServices: false,
          state: "pending_review",
          revision: 2,
          createdAt: "2026-08-24T12:00:00Z",
          updatedAt: "2026-08-24T12:00:00Z",
        },
        { headers: { "X-Request-ID": "req_nested_submit" } },
      ),
    ),
  );
  const r = await POST(
    new Request(
      "http://localhost:4200/api/v1/me/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/submit",
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Request-ID": "req_nested_submit",
        },
        body: `{"revision":1}`,
      },
    ),
  );
  expect(r.status).toBe(200);
});
