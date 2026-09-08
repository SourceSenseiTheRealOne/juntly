import { webcrypto } from "node:crypto";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { ListingPhotos } from "./listing-photos";
import { getMediaCopy } from "./media-copy";
const listingId = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  mediaId = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb";
afterEach(() => vi.unstubAllGlobals());
it("loads public verified photos with localized accessible text and without image optimization", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      Response.json({
        photos: [{ id: mediaId, ordinal: 1, width: 2, height: 3 }],
      }),
    ),
  );
  render(
    <ListingPhotos
      listingId={listingId}
      title="Jardinagem"
      scope="public"
      copy={getMediaCopy("pt-PT")}
    />,
  );
  const image = await screen.findByRole("img", {
    name: "Jardinagem, fotografia 1",
  });
  expect(image).toHaveAttribute(
    "src",
    `/api/v1/public/listings/${listingId}/media/${mediaId}`,
  );
  expect(
    screen.queryByLabelText("Adicionar fotografia"),
  ).not.toBeInTheDocument();
});
it("does not display a selected image until authoritative finalization completes", async () => {
  vi.stubGlobal("crypto", webcrypto);
  let published = false;
  let finish: (response: Response) => void = () => {};
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: RequestInfo | URL) => {
      if (String(url).endsWith("upload-intents"))
        return Response.json({
          mediaId,
          capability: {
            url: `https://storage.example.test/storage/v1/object/upload/sign/listing-images/pending/${mediaId}?token=synthetic`,
            method: "PUT",
            headers: { "Content-Type": "image/png" },
          },
        });
      if (String(url).endsWith("finalize"))
        return new Promise<Response>((resolve) => {
          finish = (response) => {
            published = true;
            resolve(response);
          };
        });
      if (String(url).startsWith("https://storage"))
        return new Response(null, { status: 200 });
      return Response.json({
        photos: published
          ? [{ id: mediaId, ordinal: 1, width: 2, height: 3 }]
          : [],
      });
    }),
  );
  const changed = vi.fn();
  render(
    <ListingPhotos
      listingId={listingId}
      title="Jardinagem"
      scope="me"
      editable
      copy={getMediaCopy("pt-PT")}
      onChanged={changed}
    />,
  );
  expect(fetch).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Fotografias" }));
  await screen.findByText("Ainda não há fotografias.");
  const file = new File([new Uint8Array([1, 2, 3])], "photo.png", {
    type: "image/png",
  });
  Object.defineProperty(file, "arrayBuffer", {
    value: async () => new Uint8Array([1, 2, 3]).buffer,
  });
  fireEvent.change(screen.getByLabelText("Adicionar fotografia"), {
    target: { files: [file] },
  });
  await waitFor(() =>
    expect(
      vi
        .mocked(fetch)
        .mock.calls.some(([url]) => String(url).endsWith("finalize")),
    ).toBe(true),
  );
  expect(screen.queryByRole("img")).not.toBeInTheDocument();
  expect(screen.getByLabelText("Adicionar fotografia")).toBeDisabled();
  finish(new Response(null, { status: 204 }));
  await screen.findByRole("img");
  expect(changed).toHaveBeenCalledOnce();
});
it.each(["pt-PT", "en", "es"] as const)(
  "renders a localized unavailable state in %s",
  async (locale) => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(null, { status: 503 })),
    );
    const copy = getMediaCopy(locale);
    render(
      <ListingPhotos
        listingId={listingId}
        title="Photo"
        scope="public"
        copy={copy}
      />,
    );
    expect(await screen.findByRole("alert")).toHaveTextContent(copy.error);
  },
);
