-- Keep the immutable original upload separate from the server-encoded PNG.
alter table public.listing_media
  add column verified_object_reference text,
  add column verified_checksum_sha256 text,
  add column verified_byte_size bigint,
  add column pixel_width integer,
  add column pixel_height integer;

-- Legacy ready rows are not automatically trusted or rewritten. Application
-- reads require verified fields; NOT VALID still enforces this on new writes.
alter table public.listing_media add constraint listing_media_verified_ready check (
  state <> 'ready' or (
    verified_object_reference is not null and
    verified_checksum_sha256 is not null and
    verified_byte_size is not null and
    pixel_width is not null and pixel_height is not null and
    verified_checksum_sha256 ~ '^[0-9a-f]{64}$' and
    verified_object_reference = 'verified/' || id::text || '/' || verified_checksum_sha256 || '.png' and
    verified_byte_size between 1 and 10485760 and
    pixel_width between 1 and 8192 and pixel_height between 1 and 8192 and
    pixel_width::bigint * pixel_height::bigint <= 16000000
  )
) not valid;
