import { afterEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({
  auth: vi.fn(),
  redirect: vi.fn(),
}));

vi.mock("@clerk/nextjs/server", () => ({ auth: mocks.auth }));
vi.mock("next/navigation", () => ({ redirect: mocks.redirect }));

import { requireAuthenticatedUser } from "./require-session";

afterEach(() => {
  mocks.auth.mockReset();
  mocks.redirect.mockReset();
});

describe("requireAuthenticatedUser", () => {
  it("fails closed by redirecting a signed-out visitor to the localized sign-in route", async () => {
    mocks.auth.mockResolvedValue({ isAuthenticated: false, userId: null });
    mocks.redirect.mockImplementation(() => {
      throw new Error("redirected");
    });

    await expect(requireAuthenticatedUser("pt-PT")).rejects.toThrow(
      "redirected",
    );
    expect(mocks.redirect).toHaveBeenCalledWith("/pt-PT/sign-in");
  });

  it("returns only the verified opaque Clerk subject for an authenticated session", async () => {
    mocks.auth.mockResolvedValue({
      isAuthenticated: true,
      userId: "user_verified_subject",
    });

    await expect(requireAuthenticatedUser("en")).resolves.toBe(
      "user_verified_subject",
    );
    expect(mocks.redirect).not.toHaveBeenCalled();
  });
});
