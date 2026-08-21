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

import { LandingShell } from "./landing-shell";

const copy = {
  eyebrow: "Local people. Real skills.",
  tagline: "Encontra quem sabe fazer.",
  heading: "Serviços locais, mais perto de si.",
  description: "Uma forma simples de encontrar pessoas com competências reais.",
  statusLabel: "A plataforma está a nascer.",
  signInLabel: "Entrar",
  signInUrl: "/pt-PT/sign-in",
  signUpLabel: "Criar conta",
  signUpUrl: "/pt-PT/sign-up",
  visionLinkLabel: "Conhecer a visão",
  visionTitle: "Criada para ligações locais reais",
  visionDescription:
    "Descoberta, contacto e confiança sem retirar a escolha às pessoas.",
  footerLabel: "Juntly — com origem em Portugal.",
};

describe("LandingShell", () => {
  it("renders semantic landmarks with supplied localized copy", () => {
    render(<LandingShell {...copy} />);

    expect(screen.getByRole("banner")).toBeInTheDocument();
    expect(screen.getByRole("main")).toBeInTheDocument();
    expect(screen.getByRole("contentinfo")).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(
      copy.heading,
    );
    expect(screen.getByText(copy.tagline)).toBeInTheDocument();
    expect(screen.getByText(copy.statusLabel)).toBeInTheDocument();
  });

  it("provides a focusable link to the vision section", () => {
    render(<LandingShell {...copy} />);

    expect(
      screen.getByRole("link", { name: copy.visionLinkLabel }),
    ).toHaveAttribute("href", "#vision");
    expect(
      screen.getByRole("heading", { name: copy.visionTitle }),
    ).toBeInTheDocument();
  });

  it("offers localized account entry points without advertising marketplace actions", () => {
    render(<LandingShell {...copy} />);

    expect(
      screen.getByRole("link", { name: copy.signInLabel }),
    ).toHaveAttribute("href", copy.signInUrl);
    expect(
      screen.getByRole("link", { name: copy.signUpLabel }),
    ).toHaveAttribute("href", copy.signUpUrl);
    expect(
      screen.queryByRole("button", {
        name: /publish|buy|checkout/i,
      }),
    ).not.toBeInTheDocument();
  });
});
