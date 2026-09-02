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
vi.mock("@/features/account/account-capabilities-card", () => ({
  AccountCapabilitiesCard: ({ copy }: { copy: { providerLabel: string } }) => (
    <div data-testid="account-capabilities-card">{copy.providerLabel}</div>
  ),
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
          "capabilities.customerDescription":
            "A sua conta pode sempre procurar prestadores.",
          "capabilities.customerLabel": "Encontrar serviços",
          "capabilities.description":
            "Escolha se também pretende disponibilizar serviços.",
          "capabilities.disabled": "Inativo",
          "capabilities.enabled": "Ativo",
          "capabilities.loadError":
            "Não foi possível carregar as capacidades da conta.",
          "capabilities.loading": "A carregar as capacidades da conta…",
          "capabilities.manageProvider": "Gerir perfil de prestador",
          "capabilities.providerDescription":
            "Ative esta opção para preparar o seu perfil de prestador.",
          "capabilities.providerLabel": "Disponibilizar serviços",
          "capabilities.retry": "Tentar novamente",
          "capabilities.saving": "A guardar…",
          "capabilities.title": "Como utiliza a Vila",
          description: "A sua sessão está ativa.",
          title: "Conta Vila",
        })[key],
    );

    render(
      await AccountPage({
        params: Promise.resolve({ locale: "pt-PT" }),
      } as never),
    );

    expect(mocks.requireAuthenticatedUser).toHaveBeenCalledWith("pt-PT");
    expect(
      screen.getByRole("heading", { name: "Conta Vila" }),
    ).toBeInTheDocument();
    expect(screen.getByText("A sua sessão está ativa.")).toBeInTheDocument();
    expect(screen.getByTestId("account-capabilities-card")).toHaveTextContent(
      "Disponibilizar serviços",
    );
  });
});
