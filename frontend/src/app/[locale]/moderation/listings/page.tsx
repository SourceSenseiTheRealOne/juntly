import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";
import { requireAuthenticatedUser } from "@/features/auth/require-session";
import { ModerationQueue } from "@/features/listings/moderation-queue";
import { routing } from "@/i18n/routing";
export const dynamic = "force-dynamic";
export default async function ModerationListingsPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) notFound();
  await requireAuthenticatedUser(locale);
  const t = await getTranslations("Moderation");
  const keys = [
    "title",
    "loading",
    "empty",
    "error",
    "retry",
    "approve",
    "reject",
  ] as const;
  const copy = Object.fromEntries(keys.map((k) => [k, t(k)])) as Record<
    (typeof keys)[number],
    string
  >;
  return (
    <main className="market-page px-4 py-8 sm:px-6 sm:py-10">
      <div className="market-panel mx-auto w-full max-w-4xl p-6 sm:p-8">
        <ModerationQueue copy={copy} />
      </div>
    </main>
  );
}
