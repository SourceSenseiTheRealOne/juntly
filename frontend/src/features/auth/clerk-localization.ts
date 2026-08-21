import { enUS, esES, ptPT } from "@clerk/localizations";

import type { AppLocale } from "@/i18n/routing";

export function getClerkLocalization(locale: AppLocale) {
  switch (locale) {
    case "pt-PT":
      return ptPT;
    case "en":
      return enUS;
    case "es":
      return esES;
  }
}
