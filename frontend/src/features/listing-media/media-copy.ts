import en from "../../../messages/en.json";
import es from "../../../messages/es.json";
import pt from "../../../messages/pt-PT.json";

export type ListingPhotosCopy = typeof en.ListingMedia;
const messages = { "pt-PT": pt.ListingMedia, en: en.ListingMedia, es: es.ListingMedia };
export function getMediaCopy(locale: "pt-PT" | "en" | "es"): ListingPhotosCopy {
  return messages[locale];
}
