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
    <main className="grid min-h-screen place-items-center bg-canvas px-5 py-12 text-ink">
      <SignUp
        fallbackRedirectUrl={`/${locale}/account`}
        signInUrl={`/${locale}/sign-in`}
      />
    </main>
  );
}
