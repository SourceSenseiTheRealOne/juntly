import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({ auth: vi.fn(), currentUser: vi.fn() }));
vi.mock("@clerk/nextjs/server", () => ({
  auth: mocks.auth,
  currentUser: mocks.currentUser,
}));

import { GET } from "./route";

const order = {
  id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  bookingId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  customerId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
  providerId: "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
  state: "disputed",
  grossMinor: 12500,
  platformFeeMinor: 1250,
  providerNetMinor: 11250,
  currency: "EUR",
  createdAt: "2026-09-02T12:00:00Z",
  updatedAt: "2026-09-02T12:00:00Z",
};

describe("administrative payments BFF", () => {
  beforeEach(() => vi.stubEnv("JUNTLY_API_ORIGIN", "http://go-api:8080"));
  afterEach(() => {
    mocks.auth.mockReset();
    mocks.currentUser.mockReset();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("allows only the exact verified administrator and forwards its server token", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      userId: "user_admin",
      getToken: vi.fn().mockResolvedValue("server-token"),
    });
    mocks.currentUser.mockResolvedValue({
      id: "user_admin",
      emailAddresses: [
        {
          emailAddress: "source.sensei1205@gmail.com",
          verification: { status: "verified" },
        },
      ],
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string, init?: RequestInit) => {
        expect(url).toBe("http://go-api:8080/api/v1/admin/payments");
        expect(new Headers(init?.headers).get("Authorization")).toBe(
          "Bearer server-token",
        );
        return Response.json(
          { orders: [order] },
          {
            headers: { "X-Request-ID": "req_admin_payments" },
          },
        );
      }),
    );

    const response = await GET(
      new Request("http://localhost/api/v1/admin/payments", {
        headers: { "X-Request-ID": "req_admin_payments" },
      }),
    );
    expect(response.status).toBe(200);
    await expect(response.json()).resolves.toEqual({ orders: [order] });
  });

  it("rejects another authenticated account before token or upstream access", async () => {
    const getToken = vi.fn();
    const fetch = vi.fn();
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      userId: "user_other",
      getToken,
    });
    mocks.currentUser.mockResolvedValue({
      id: "user_other",
      emailAddresses: [
        {
          emailAddress: "other@example.com",
          verification: { status: "verified" },
        },
      ],
    });
    vi.stubGlobal("fetch", fetch);

    const response = await GET(
      new Request("http://localhost/api/v1/admin/payments"),
    );
    expect(response.status).toBe(403);
    expect(getToken).not.toHaveBeenCalled();
    expect(fetch).not.toHaveBeenCalled();
  });
});
