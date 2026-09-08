import { webcrypto } from "node:crypto";
import { afterEach, expect, it, vi } from "vitest";
import { uploadPhoto } from "./media-upload";
const listingId = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
const mediaId = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb";
const capability = {
  mediaId,
  capability: {
    url: `https://storage.example.test/storage/v1/object/upload/sign/listing-images/pending/${mediaId}?token=synthetic-capability`,
    method: "PUT",
    headers: { "Content-Type": "image/png" },
  },
};
afterEach(() => vi.unstubAllGlobals());
function file() {
  const value = new File([new Uint8Array([1, 2, 3])], "photo.png", {
    type: "image/png",
  });
  Object.defineProperty(value, "arrayBuffer", {
    value: async () => new Uint8Array([1, 2, 3]).buffer,
  });
  return value;
}
it("uploads with only the scoped capability and waits for authoritative finalization", async () => {
  vi.stubGlobal("crypto", webcrypto);
  const fetcher = vi
    .fn()
    .mockResolvedValueOnce(Response.json(capability))
    .mockResolvedValueOnce(new Response(null, { status: 200 }))
    .mockResolvedValueOnce(new Response(null, { status: 204 }));
  vi.stubGlobal("fetch", fetcher);
  expect(await uploadPhoto(listingId, file(), 1)).toBe(mediaId);
  const [url, options] = fetcher.mock.calls[1];
  expect(url).toBe(capability.capability.url);
  expect(options.credentials).toBe("omit");
  expect(options.redirect).toBe("error");
  expect(options.headers).toEqual({ "Content-Type": "image/png" });
  expect(fetcher.mock.calls[2][0]).toBe(
    `/api/v1/me/listings/${listingId}/media/${mediaId}/finalize`,
  );
  const input = JSON.parse(fetcher.mock.calls[0][1].body);
  expect(input.checksumSha256).toMatch(/^[0-9a-f]{64}$/);
  expect(input.byteSize).toBe(3);
});
it("recovers an ambiguous PUT only when finalization confirms stored bytes", async () => {
  vi.stubGlobal("crypto", webcrypto);
  const fetcher = vi
    .fn()
    .mockResolvedValueOnce(Response.json(capability))
    .mockRejectedValueOnce(new Error("network interrupted"))
    .mockResolvedValueOnce(new Response(null, { status: 204 }));
  vi.stubGlobal("fetch", fetcher);
  expect(await uploadPhoto(listingId, file(), 1)).toBe(mediaId);
});
it("rejects invalid input and never treats an unverified upload as success", async () => {
  vi.stubGlobal("crypto", webcrypto);
  const fetcher = vi.fn();
  vi.stubGlobal("fetch", fetcher);
  await expect(
    uploadPhoto(
      listingId,
      new File(["script"], "bad.svg", { type: "image/svg+xml" }),
      1,
    ),
  ).rejects.toThrow();
  expect(fetcher).not.toHaveBeenCalled();
  fetcher
    .mockResolvedValueOnce(Response.json(capability))
    .mockResolvedValueOnce(new Response(null, { status: 200 }))
    .mockResolvedValueOnce(new Response(null, { status: 400 }));
  await expect(uploadPhoto(listingId, file(), 1)).rejects.toThrow();
});
