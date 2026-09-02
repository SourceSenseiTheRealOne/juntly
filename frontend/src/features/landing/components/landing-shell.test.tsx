import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { LandingShell } from "./landing-shell";

const copy = {
  tagline: "Encontra quem sabe fazer.",
  heading: "Serviços locais, mais perto de si.",
  description: "Uma forma simples de encontrar pessoas com competências reais.",
  visionLinkLabel: "Conhecer a visão",
  discoverLinkLabel: "Explorar serviços",
  discoverUrl: "/pt-PT/discover",
  accountUrl: "/pt-PT/account",
  signUpUrl: "/pt-PT/sign-up",
  visionTitle: "Criada para ligações locais reais",
  visionDescription:
    "Descoberta, contacto e confiança sem retirar a escolha às pessoas.",
  showcaseTitle: "Fazer local, mais simples.",
  howTitle: "Da necessidade à solução, sem complicações.",
  howDescription: "Um percurso claro.",
  discoverBlock: { title: "Procure", description: "Encontre serviços." },
  compareBlock: { title: "Compare", description: "Consulte propostas." },
  contactBlock: { title: "Contacte", description: "Fale em segurança." },
  customerBlock: {
    title: "Para clientes",
    description: "Encontre ajuda.",
    action: "Encontrar um serviço",
  },
  providerBlock: {
    title: "Para prestadores",
    description: "Disponibilize competências.",
    action: "Preparar perfil",
    imageAlt: "Profissional a trabalhar numa oficina",
  },
  trustTitle: "Confiança construída no produto.",
  trustDescription: "Privacidade e contexto local.",
  privacyBlock: { title: "Privacidade", description: "Contactos protegidos." },
  localBlock: { title: "Local", description: "Resultados relevantes." },
  reputationBlock: {
    title: "Reputação",
    description: "Avaliações verificadas.",
  },
  closingTitle: "Comece pela sua comunidade.",
  closingDescription: "Crie uma conta.",
  closingAction: "Criar conta",
  footerLabel: "Vila, com origem em Portugal.",
};

describe("LandingShell", () => {
  it("renders a complete localized marketplace home", () => {
    render(<LandingShell {...copy} />);

    expect(screen.getByRole("main")).toBeInTheDocument();
    expect(screen.getByRole("contentinfo")).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(
      copy.heading,
    );
    expect(
      screen.getByRole("heading", { name: copy.howTitle }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: copy.trustTitle }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: copy.closingTitle }),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Vila vision")).toHaveTextContent("Vila.");
    expect(
      screen.getByRole("img", { name: copy.providerBlock.imageAlt }),
    ).toHaveAttribute("src", expect.stringContaining("local-provider.jpg"));
  });

  it("links to discovery, account onboarding, and the vision section", () => {
    render(<LandingShell {...copy} />);

    expect(
      screen.getByRole("link", { name: copy.discoverLinkLabel }),
    ).toHaveAttribute("href", copy.discoverUrl);
    expect(
      screen.getByRole("link", { name: copy.visionLinkLabel }),
    ).toHaveAttribute("href", "#vision");
    expect(
      screen.getByRole("link", { name: copy.providerBlock.action }),
    ).toHaveAttribute("href", copy.accountUrl);
    expect(
      screen.getByRole("link", { name: copy.closingAction }),
    ).toHaveAttribute("href", copy.signUpUrl);
  });

  it("does not run or display an API status check", () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    render(<LandingShell {...copy} />);

    expect(fetchMock).not.toHaveBeenCalled();
    expect(screen.queryByText(/API/i)).not.toBeInTheDocument();
  });
});
