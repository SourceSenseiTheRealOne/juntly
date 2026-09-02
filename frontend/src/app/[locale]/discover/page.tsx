import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";

import { PublicDiscovery } from "@/features/discovery/public-discovery";
import { routing } from "@/i18n/routing";

export const dynamic = "force-dynamic";

export default async function DiscoverPage({
  params,
}: PageProps<"/[locale]/discover">) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) notFound();
  const t = await getTranslations({ locale, namespace: "Discovery" });
  return (
    <main className="market-page px-4 py-8 sm:px-6 sm:py-10">
      <div className="market-container">
        <PublicDiscovery
          locale={locale}
          copy={{
            title: t("title"),
            description: t("description"),
            loading: t("loading"),
            empty: t("empty"),
            error: t("error"),
            retry: t("retry"),
            searchLabel: t("searchLabel"),
            searchButton: t("searchButton"),
            categoryLabel: t("categoryLabel"),
            localityLabel: t("localityLabel"),
            radiusLabel: t("radiusLabel"),
            priceLabel: t("priceLabel"),
            modeLabel: t("modeLabel"),
            details: t("details"),
            marketplaceLabel: t("marketplaceLabel"),
            locationContextLabel: t("locationContextLabel"),
            filtersLabel: t("filtersLabel"),
            promoted: t("promoted"),
            allCategories: t("allCategories"),
            allLocalities: t("allLocalities"),
            anyPrice: t("anyPrice"),
            anyMode: t("anyMode"),
            priceFixed: t("priceFixed"),
            priceHourly: t("priceHourly"),
            priceDaily: t("priceDaily"),
            priceQuote: t("priceQuote"),
            priceNegotiable: t("priceNegotiable"),
            modeTravels: t("modeTravels"),
            modeReceives: t("modeReceives"),
            modeRemote: t("modeRemote"),
            applyFilters: t("applyFilters"),
          }}
        />
      </div>
    </main>
  );
}
