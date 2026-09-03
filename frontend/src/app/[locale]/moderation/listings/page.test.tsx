import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({
  getTranslations: vi.fn(),
  requireSoleAdministrator: vi.fn(),
}));

vi.mock("next-intl/server", () => ({
  getTranslations: mocks.getTranslations,
}));
vi.mock("@/features/auth/sole-administrator", () => ({
  requireSoleAdministrator: mocks.requireSoleAdministrator,
}));
vi.mock("@/features/auth/require-session", () => ({
  requireAuthenticatedUser: vi.fn(),
}));
vi.mock("@/features/listings/moderation-queue", () => ({
  ModerationQueue: ({ copy }: { copy: { title: string } }) => (
    <div data-testid="moderation-queue">{copy.title}</div>
  ),
}));

import ModerationListingsPage, { dynamic } from "./page";

afterEach(() => {
  mocks.getTranslations.mockReset();
  mocks.requireSoleAdministrator.mockReset();
});

describe("ModerationListingsPage", () => {
  it("is dynamic and requires the sole administrator before rendering", async () => {
    mocks.requireSoleAdministrator.mockResolvedValue("user_admin");
    mocks.getTranslations.mockResolvedValue((key: string) => key);

    render(
      await ModerationListingsPage({
        params: Promise.resolve({ locale: "pt-PT" }),
      }),
    );

    expect(dynamic).toBe("force-dynamic");
    expect(mocks.requireSoleAdministrator).toHaveBeenCalledWith("pt-PT");
    expect(screen.getByTestId("moderation-queue")).toHaveTextContent("title");
  });
});
