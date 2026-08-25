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
    <main className="min-h-screen bg-canvas px-5 py-10 text-ink">
      <div className="mx-auto w-full max-w-3xl rounded-3xl border border-line bg-surface p-6 shadow-sm sm:p-8">
        <ModerationQueue copy={copy} />
      </div>
    </main>
  );
}
