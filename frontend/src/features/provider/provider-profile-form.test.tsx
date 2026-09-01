import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { ProviderProfileForm } from "./provider-profile-form";

const copy = {
  title: "Perfil de prestador",
  description: "Prepare os dados que serão usados nos próximos passos.",
  displayName: "Nome de apresentação",
  providerType: "Tipo de prestador",
  individual: "Particular",
  professional: "Profissional",
  business: "Empresa",
  bio: "Apresentação",
  primaryLocality: "Localidade principal",
  serviceLocalities: "Áreas de serviço",
  languages: "Idiomas",
  travelRadius: "Distância máxima em quilómetros",
  serviceModes: "Modos de serviço",
  travels: "Desloca-se ao cliente",
  receives: "Recebe clientes",
  remote: "Trabalha à distância",
  save: "Guardar perfil",
  saving: "A guardar…",
  loading: "A carregar perfil…",
  error: "Não foi possível carregar ou guardar o perfil.",
  retry: "Tentar novamente",
  saved: "Perfil guardado.",
};

const locality = {
  id: "11111111-1111-4111-8111-111111111111",
  slug: "zebreira",
  name: "Zebreira",
  parishName: "Zebreira e Segura",
  municipalityName: "Idanha-a-Nova",
  districtName: "Castelo Branco",
};

const language = { code: "pt-PT", name: "Português" };

afterEach(() => vi.restoreAllMocks());

describe("ProviderProfileForm", () => {
  it("loads an owner-only empty profile with public references and attribution", async () => {
    vi.stubGlobal("fetch", referenceFetch());
    render(<ProviderProfileForm locale="pt-PT" copy={copy} />);

    expect(screen.getByText(copy.loading)).toBeInTheDocument();
    expect(
      await screen.findByRole("heading", { name: copy.title }),
    ).toBeInTheDocument();
    expect(screen.getByLabelText(copy.displayName)).toBeInTheDocument();
    expect(screen.getByLabelText(copy.primaryLocality)).toHaveValue(
      locality.id,
    );
    expect(
      screen.getByRole("group", { name: copy.serviceModes }),
    ).toBeInTheDocument();
    expect(screen.getByLabelText(language.name)).toBeChecked();
    expect(
      screen.getByText("© OpenStreetMap contributors"),
    ).toBeInTheDocument();
    expect(document.body.textContent).not.toContain("internalUserId");
    expect(document.body.textContent).not.toContain("latitude");
    expect(document.body.textContent).not.toContain("phone");
  });

  it("submits the exact provider replacement and locks the form while saving", async () => {
    const pending = deferred<Response>();
    const fetchMock = referenceFetch();
    fetchMock.mockImplementationOnce(async () => localitiesResponse());
    fetchMock.mockImplementationOnce(async () => languagesResponse());
    fetchMock.mockImplementationOnce(async () =>
      Response.json({ profile: null }),
    );
    fetchMock.mockImplementationOnce(() => pending.promise);
    vi.stubGlobal("fetch", fetchMock);
    render(<ProviderProfileForm locale="pt-PT" copy={copy} />);

    const name = await screen.findByLabelText(copy.displayName);
    fireEvent.change(name, { target: { value: "Prestador local" } });
    fireEvent.click(screen.getByRole("button", { name: copy.save }));

    expect(screen.getByRole("button", { name: copy.saving })).toBeDisabled();
    expect(fetchMock).toHaveBeenLastCalledWith(
      "/api/v1/me/provider-profile",
      expect.objectContaining({
        method: "PUT",
        body: expect.stringContaining('"displayName":"Prestador local"'),
      }),
    );

    pending.resolve(
      Response.json({
        profile: {
          displayName: "Prestador local",
          providerType: "individual",
          bio: "",
          primaryLocalityId: locality.id,
          serviceLocalityIds: [locality.id],
          maxTravelDistanceKm: 25,
          travelsToCustomer: true,
          receivesCustomer: false,
          remoteServices: false,
          languageCodes: ["pt-PT"],
          createdAt: "2026-08-23T16:00:00Z",
          updatedAt: "2026-08-23T16:00:00Z",
        },
      }),
    );
    expect(await screen.findByText(copy.saved)).toBeInTheDocument();
  });

  it("shows a controlled retry state without upstream diagnostics", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new Error("ECONNREFUSED internal-api")),
    );
    render(<ProviderProfileForm locale="pt-PT" copy={copy} />);
    expect(await screen.findByText(copy.error)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: copy.retry }),
    ).toBeInTheDocument();
    expect(document.body.textContent).not.toContain("internal-api");
  });
});

function referenceFetch() {
  return vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.includes("reference/localities")) return localitiesResponse();
    if (url.includes("reference/languages")) return languagesResponse();
    if (url.includes("provider-profile"))
      return Response.json({ profile: null });
    throw new Error("unexpected request");
  });
}

function localitiesResponse() {
  return Response.json({
    localities: [locality],
    attribution: {
      text: "© OpenStreetMap contributors",
      url: "https://www.openstreetmap.org/copyright",
    },
  });
}
function languagesResponse() {
  return Response.json({ languages: [language] });
}
function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((value) => {
    resolve = value;
  });
  return { promise, resolve };
}
