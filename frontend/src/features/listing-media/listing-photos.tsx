"use client";

import Image from "next/image";
import { useEffect, useId, useRef, useState } from "react";
import type { ChangeEvent } from "react";
import type { ListingPhoto } from "@/shared/api/generated";
import { validPhotos } from "./media-contract";
import type { MediaScope } from "./media-contract";
import type { ListingPhotosCopy } from "./media-copy";
import { uploadPhoto } from "./media-upload";

type Props = {
  listingId: string;
  title: string;
  scope: MediaScope;
  copy: ListingPhotosCopy;
  editable?: boolean;
  busy?: boolean;
  onChanged?: () => void | Promise<void>;
  onBusyChange?: (busy: boolean) => void;
};

export function ListingPhotos(props: Props) {
  // A changed authority/listing gets a fresh state boundary and aborts old work.
  return <PhotoPanel key={`${props.scope}:${props.listingId}`} {...props} />;
}

function PhotoPanel({
  listingId,
  title,
  scope,
  copy,
  editable = false,
  busy = false,
  onChanged,
  onBusyChange,
}: Props) {
  const id = useId();
  const [expanded, setExpanded] = useState(scope === "public");
  const [photos, setPhotos] = useState<ListingPhoto[] | null>(null);
  const [failed, setFailed] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadFailed, setUploadFailed] = useState(false);
  const [saved, setSaved] = useState(false);
  const [retry, setRetry] = useState(0);
  const uploadController = useRef<AbortController | null>(null);
  const base = `/api/v1/${scope}/listings/${listingId}/media`;

  useEffect(() => {
    if (!expanded) return;
    const controller = new AbortController();
    void readPhotos(base, controller.signal)
      .then((items) => {
        if (!controller.signal.aborted) {
          setPhotos(items);
          setFailed(false);
        }
      })
      .catch(() => {
        if (!controller.signal.aborted) setFailed(true);
      });
    return () => controller.abort();
  }, [base, expanded, retry]);

  useEffect(
    () => () => {
      uploadController.current?.abort();
    },
    [],
  );

  async function selectFile(event: ChangeEvent<HTMLInputElement>) {
    const file = event.currentTarget.files?.[0];
    event.currentTarget.value = "";
    if (
      !file ||
      uploading ||
      busy ||
      !editable ||
      scope !== "me" ||
      photos === null
    )
      return;
    const ordinal = Array.from({ length: 10 }, (_, index) => index + 1).find(
      (slot) => !photos.some((photo) => photo.ordinal === slot),
    );
    if (!ordinal) return;
    const controller = new AbortController();
    uploadController.current = controller;
    setUploading(true);
    setUploadFailed(false);
    setSaved(false);
    onBusyChange?.(true);
    try {
      await uploadPhoto(listingId, file, ordinal, controller.signal);
      const updated = await readPhotos(base, controller.signal);
      if (!controller.signal.aborted) {
        setPhotos(updated);
        setSaved(true);
        await onChanged?.();
      }
    } catch {
      if (!controller.signal.aborted) setUploadFailed(true);
    } finally {
      onBusyChange?.(false);
      if (!controller.signal.aborted) setUploading(false);
    }
  }

  return (
    <section className="mt-5 border-t border-line pt-4" aria-label={copy.title}>
      {scope === "public" ? (
        <h2 className="font-semibold">{copy.title}</h2>
      ) : (
        <button
          type="button"
          className="market-button-secondary min-h-11"
          aria-expanded={expanded}
          aria-controls={`${id}-panel`}
          onClick={() => setExpanded((value) => !value)}
        >
          {copy.title}
        </button>
      )}
      <div id={`${id}-panel`} hidden={!expanded} className="mt-3 space-y-3">
        {failed ? (
          <div className="market-alert space-y-2" role="alert">
            <p>{copy.error}</p>
            <button
              type="button"
              className="market-button-secondary min-h-11"
              onClick={() => setRetry((value) => value + 1)}
            >
              {copy.retry}
            </button>
          </div>
        ) : photos === null ? (
          <p className="text-sm text-muted" role="status">
            {copy.loading}
          </p>
        ) : photos.length === 0 ? (
          <p className="text-sm text-muted">{copy.empty}</p>
        ) : (
          <div className="grid grid-cols-2 gap-3">
            {photos.map((photo) => (
              <Image
                key={photo.id}
                src={`${base}/${photo.id}`}
                width={photo.width}
                height={photo.height}
                unoptimized
                alt={copy.photoAlt
                  .replace("{title}", title)
                  .replace("{index}", String(photo.ordinal))}
                className="aspect-[4/3] h-auto w-full rounded-xl border border-line object-cover"
              />
            ))}
          </div>
        )}
        {editable && scope === "me" ? (
          <div className="space-y-2">
            <label
              htmlFor={`${id}-file`}
              className="block text-sm font-semibold"
            >
              {copy.add}
            </label>
            <input
              id={`${id}-file`}
              type="file"
              accept="image/jpeg,image/png,image/webp"
              className="market-control min-h-11 w-full min-w-0 text-sm"
              aria-describedby={`${id}-hint`}
              disabled={
                uploading ||
                busy ||
                failed ||
                photos === null ||
                photos.length >= 10
              }
              onChange={(event) => void selectFile(event)}
            />
            <p id={`${id}-hint`} className="text-xs leading-5 text-muted">
              {copy.hint} {copy.publication}
            </p>
            {photos?.length === 10 ? (
              <p className="text-sm text-muted">{copy.limit}</p>
            ) : null}
          </div>
        ) : null}
        <div aria-live="polite" className="text-sm text-muted">
          {uploading ? copy.uploading : saved ? copy.saved : null}
        </div>
        {uploadFailed ? (
          <p className="market-alert" role="alert">
            {copy.uploadError}
          </p>
        ) : null}
      </div>
    </section>
  );
}

async function readPhotos(
  base: string,
  signal: AbortSignal,
): Promise<ListingPhoto[]> {
  const response = await fetch(base, {
    cache: "no-store",
    credentials: "same-origin",
    signal,
  });
  if (!response.ok) throw new Error("Photos unavailable");
  const value: unknown = await response.json();
  if (!validPhotos(value)) throw new Error("Invalid photo response");
  return [...value.photos].sort((a, b) => a.ordinal - b.ordinal);
}
