import { clerkMiddleware } from "@clerk/nextjs/server";
import { NextResponse } from "next/server";
import createMiddleware from "next-intl/middleware";

import { routing } from "./i18n/routing";

const intlMiddleware = createMiddleware(routing);
const clerkOptions =
  process.env.NODE_ENV === "development"
    ? { clockSkewInMs: 30_000 }
    : undefined;

export default clerkMiddleware(async (_auth, request) => {
  if (
    request.nextUrl.pathname === "/api" ||
    request.nextUrl.pathname.startsWith("/api/") ||
    request.nextUrl.pathname.startsWith("/__clerk/")
  ) {
    return NextResponse.next();
  }

  return intlMiddleware(request);
}, clerkOptions);

export const config = {
  matcher: [
    "/((?!_next|_vercel|.*\\..*).*)",
    "/(api|trpc)(.*)",
    "/__clerk/(.*)",
  ],
};
