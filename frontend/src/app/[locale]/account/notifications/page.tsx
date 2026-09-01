import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";

import { requireAuthenticatedUser } from "@/features/auth/require-session";
import { NotificationsPanel } from "@/features/messaging/notifications-panel";
import { routing } from "@/i18n/routing";

export const dynamic = "force-dynamic";
export default async function NotificationsPage({
  params,
}: PageProps<"/[locale]/account/notifications">) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) notFound();
  await requireAuthenticatedUser(locale);
  const t = await getTranslations({ locale, namespace: "Notifications" });
  return (
    <main className="market-page px-4 py-8 sm:px-6 sm:py-10">
      <div className="market-panel mx-auto w-full max-w-4xl p-6 sm:p-8">
        <NotificationsPanel
          copy={{
            title: t("title"),
            description: t("description"),
            empty: t("empty"),
            loading: t("loading"),
            error: t("error"),
            inApp: t("inApp"),
            email: t("email"),
            save: t("save"),
            saved: t("saved"),
            markRead: t("markRead"),
            conversationStarted: t("conversationStarted"),
            messageReceived: t("messageReceived"),
            conversationReported: t("conversationReported"),
            requestPublished: t("requestPublished"),
            proposalReceived: t("proposalReceived"),
            proposalAccepted: t("proposalAccepted"),
            proposalRejected: t("proposalRejected"),
            bookingCreated: t("bookingCreated"),
            bookingUpdated: t("bookingUpdated"),
            reviewReceived: t("reviewReceived"),
            reviewResponse: t("reviewResponse"),
            subscriptionUpdated: t("subscriptionUpdated"),
            promotionUpdated: t("promotionUpdated"),
          }}
        />
      </div>
    </main>
  );
}
