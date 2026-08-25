import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";

import { requireAuthenticatedUser } from "@/features/auth/require-session";
import { ProviderProfileForm } from "@/features/provider/provider-profile-form";
import { routing } from "@/i18n/routing";

export const dynamic = "force-dynamic";
export default async function ProviderProfilePage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) notFound();
  await requireAuthenticatedUser(locale);
  const t = await getTranslations("ProviderProfile");
  const keys = [
    "title",
    "description",
    "displayName",
    "providerType",
    "individual",
    "professional",
    "business",
    "bio",
    "primaryLocality",
    "serviceLocalities",
    "languages",
    "travelRadius",
    "travels",
    "receives",
    "remote",
    "save",
    "saving",
    "loading",
    "error",
    "retry",
    "saved",
  ] as const;
  const copy = Object.fromEntries(keys.map((key) => [key, t(key)])) as Record<
    (typeof keys)[number],
    string
  >;
  return (
    <main className="min-h-screen bg-canvas px-5 py-10 text-ink">
      <div className="mx-auto w-full max-w-3xl rounded-3xl border border-line bg-surface p-6 shadow-sm sm:p-8">
        <ProviderProfileForm locale={locale} copy={copy} />
      </div>
    </main>
  );
}
