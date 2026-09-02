import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { AccountCapabilitiesCard } from "./account-capabilities-card";

const copy = {
  title: "Como utiliza a Vila",
  description: "Escolha se também pretende disponibilizar serviços.",
  customerLabel: "Encontrar serviços",
  customerDescription: "A sua conta pode sempre procurar prestadores.",
  providerLabel: "Disponibilizar serviços",
  providerDescription:
    "Ative esta opção para preparar o seu perfil de prestador.",
  enabled: "Ativo",
  disabled: "Inativo",
  loading: "A carregar as capacidades da conta…",
  saving: "A guardar…",
  loadError: "Não foi possível carregar as capacidades da conta.",
  retry: "Tentar novamente",
  manageProvider: "Gerir perfil de prestador",
  manageListings: "Gerir anúncios",
};

afterEach(() => {
  vi.restoreAllMocks();
});

describe("AccountCapabilitiesCard", () => {
  it("shows loading then the implicit customer and current provider capabilities", async () => {
    const response = deferred<Response>();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => response.promise),
    );

    render(<AccountCapabilitiesCard copy={copy} />);

    expect(screen.getByText(copy.loading)).toBeInTheDocument();
    response.resolve(accountResponse(false));

    expect(await screen.findByText(copy.customerLabel)).toBeInTheDocument();
    expect(screen.getByText(copy.enabled)).toBeInTheDocument();
    expect(
      screen.getByRole("switch", { name: copy.providerLabel }),
    ).toHaveAttribute("aria-checked", "false");
    expect(document.body.textContent).not.toMatch(
      /[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}/i,
    );
  });

  it("updates provider capability through the same-origin BFF and locks controls while saving", async () => {
    const update = deferred<Response>();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(accountResponse(false))
      .mockImplementationOnce(() => update.promise);
    vi.stubGlobal("fetch", fetchMock);

    render(
      <AccountCapabilitiesCard
        copy={copy}
        providerProfileUrl="/pt-PT/account/provider-profile"
      />,
    );
    const toggle = await screen.findByRole("switch", {
      name: copy.providerLabel,
    });

    fireEvent.click(toggle);

    expect(toggle).toBeDisabled();
    expect(screen.getByText(copy.saving)).toBeInTheDocument();
    expect(fetchMock).toHaveBeenLastCalledWith(
      "/api/v1/me/account",
      expect.objectContaining({
        method: "PUT",
        body: JSON.stringify({ providerEnabled: true }),
      }),
    );

    update.resolve(accountResponse(true));
    await waitFor(() => expect(toggle).toHaveAttribute("aria-checked", "true"));
    expect(toggle).not.toBeDisabled();
    expect(
      screen.getByRole("link", { name: copy.manageProvider }),
    ).toHaveAttribute("href", "/pt-PT/account/provider-profile");
  });

  it("shows a controlled error and retries without exposing upstream details", async () => {
    const fetchMock = vi
      .fn()
      .mockRejectedValueOnce(new Error("ECONNREFUSED http://internal-api"))
      .mockResolvedValueOnce(accountResponse(false));
    vi.stubGlobal("fetch", fetchMock);

    render(<AccountCapabilitiesCard copy={copy} />);

    expect(await screen.findByText(copy.loadError)).toBeInTheDocument();
    expect(document.body.textContent).not.toContain("internal-api");
    expect(document.body.textContent).not.toContain("ECONNREFUSED");

    fireEvent.click(screen.getByRole("button", { name: copy.retry }));

    expect(await screen.findByText(copy.customerLabel)).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("rejects malformed or expanded account responses", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        Response.json({
          customerEnabled: true,
          providerEnabled: false,
          onboardingCompletedAt: "2026-08-23T12:05:00Z",
          internalUserId: "not-public",
        }),
      ),
    );

    render(<AccountCapabilitiesCard copy={copy} />);

    expect(await screen.findByText(copy.loadError)).toBeInTheDocument();
    expect(document.body.textContent).not.toContain("not-public");
  });
});

function accountResponse(providerEnabled: boolean): Response {
  return Response.json({
    customerEnabled: true,
    providerEnabled,
    onboardingCompletedAt: "2026-08-23T12:05:00.123456Z",
  });
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, reject, resolve };
}
