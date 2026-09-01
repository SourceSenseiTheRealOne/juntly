import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { LandingShell } from "./landing-shell";

const copy = {
  tagline: "Encontra quem sabe fazer.",
  heading: "Serviços locais, mais perto de si.",
  description: "Uma forma simples de encontrar pessoas com competências reais.",
  statusLabel: "A plataforma está a nascer.",
  healthStatusLabels: {
    checking: "A verificar a API.",
    available: "API disponível.",
    unavailable: "API temporariamente indisponível.",
  },

  visionLinkLabel: "Conhecer a visão",
  visionTitle: "Criada para ligações locais reais",
  visionDescription:
    "Descoberta, contacto e confiança sem retirar a escolha às pessoas.",
  showcaseTitle: "Fazer local, mais simples.",
  footerLabel: "Juntly — com origem em Portugal.",
};

describe("LandingShell", () => {
  it("renders semantic landmarks with supplied localized copy", () => {
    render(<LandingShell {...copy} />);

    expect(screen.getByRole("main")).toBeInTheDocument();
    expect(screen.getByRole("contentinfo")).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(
      copy.heading,
    );
    expect(screen.getByText(copy.tagline)).toBeInTheDocument();
    expect(screen.getByText(copy.statusLabel)).toBeInTheDocument();
    expect(screen.getByText(copy.showcaseTitle)).toBeInTheDocument();
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

  it("keeps the landing surface focused without advertising incomplete marketplace actions", () => {
    render(<LandingShell {...copy} />);

    expect(
      screen.queryByRole("button", {
        name: /publish|buy|checkout/i,
      }),
    ).not.toBeInTheDocument();
  });
});
