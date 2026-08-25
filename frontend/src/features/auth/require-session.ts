import { auth } from "@clerk/nextjs/server";
import { redirect } from "next/navigation";

import type { AppLocale } from "@/i18n/routing";

export async function requireAuthenticatedUser(
  locale: AppLocale,
): Promise<string> {
  const { isAuthenticated, userId } = await auth();

  if (!isAuthenticated || !userId) {
    redirect(`/${locale}/sign-in`);
  }

  return userId;
}
