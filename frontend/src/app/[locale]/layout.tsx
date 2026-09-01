import { ClerkProvider } from "@clerk/nextjs";
import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { hasLocale, NextIntlClientProvider } from "next-intl";
import { getMessages, getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";

import { getClerkLocalization } from "@/features/auth/clerk-localization";
import { MarketplaceNavigation } from "@/features/navigation/marketplace-navigation";
import { routing } from "@/i18n/routing";

import "../globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }));
}

export async function generateMetadata({
  params,
}: LayoutProps<"/[locale]">): Promise<Metadata> {
  const { locale } = await params;

  if (!hasLocale(routing.locales, locale)) {
    notFound();
  }

  const t = await getTranslations({ locale, namespace: "Metadata" });

  return {
    title: t("title"),
    description: t("description"),
    applicationName: "Juntly",
    alternates: {
      languages: Object.fromEntries(
        routing.locales.map((supportedLocale) => [
          supportedLocale,
          `/${supportedLocale}`,
        ]),
      ),
    },
  };
}

export default async function LocaleLayout({
  children,
  params,
}: LayoutProps<"/[locale]">) {
  const { locale } = await params;

  if (!hasLocale(routing.locales, locale)) {
    notFound();
  }

  const messages = await getMessages({ locale });
  const navigation = await getTranslations({ locale, namespace: "Navigation" });
  const auth = await getTranslations({ locale, namespace: "Auth" });

  return (
    <html
      lang={locale}
      className={`${geistSans.variable} ${geistMono.variable}`}
    >
      <body className="min-h-full antialiased">
        <ClerkProvider
          afterSignOutUrl={`/${locale}`}
          localization={getClerkLocalization(locale)}
          signInFallbackRedirectUrl={`/${locale}/account`}
          signInUrl={`/${locale}/sign-in`}
          signUpFallbackRedirectUrl={`/${locale}/account`}
          signUpUrl={`/${locale}/sign-up`}
        >
          <NextIntlClientProvider messages={messages}>
            <MarketplaceNavigation
              accountLabel={navigation("account")}
              accountUrl={`/${locale}/account`}
              accountNavigationLabel={navigation("accountNavigation")}
              discoverLabel={navigation("discover")}
              discoverUrl={`/${locale}/discover`}
              navigationLabel={navigation("marketplace")}
              signInLabel={auth("signIn")}
              signInUrl={`/${locale}/sign-in`}
              signUpLabel={auth("signUp")}
              signUpUrl={`/${locale}/sign-up`}
            />
            {children}
          </NextIntlClientProvider>
        </ClerkProvider>
      </body>
    </html>
  );
}
