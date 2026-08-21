import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({
  getTranslations: vi.fn(),
  requireAuthenticatedUser: vi.fn(),
}));

vi.mock("next-intl/server", () => ({
  getTranslations: mocks.getTranslations,
}));
vi.mock("@/features/auth/require-session", () => ({
  requireAuthenticatedUser: mocks.requireAuthenticatedUser,
}));

import AccountPage, { dynamic } from "./page";

afterEach(() => {
  mocks.getTranslations.mockReset();
  mocks.requireAuthenticatedUser.mockReset();
});

describe("AccountPage", () => {
  it("is explicitly dynamically rendered because it resolves request-scoped identity", () => {
    expect(dynamic).toBe("force-dynamic");
  });

  it("requires a verified session before rendering the localized account confirmation", async () => {
    mocks.requireAuthenticatedUser.mockResolvedValue("user_verified_subject");
    mocks.getTranslations.mockResolvedValue(
      (key: string) =>
        ({
          description: "A sua sessão está ativa.",
          title: "Conta Juntly",
        })[key],
    );

    render(
      await AccountPage({
        params: Promise.resolve({ locale: "pt-PT" }),
      } as never),
    );

    expect(mocks.requireAuthenticatedUser).toHaveBeenCalledWith("pt-PT");
    expect(
      screen.getByRole("heading", { name: "Conta Juntly" }),
    ).toBeInTheDocument();
    expect(screen.getByText("A sua sessão está ativa.")).toBeInTheDocument();
  });
});
