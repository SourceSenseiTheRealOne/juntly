import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";

import { requireAuthenticatedUser } from "@/features/auth/require-session";
import { PayoutDashboard } from "@/features/payments/payout-dashboard";
import { routing } from "@/i18n/routing";

export const dynamic = "force-dynamic";

export default async function PayoutsPage({
  params,
}: PageProps<"/[locale]/account/payouts">) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) notFound();
  await requireAuthenticatedUser(locale);
  const t = await getTranslations({ locale, namespace: "Payouts" });
  return (
    <main className="market-page px-4 py-8 sm:px-6 sm:py-10">
      <div className="market-panel mx-auto w-full max-w-4xl p-6 sm:p-8">
        <PayoutDashboard
          locale={locale}
          copy={{
            title: t("title"),
            description: t("description"),
            loading: t("loading"),
            unavailable: t("unavailable"),
            notStarted: t("notStarted"),
            ready: t("ready"),
            incomplete: t("incomplete"),
            charges: t("charges"),
            payouts: t("payouts"),
            details: t("details"),
            start: t("start"),
            continue: t("continue"),
          }}
        />
      </div>
    </main>
  );
}
