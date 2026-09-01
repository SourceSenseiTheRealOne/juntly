import { hasLocale } from "next-intl";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";
import { requireAuthenticatedUser } from "@/features/auth/require-session";
import { EntitlementsDashboard } from "@/features/entitlements/entitlements-dashboard";
import { routing } from "@/i18n/routing";
export const dynamic="force-dynamic";
export default async function EntitlementsPage({params}:PageProps<"/[locale]/account/entitlements">){const{locale}=await params;if(!hasLocale(routing.locales,locale))notFound();await requireAuthenticatedUser(locale);const t=await getTranslations({locale,namespace:"Entitlements"});return <main className="market-page px-4 py-8 sm:px-6 sm:py-10"><div className="market-panel mx-auto w-full max-w-5xl p-6 sm:p-8"><EntitlementsDashboard copy={{title:t("title"),description:t("description"),access:t("access"),activeListings:t("activeListings"),photos:t("photos"),analytics:t("analytics"),enabled:t("enabled"),disabled:t("disabled"),plans:t("plans"),choosePlan:t("choosePlan"),promotion:t("promotion"),listingId:t("listingId"),period:t("period"),promote:t("promote"),current:t("current"),none:t("none"),loading:t("loading"),error:t("error"),requested:t("requested"),pending:t("pending"),active:t("active"),cancelled:t("cancelled"),expired:t("expired")}}/></div></main>}
