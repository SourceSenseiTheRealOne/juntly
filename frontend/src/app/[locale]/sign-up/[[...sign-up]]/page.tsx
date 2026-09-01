import { SignUp } from "@clerk/nextjs";
import { hasLocale } from "next-intl";
import { notFound } from "next/navigation";

import { routing } from "@/i18n/routing";

type AuthPageProps = {
  params: Promise<{ locale: string }>;
};

export default async function SignUpPage({ params }: AuthPageProps) {
  const { locale } = await params;

  if (!hasLocale(routing.locales, locale)) {
    notFound();
  }

  return (
    <main className="market-page grid place-items-center px-5 py-12">
      <SignUp
        fallbackRedirectUrl={`/${locale}/account`}
        signInUrl={`/${locale}/sign-in`}
      />
    </main>
  );
}
