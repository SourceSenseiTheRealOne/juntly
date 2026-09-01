import type { ReactNode } from "react";

import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const clerkState = vi.hoisted(() => ({ signedIn: false }));

vi.mock("@clerk/nextjs", () => ({
  Show: ({ children, when }: { children: ReactNode; when: string }) =>
    (when === "signed-in" ? clerkState.signedIn : !clerkState.signedIn)
      ? children
      : null,
  UserButton: (props: Record<string, unknown>) => {
    if ("afterSignOutUrl" in props) {
      throw new Error("UserButton does not accept afterSignOutUrl");
    }

    return <span>User menu</span>;
  },
}));

import { AuthNavigation } from "./auth-navigation";

const copy = {
  navigationLabel: "Opções da conta",
  signInLabel: "Entrar",
  signInUrl: "/pt-PT/sign-in",
  signUpLabel: "Criar conta",
  signUpUrl: "/pt-PT/sign-up",
};

describe("AuthNavigation", () => {
  it("renders localized touch-safe sign-in and sign-up links for signed-out visitors", () => {
    clerkState.signedIn = false;
    render(<AuthNavigation {...copy} />);

    expect(
      screen.getByRole("navigation", { name: copy.navigationLabel }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: copy.signInLabel }),
    ).toHaveAttribute("href", copy.signInUrl);
    expect(
      screen.getByRole("link", { name: copy.signUpLabel }),
    ).toHaveAttribute("href", copy.signUpUrl);
    expect(screen.getByRole("link", { name: copy.signInLabel })).toHaveClass(
      "min-h-11",
    );
    expect(screen.queryByText("User menu")).not.toBeInTheDocument();
  });

  it("renders only the account menu for signed-in visitors without an unsupported redirect prop", () => {
    clerkState.signedIn = true;
    render(<AuthNavigation {...copy} />);

    expect(
      screen.queryByRole("link", { name: copy.signInLabel }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: copy.signUpLabel }),
    ).not.toBeInTheDocument();
    expect(screen.getByText("User menu")).toBeInTheDocument();
  });
});
