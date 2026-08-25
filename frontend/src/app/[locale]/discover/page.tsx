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
    <main className="mx-auto w-full max-w-4xl px-5 py-10">
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
        }}
      />
    </main>
  );
}
