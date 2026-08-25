import { render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
const mocks = vi.hoisted(() => ({
  getTranslations: vi.fn(),
  requireAuthenticatedUser: vi.fn(),
}));
vi.mock("next-intl/server", () => ({ getTranslations: mocks.getTranslations }));
vi.mock("@/features/auth/require-session", () => ({
  requireAuthenticatedUser: mocks.requireAuthenticatedUser,
}));
vi.mock("@/features/listings/listing-dashboard", () => ({
  ListingDashboard: ({ locale }: { locale: string }) => (
    <div data-testid="listing-dashboard">{locale}</div>
  ),
}));
import ListingsPage, { dynamic } from "./page";
afterEach(() => {
  mocks.getTranslations.mockReset();
  mocks.requireAuthenticatedUser.mockReset();
});
it("protects and renders dynamic localized listings page", async () => {
  mocks.requireAuthenticatedUser.mockResolvedValue("verified");
  mocks.getTranslations.mockResolvedValue((key: string) => key);
  render(
    await ListingsPage({
      params: Promise.resolve({ locale: "pt-PT" }),
    } as never),
  );
  expect(dynamic).toBe("force-dynamic");
  expect(mocks.requireAuthenticatedUser).toHaveBeenCalledWith("pt-PT");
  expect(screen.getByTestId("listing-dashboard")).toHaveTextContent("pt-PT");
});
