import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";
import { requireAuthenticatedUser } from "@/features/auth/require-session";
import { ListingDashboard } from "@/features/listings/listing-dashboard";
import { routing } from "@/i18n/routing";
export const dynamic = "force-dynamic";
export default async function ListingsPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) notFound();
  await requireAuthenticatedUser(locale);
  const t = await getTranslations("Listings");
  const keys = [
    "title",
    "description",
    "newListing",
    "create",
    "submit",
    "pause",
    "archive",
    "loading",
    "error",
    "retry",
    "empty",
    "saved",
    "titleLabel",
    "descriptionLabel",
    "categoryLabel",
    "localityLabel",
    "priceLabel",
  ] as const;
  const copy = Object.fromEntries(keys.map((k) => [k, t(k)])) as Record<
    (typeof keys)[number],
    string
  >;
  return (
    <main className="min-h-screen bg-canvas px-5 py-10 text-ink">
      <div className="mx-auto w-full max-w-3xl rounded-3xl border border-line bg-surface p-6 shadow-sm sm:p-8">
        <ListingDashboard
          copy={copy}
          categories={[]}
          localities={[]}
          locale={locale}
        />
      </div>
    </main>
  );
}
