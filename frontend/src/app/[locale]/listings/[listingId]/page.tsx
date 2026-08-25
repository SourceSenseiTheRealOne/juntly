import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";

import { PublicListingDetail } from "@/features/discovery/public-listing-detail";
import { routing } from "@/i18n/routing";

export const dynamic = "force-dynamic";

export default async function PublicListingPage({
  params,
}: PageProps<"/[locale]/listings/[listingId]">) {
  const { locale, listingId } = await params;
  if (!hasLocale(routing.locales, locale) || !uuid(listingId)) notFound();
  const t = await getTranslations({ locale, namespace: "PublicListing" });
  return (
    <main className="mx-auto w-full max-w-3xl px-5 py-10">
      <PublicListingDetail
        locale={locale}
        listingId={listingId}
        copy={{
          loading: t("loading"),
          error: t("error"),
          retry: t("retry"),
          provider: t("provider"),
          locality: t("locality"),
          category: t("category"),
          phone: t("phone"),
          whatsapp: t("whatsapp"),
          revealError: t("revealError"),
        }}
      />
    </main>
  );
}

function uuid(value: string): boolean {
  return /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(value);
}
