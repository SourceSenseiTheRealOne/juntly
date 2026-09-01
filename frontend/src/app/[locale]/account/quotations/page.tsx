import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";
import { requireAuthenticatedUser } from "@/features/auth/require-session";
import { QuotationsDashboard } from "@/features/quotations/quotations-dashboard";
import { routing } from "@/i18n/routing";
export const dynamic = "force-dynamic";
export default async function QuotationsPage({
  params,
}: PageProps<"/[locale]/account/quotations">) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) notFound();
  await requireAuthenticatedUser(locale);
  const t = await getTranslations({ locale, namespace: "Quotations" });
  return (
    <main className="market-page px-4 py-8 sm:px-6 sm:py-10">
      <div className="market-panel mx-auto w-full max-w-7xl p-6 sm:p-8">
        <QuotationsDashboard
          locale={locale}
          copy={{
            title: t("title"),
            description: t("description"),
            customerRequests: t("customerRequests"),
            opportunities: t("opportunities"),
            newRequest: t("newRequest"),
            requestTitle: t("requestTitle"),
            requestDescription: t("requestDescription"),
            category: t("category"),
            locality: t("locality"),
            budget: t("budget"),
            deadline: t("deadline"),
            publish: t("publish"),
            emptyRequests: t("emptyRequests"),
            emptyOpportunities: t("emptyOpportunities"),
            viewProposals: t("viewProposals"),
            proposals: t("proposals"),
            proposalPrice: t("proposalPrice"),
            proposalMessage: t("proposalMessage"),
            availableAt: t("availableAt"),
            estimatedMinutes: t("estimatedMinutes"),
            submitProposal: t("submitProposal"),
            accept: t("accept"),
            loading: t("loading"),
            error: t("error"),
            created: t("created"),
            submitted: t("submitted"),
            accepted: t("accepted"),
          }}
        />
      </div>
    </main>
  );
}
