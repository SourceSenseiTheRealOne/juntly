import { SignIn } from "@clerk/nextjs";
import { hasLocale } from "next-intl";
import { notFound } from "next/navigation";

import { routing } from "@/i18n/routing";

type AuthPageProps = {
  params: Promise<{ locale: string }>;
};

export default async function SignInPage({ params }: AuthPageProps) {
  const { locale } = await params;

  if (!hasLocale(routing.locales, locale)) {
    notFound();
  }

  return (
    <main className="market-page grid place-items-center px-4 py-10 sm:px-6 sm:py-16">
      <SignIn
        appearance={{
          variables: {
            colorPrimary: "var(--accent)",
            colorBackground: "var(--surface)",
            borderRadius: "0.875rem",
          },
          elements: {
            rootBox: "w-full max-w-md",
            cardBox:
              "w-full rounded-3xl border border-line shadow-[var(--shadow-card)]",
          },
        }}
        fallbackRedirectUrl={`/${locale}/account`}
        signUpUrl={`/${locale}/sign-up`}
      />
    </main>
  );
}
