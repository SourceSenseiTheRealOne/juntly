import { getTranslations } from "next-intl/server";
import { hasLocale } from "next-intl";
import { notFound } from "next/navigation";

import { AccountCapabilitiesCard } from "@/features/account/account-capabilities-card";
import { requireAuthenticatedUser } from "@/features/auth/require-session";
import { routing } from "@/i18n/routing";

export const dynamic = "force-dynamic";

type AccountPageProps = {
  params: Promise<{ locale: string }>;
};

export default async function AccountPage({ params }: AccountPageProps) {
  const { locale } = await params;

  if (!hasLocale(routing.locales, locale)) {
    notFound();
  }

  await requireAuthenticatedUser(locale);
  const t = await getTranslations("Account");
  const capabilityCopy = {
    title: t("capabilities.title"),
    description: t("capabilities.description"),
    customerLabel: t("capabilities.customerLabel"),
    customerDescription: t("capabilities.customerDescription"),
    providerLabel: t("capabilities.providerLabel"),
    providerDescription: t("capabilities.providerDescription"),
    enabled: t("capabilities.enabled"),
    disabled: t("capabilities.disabled"),
    loading: t("capabilities.loading"),
    saving: t("capabilities.saving"),
    loadError: t("capabilities.loadError"),
    retry: t("capabilities.retry"),
    manageProvider: t("capabilities.manageProvider"),
    manageListings: t("capabilities.manageListings"),
  };

  return (
    <main className="market-page px-4 py-8 sm:px-6 sm:py-10">
      <section className="market-panel mx-auto w-full max-w-3xl p-6 sm:p-8">
        <p className="text-sm font-semibold tracking-[0.16em] text-accent uppercase">
          Juntly
        </p>
        <h1 className="mt-3 text-4xl font-bold tracking-[-0.05em]">
          {t("title")}
        </h1>
        <p className="mt-4 text-lg leading-8 text-muted">{t("description")}</p>
        <AccountCapabilitiesCard
          copy={capabilityCopy}
          providerProfileUrl={`/${locale}/account/provider-profile`}
          listingsUrl={`/${locale}/account/listings`}
        />
        <nav
          aria-label={t("title")}
          className="mt-6 grid gap-3 border-t border-line pt-6 sm:grid-cols-2"
        >
          <a
            className="market-button-secondary"
            href={`/${locale}/account/messages`}
          >
            {t("capabilities.manageMessages")}
          </a>
          <a
            className="market-button-secondary"
            href={`/${locale}/account/notifications`}
          >
            {t("capabilities.manageNotifications")}
          </a>
        </nav>
      </section>
    </main>
  );
}
