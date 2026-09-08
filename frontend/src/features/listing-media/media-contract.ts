import type {
  CreateUploadIntentRequest,
  ListingPhotosResponse,
  UploadIntentResponse,
} from "@/shared/api/generated";

export const maxPhotoBytes = 10 * 1024 * 1024;
export type MediaScope = "public" | "me" | "moderation";

export function exact(
  value: unknown,
  keys: string[],
): value is Record<string, unknown> {
  return (
    value !== null &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    Object.keys(value).length === keys.length &&
    keys.every((key) => Object.hasOwn(value, key))
  );
}
export function mediaID(value: unknown): value is string {
  return (
    typeof value === "string" &&
    value !== "00000000-0000-0000-0000-000000000000" &&
    /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/.test(value)
  );
}
function integer(value: unknown, min: number, max: number): value is number {
  return (
    typeof value === "number" &&
    Number.isInteger(value) &&
    value >= min &&
    value <= max
  );
}
export function validPhotos(value: unknown): value is ListingPhotosResponse {
  if (
    !exact(value, ["photos"]) ||
    !Array.isArray(value.photos) ||
    value.photos.length > 10
  )
    return false;
  const ids = new Set<string>(),
    ordinals = new Set<number>();
  return value.photos.every((photo) => {
    if (
      !exact(photo, ["id", "ordinal", "width", "height"]) ||
      !mediaID(photo.id) ||
      !integer(photo.ordinal, 1, 10) ||
      !integer(photo.width, 1, 8192) ||
      !integer(photo.height, 1, 8192) ||
      photo.width * photo.height > 16_000_000 ||
      ids.has(photo.id) ||
      ordinals.has(photo.ordinal)
    )
      return false;
    ids.add(photo.id);
    ordinals.add(photo.ordinal);
    return true;
  });
}
export function validUploadRequest(
  value: unknown,
): value is CreateUploadIntentRequest {
  return (
    exact(value, ["ordinal", "contentType", "byteSize", "checksumSha256"]) &&
    integer(value.ordinal, 1, 10) &&
    ["image/png", "image/jpeg", "image/webp"].includes(
      String(value.contentType),
    ) &&
    integer(value.byteSize, 1, maxPhotoBytes) &&
    typeof value.checksumSha256 === "string" &&
    /^[0-9a-f]{64}$/.test(value.checksumSha256)
  );
}
export function validUploadIntent(
  value: unknown,
  expectedOrigin?: string,
): value is UploadIntentResponse {
  if (
    !exact(value, ["mediaId", "capability"]) ||
    !mediaID(value.mediaId) ||
    !exact(value.capability, ["url", "method", "headers"]) ||
    value.capability.method !== "PUT" ||
    typeof value.capability.url !== "string" ||
    value.capability.url.length > 16384 ||
    !exact(value.capability.headers, ["Content-Type"]) ||
    !["image/png", "image/jpeg", "image/webp"].includes(
      String(value.capability.headers["Content-Type"]),
    )
  )
    return false;
  try {
    const url = new URL(value.capability.url);
    if (
      url.username ||
      url.password ||
      url.hash ||
      (expectedOrigin !== undefined && url.origin !== expectedOrigin) ||
      (url.protocol !== "https:" &&
        !(
          url.protocol === "http:" &&
          ["localhost", "127.0.0.1"].includes(url.hostname)
        ))
    )
      return false;
    const match =
      /^\/storage\/v1\/object\/upload\/sign\/([a-z][a-z0-9-]{2,62})\/pending\/([0-9a-f-]+)$/.exec(
        url.pathname,
      );
    const entries = [...url.searchParams.entries()];
    return (
      !!match &&
      match[2] === value.mediaId &&
      entries.length === 1 &&
      entries[0][0] === "token" &&
      entries[0][1].length > 0 &&
      entries[0][1].length <= 8192 &&
      !/[\r\n\x00]/.test(entries[0][1])
    );
  } catch {
    return false;
  }
}
