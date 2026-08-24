import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";

import { requireAuthenticatedUser } from "@/features/auth/require-session";
import { ContactChannelsCard } from "@/features/contact/contact-channels-card";
import { routing } from "@/i18n/routing";

export const dynamic = "force-dynamic";

export default async function ContactChannelsPage({
  params,
}: PageProps<"/[locale]/account/contact-channels">) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) notFound();
  await requireAuthenticatedUser(locale);
  const t = await getTranslations({ locale, namespace: "ContactChannels" });
  return (
    <main className="mx-auto w-full max-w-3xl px-5 py-10">
      <ContactChannelsCard
        copy={{
          title: t("title"),
          description: t("description"),
          loading: t("loading"),
          error: t("error"),
          retry: t("retry"),
          phone: t("phone"),
          whatsapp: t("whatsapp"),
          contact: t("contact"),
          formatHint: t("formatHint"),
          enabled: t("enabled"),
          consent: t("consent"),
          save: t("save"),
          saved: t("saved"),
        }}
      />
    </main>
  );
}
