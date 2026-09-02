import { getTranslations } from "next-intl/server";

import { LandingShell } from "@/features/landing/components/landing-shell";

export default async function HomePage({ params }: PageProps<"/[locale]">) {
  const { locale } = await params;
  const t = await getTranslations("Landing");

  return (
    <LandingShell
      tagline={t("tagline")}
      heading={t("heading")}
      description={t("description")}
      visionLinkLabel={t("visionLinkLabel")}
      discoverLinkLabel={t("discoverLinkLabel")}
      discoverUrl={`/${locale}/discover`}
      accountUrl={`/${locale}/account`}
      signUpUrl={`/${locale}/sign-up`}
      visionTitle={t("visionTitle")}
      visionDescription={t("visionDescription")}
      showcaseTitle={t("showcaseTitle")}
      howTitle={t("how.title")}
      howDescription={t("how.description")}
      discoverBlock={{
        title: t("how.discover.title"),
        description: t("how.discover.description"),
      }}
      compareBlock={{
        title: t("how.compare.title"),
        description: t("how.compare.description"),
      }}
      contactBlock={{
        title: t("how.contact.title"),
        description: t("how.contact.description"),
      }}
      customerBlock={{
        title: t("audience.customer.title"),
        description: t("audience.customer.description"),
        action: t("audience.customer.action"),
      }}
      providerBlock={{
        title: t("audience.provider.title"),
        description: t("audience.provider.description"),
        action: t("audience.provider.action"),
        imageAlt: t("audience.provider.imageAlt"),
      }}
      trustTitle={t("trust.title")}
      trustDescription={t("trust.description")}
      privacyBlock={{
        title: t("trust.privacy.title"),
        description: t("trust.privacy.description"),
      }}
      localBlock={{
        title: t("trust.local.title"),
        description: t("trust.local.description"),
      }}
      reputationBlock={{
        title: t("trust.reputation.title"),
        description: t("trust.reputation.description"),
      }}
      closingTitle={t("closing.title")}
      closingDescription={t("closing.description")}
      closingAction={t("closing.action")}
      footerLabel={t("footerLabel")}
    />
  );
}
