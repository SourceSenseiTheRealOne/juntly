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
  };

  return (
    <main className="grid min-h-screen place-items-center bg-canvas px-5 py-10 text-ink">
      <section className="w-full max-w-2xl rounded-3xl border border-line bg-surface p-6 shadow-sm sm:p-8">
        <p className="text-sm font-semibold tracking-[0.16em] text-accent uppercase">
          Juntly
        </p>
        <h1 className="mt-4 text-3xl font-semibold tracking-[-0.04em]">
          {t("title")}
        </h1>
        <p className="mt-4 text-lg leading-8 text-muted">{t("description")}</p>
        <AccountCapabilitiesCard copy={capabilityCopy} />
      </section>
    </main>
  );
}
