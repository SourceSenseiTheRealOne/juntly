import { describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => {
  const apiResponse = { source: "api" };
  const intlResponse = { source: "intl" };

  return {
    apiResponse,
    clerkMiddleware: vi.fn(
      (handler: (...args: unknown[]) => unknown) => handler,
    ),
    createIntlMiddleware: vi.fn(() => vi.fn(() => intlResponse)),
    intlResponse,
    nextResponse: vi.fn(() => apiResponse),
  };
});

vi.mock("@clerk/nextjs/server", () => ({
  clerkMiddleware: mocks.clerkMiddleware,
}));
vi.mock("next-intl/middleware", () => ({
  default: mocks.createIntlMiddleware,
}));
vi.mock("next/server", () => ({
  NextResponse: { next: mocks.nextResponse },
}));

type ProxyHandler = (
  auth: unknown,
  request: { nextUrl: { pathname: string } },
) => Promise<unknown> | unknown;

async function loadProxy() {
  vi.resetModules();
  return import("./proxy");
}

describe("proxy", () => {
  it("composes Clerk with locale routing and matches API routes", async () => {
    const { config } = await loadProxy();

    expect(mocks.clerkMiddleware).toHaveBeenCalledOnce();
    expect(config.matcher).toContain("/(api|trpc)(.*)");
  });

  it("does not send API requests through locale routing", async () => {
    const { default: proxy } = await loadProxy();
    const clerkProxy = proxy as unknown as ProxyHandler;

    await expect(
      clerkProxy({}, { nextUrl: { pathname: "/api/v1/health" } }),
    ).resolves.toBe(mocks.apiResponse);
    expect(mocks.nextResponse).toHaveBeenCalledOnce();

    await expect(
      clerkProxy({}, { nextUrl: { pathname: "/pt-PT" } }),
    ).resolves.toBe(mocks.intlResponse);
  });
});
