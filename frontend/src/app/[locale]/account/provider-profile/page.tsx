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
    "serviceModes",
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
    <main className="market-page px-4 py-8 sm:px-6 sm:py-10">
      <div className="market-panel mx-auto w-full max-w-3xl p-6 sm:p-8">
        <ProviderProfileForm locale={locale} copy={copy} />
      </div>
    </main>
  );
}
