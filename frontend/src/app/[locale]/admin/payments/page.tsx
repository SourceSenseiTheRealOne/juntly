import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";

import { requireSoleAdministrator } from "@/features/auth/sole-administrator";
import { PaymentAdministration } from "@/features/payments/payment-administration";
import { routing } from "@/i18n/routing";

export const dynamic = "force-dynamic";

export default async function AdministrativePaymentsPage({
  params,
}: PageProps<"/[locale]/admin/payments">) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) notFound();
  await requireSoleAdministrator(locale);
  const t = await getTranslations({
    locale,
    namespace: "PaymentAdministration",
  });
  return (
    <main className="market-page px-4 py-8 sm:px-6 sm:py-10">
      <div className="market-panel mx-auto w-full max-w-6xl p-6 sm:p-8">
        <PaymentAdministration
          locale={locale}
          copy={{
            title: t("title"),
            description: t("description"),
            loading: t("loading"),
            error: t("error"),
            empty: t("empty"),
            booking: t("booking"),
            gross: t("gross"),
            fee: t("fee"),
            providerNet: t("providerNet"),
            refund: t("refund"),
            refundConfirm: t("refundConfirm"),
            refundPending: t("refundPending"),
            state: t("state"),
          }}
        />
      </div>
    </main>
  );
}
