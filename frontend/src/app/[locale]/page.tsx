import { getTranslations } from "next-intl/server";

import { LandingShell } from "@/features/landing/components/landing-shell";

export default async function HomePage() {
  const t = await getTranslations("Landing");

  return (
    <LandingShell
      eyebrow={t("eyebrow")}
      tagline={t("tagline")}
      heading={t("heading")}
      description={t("description")}
      statusLabel={t("statusLabel")}
      visionLinkLabel={t("visionLinkLabel")}
      visionTitle={t("visionTitle")}
      visionDescription={t("visionDescription")}
      footerLabel={t("footerLabel")}
    />
  );
}
