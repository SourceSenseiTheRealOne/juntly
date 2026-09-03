import { auth, currentUser } from "@clerk/nextjs/server";
import { notFound, redirect } from "next/navigation";

import type { AppLocale } from "@/i18n/routing";

export const SOLE_ADMIN_EMAIL = "source.sensei1205@gmail.com";

type EmailIdentity = {
  emailAddress: string;
  verification?: { status?: string | null } | null;
};

type UserIdentity = {
  id: string;
  emailAddresses: EmailIdentity[];
};

type SoleAdministratorSession =
  | { status: "unauthenticated" }
  | { status: "forbidden" }
  | { status: "unavailable" }
  | { status: "authorized"; token: string; userId: string };

export function isSoleAdministrator(
  sessionUserId: string,
  user: UserIdentity | null,
): boolean {
  if (!user || user.id !== sessionUserId) return false;

  return user.emailAddresses.some(
    (identity) =>
      identity.verification?.status === "verified" &&
      identity.emailAddress.trim().toLowerCase() === SOLE_ADMIN_EMAIL,
  );
}

export async function resolveSoleAdministratorSession(): Promise<SoleAdministratorSession> {
  try {
    const session = await auth();
    if (!session.isAuthenticated || !session.userId) {
      return { status: "unauthenticated" };
    }

    const user = await currentUser();
    if (!isSoleAdministrator(session.userId, user)) {
      return { status: "forbidden" };
    }

    const token = await session.getToken();
    if (!token) return { status: "unauthenticated" };

    return { status: "authorized", token, userId: session.userId };
  } catch {
    return { status: "unavailable" };
  }
}

export async function requireSoleAdministrator(
  locale: AppLocale,
): Promise<string> {
  const session = await resolveSoleAdministratorSession();
  if (session.status === "unauthenticated") {
    redirect(`/${locale}/sign-in`);
  }
  if (session.status !== "authorized") {
    notFound();
  }
  return session.userId;
}

export async function currentUserIsSoleAdministrator(
  sessionUserId: string,
): Promise<boolean> {
  try {
    return isSoleAdministrator(sessionUserId, await currentUser());
  } catch {
    return false;
  }
}
