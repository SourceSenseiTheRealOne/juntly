import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";
import { requireAuthenticatedUser } from "@/features/auth/require-session";
import { AdministrationDashboardView } from "@/features/administration/administration-dashboard";
import { routing } from "@/i18n/routing";
export const dynamic = "force-dynamic";
export default async function AdministrationPage({
  params,
}: PageProps<"/[locale]/admin">) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) notFound();
  await requireAuthenticatedUser(locale);
  const t = await getTranslations({ locale, namespace: "Administration" });
  return (
    <main className="market-page px-4 py-8 sm:px-6 sm:py-10">
      <div className="market-panel mx-auto w-full max-w-6xl p-6 sm:p-8">
        <AdministrationDashboardView
          copy={{
            title: t("title"),
            description: t("description"),
            loading: t("loading"),
            error: t("error"),
            users: t("users"),
            providers: t("providers"),
            listings: t("listings"),
            bookings: t("bookings"),
            reviews: t("reviews"),
            reports: t("reports"),
            reportQueue: t("reportQueue"),
            reviewQueue: t("reviewQueue"),
            empty: t("empty"),
            reason: t("reason"),
            resolve: t("resolve"),
            hide: t("hide"),
            publish: t("publish"),
            saved: t("saved"),
          }}
        />
      </div>
    </main>
  );
}
