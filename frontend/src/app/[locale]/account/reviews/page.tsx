import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";
import { requireAuthenticatedUser } from "@/features/auth/require-session";
import { ReviewsDashboard } from "@/features/reviews/reviews-dashboard";
import { routing } from "@/i18n/routing";
export const dynamic = "force-dynamic";
export default async function ReviewsPage({
  params,
}: PageProps<"/[locale]/account/reviews">) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) notFound();
  await requireAuthenticatedUser(locale);
  const t = await getTranslations({ locale, namespace: "Reviews" });
  return (
    <main className="market-page px-4 py-8 sm:px-6 sm:py-10">
      <div className="market-panel mx-auto w-full max-w-5xl p-6 sm:p-8">
        <ReviewsDashboard
          copy={{
            title: t("title"),
            description: t("description"),
            newReview: t("newReview"),
            bookingId: t("bookingId"),
            rating: t("rating"),
            body: t("body"),
            submit: t("submit"),
            received: t("received"),
            empty: t("empty"),
            response: t("response"),
            respond: t("respond"),
            loading: t("loading"),
            error: t("error"),
            created: t("created"),
            verified: t("verified"),
          }}
        />
      </div>
    </main>
  );
}
