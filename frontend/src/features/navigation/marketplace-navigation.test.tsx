import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/features/auth/auth-navigation", () => ({
  AuthNavigation: ({
    signInLabel,
    signInUrl,
    signUpLabel,
    signUpUrl,
  }: {
    signInLabel: string;
    signInUrl: string;
    signUpLabel: string;
    signUpUrl: string;
  }) => (
    <nav aria-label="Account">
      <a href={signInUrl}>{signInLabel}</a>
      <a href={signUpUrl}>{signUpLabel}</a>
    </nav>
  ),
}));

import { MarketplaceNavigation } from "./marketplace-navigation";

describe("MarketplaceNavigation", () => {
  it("keeps localized browse and account routes available in the shared app header", () => {
    render(
      <MarketplaceNavigation
        accountLabel="A minha conta"
        accountUrl="/pt-PT/account"
        accountNavigationLabel="Opções da conta"
        discoverLabel="Encontrar serviços"
        discoverUrl="/pt-PT/discover"
        navigationLabel="Navegação do mercado"
        signInLabel="Entrar"
        signInUrl="/pt-PT/sign-in"
        signUpLabel="Criar conta"
        signUpUrl="/pt-PT/sign-up"
      />,
    );

    expect(screen.getByRole("banner")).toBeInTheDocument();
    expect(
      screen.getByRole("navigation", { name: "Navegação do mercado" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Juntly" })).toHaveAttribute(
      "href",
      "/pt-PT",
    );
    expect(
      screen.getByRole("link", { name: "Encontrar serviços" }),
    ).toHaveAttribute("href", "/pt-PT/discover");
    expect(screen.getByRole("link", { name: "A minha conta" })).toHaveAttribute(
      "href",
      "/pt-PT/account",
    );
    expect(screen.getByRole("link", { name: "Entrar" })).toHaveAttribute(
      "href",
      "/pt-PT/sign-in",
    );
  });
});
