import { getLocale, getTranslations } from "next-intl/server";

import { LandingShell } from "@/features/landing/components/landing-shell";

export default async function HomePage() {
  const locale = await getLocale();
  const auth = await getTranslations("Auth");
  const t = await getTranslations("Landing");

  return (
    <LandingShell
      eyebrow={t("eyebrow")}
      tagline={t("tagline")}
      heading={t("heading")}
      description={t("description")}
      statusLabel={t("statusLabel")}
      healthStatusLabels={{
        checking: t("healthStatus.checking"),
        available: t("healthStatus.available"),
        unavailable: t("healthStatus.unavailable"),
      }}
      signInLabel={auth("signIn")}
      signInUrl={`/${locale}/sign-in`}
      signUpLabel={auth("signUp")}
      signUpUrl={`/${locale}/sign-up`}
      visionLinkLabel={t("visionLinkLabel")}
      visionTitle={t("visionTitle")}
      visionDescription={t("visionDescription")}
      footerLabel={t("footerLabel")}
    />
  );
}
