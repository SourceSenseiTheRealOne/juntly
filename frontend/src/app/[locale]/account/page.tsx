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
    <main className="market-page px-4 py-8 sm:px-6 sm:py-12">
      <section className="market-panel mx-auto w-full max-w-6xl overflow-hidden p-6 sm:p-8 lg:p-10">
        <div className="market-page-header border-b border-line pb-8">
          <p className="market-kicker">Juntly</p>
          <h1 className="text-4xl font-bold tracking-[-0.055em] sm:text-5xl">
            {t("title")}
          </h1>
          <p>{t("description")}</p>
        </div>
        <div className="grid gap-10 lg:grid-cols-[1.1fr_0.9fr] lg:gap-12">
          <div>
            <AccountCapabilitiesCard
              copy={capabilityCopy}
              providerProfileUrl={`/${locale}/account/provider-profile`}
              listingsUrl={`/${locale}/account/listings`}
            />
          </div>
          <nav
            aria-label={t("title")}
            className="grid content-start gap-3 border-t border-line pt-8 lg:border-t-0 lg:border-l lg:pl-10"
          >
            <a
              className="market-button-secondary justify-between px-4"
              href={`/${locale}/account/messages`}
            >
              {t("capabilities.manageMessages")}
            </a>
            <a
              className="market-button-secondary justify-between px-4"
              href={`/${locale}/account/notifications`}
            >
              {t("capabilities.manageNotifications")}
            </a>
            <a
              className="market-button-secondary justify-between px-4"
              href={`/${locale}/account/quotations`}
            >
              {t("capabilities.manageQuotations")}
            </a>
            <a
              className="market-button-secondary justify-between px-4"
              href={`/${locale}/account/bookings`}
            >
              {t("capabilities.manageBookings")}
            </a>
            <a
              className="market-button-secondary justify-between px-4"
              href={`/${locale}/account/reviews`}
            >
              {t("capabilities.manageReviews")}
            </a>
            <a
              className="market-button-secondary justify-between px-4"
              href={`/${locale}/account/entitlements`}
            >
              {t("capabilities.manageEntitlements")}
            </a>
          </nav>
        </div>
      </section>
    </main>
  );
}
