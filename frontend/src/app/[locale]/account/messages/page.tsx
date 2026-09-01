import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";

import { requireAuthenticatedUser } from "@/features/auth/require-session";
import { MessagingInbox } from "@/features/messaging/messaging-inbox";
import { routing } from "@/i18n/routing";

export const dynamic = "force-dynamic";
export default async function MessagesPage({
  params,
}: PageProps<"/[locale]/account/messages">) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) notFound();
  await requireAuthenticatedUser(locale);
  const t = await getTranslations({ locale, namespace: "Messaging" });
  return (
    <main className="market-page px-4 py-8 sm:px-6 sm:py-10">
      <div className="market-panel mx-auto w-full max-w-5xl p-6 sm:p-8">
        <MessagingInbox
          copy={{
            title: t("title"),
            description: t("description"),
            empty: t("empty"),
            selectConversation: t("selectConversation"),
            messageLabel: t("messageLabel"),
            send: t("send"),
            sending: t("sending"),
            loading: t("loading"),
            error: t("error"),
            conversation: t("conversation"),
          }}
        />
      </div>
    </main>
  );
}
