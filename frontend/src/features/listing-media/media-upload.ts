import { maxPhotoBytes, mediaID, validUploadIntent } from "./media-contract";

export async function uploadPhoto(
  listingId: string,
  file: File,
  ordinal: number,
  signal?: AbortSignal,
): Promise<string> {
  if (
    !mediaID(listingId) ||
    !Number.isInteger(ordinal) ||
    ordinal < 1 ||
    ordinal > 10 ||
    !["image/jpeg", "image/png", "image/webp"].includes(file.type) ||
    file.size < 1 ||
    file.size > maxPhotoBytes
  )
    throw new Error("Invalid image");
  const bytes = await file.arrayBuffer();
  if (bytes.byteLength !== file.size) throw new Error("Invalid image");
  const checksum = Array.from(
    new Uint8Array(await crypto.subtle.digest("SHA-256", bytes)),
    (byte) => byte.toString(16).padStart(2, "0"),
  ).join("");
  const timeout = () =>
    signal
      ? AbortSignal.any([signal, AbortSignal.timeout(60000)])
      : AbortSignal.timeout(60000);
  const root = `/api/v1/me/listings/${listingId}/media`;
  const intentResponse = await fetch(`${root}/upload-intents`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "same-origin",
    cache: "no-store",
    redirect: "error",
    signal: timeout(),
    body: JSON.stringify({
      ordinal,
      contentType: file.type,
      byteSize: file.size,
      checksumSha256: checksum,
    }),
  });
  if (!intentResponse.ok) throw new Error("Upload unavailable");
  const intent: unknown = await intentResponse.json();
  if (
    !validUploadIntent(intent) ||
    intent.capability.headers["Content-Type"] !== file.type
  )
    throw new Error("Invalid upload response");
  try {
    await fetch(intent.capability.url, {
      method: "PUT",
      headers: intent.capability.headers,
      body: bytes,
      credentials: "omit",
      referrerPolicy: "no-referrer",
      redirect: "error",
      signal: timeout(),
    });
  } catch {
    // A lost PUT response does not prove failure. Only server finalization can
    // confirm whether the immutable object exists and matches the reservation.
    if (signal?.aborted) throw new Error("Upload cancelled");
  }
  const finalized = await fetch(`${root}/${intent.mediaId}/finalize`, {
    method: "POST",
    credentials: "same-origin",
    cache: "no-store",
    redirect: "error",
    signal: timeout(),
  });
  if (finalized.status !== 204) throw new Error("Image could not be verified");
  return intent.mediaId;
}
