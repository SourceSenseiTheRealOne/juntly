import { readdirSync, readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

const sourceRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const globalsCss = readFileSync(resolve(sourceRoot, "app/globals.css"), "utf8");
function filesWithExtension(directory: string, extension: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);

    if (entry.isDirectory()) {
      return filesWithExtension(path, extension);
    }

    return entry.isFile() && entry.name.endsWith(extension) ? [path] : [];
  });
}

const componentSources = filesWithExtension(sourceRoot, ".tsx").map((path) => ({
  path,
  source: readFileSync(path, "utf8"),
}));

const semanticColorTokens = [
  "--color-bg-canvas",
  "--color-bg-surface",
  "--color-bg-elevated",
  "--color-text-primary",
  "--color-text-secondary",
  "--color-text-inverse",
  "--color-border-default",
  "--color-border-strong",
  "--color-action-primary",
  "--color-action-primary-hover",
  "--color-action-primary-soft",
  "--color-accent-earth",
  "--color-accent-earth-soft",
  "--color-focus-ring",
  "--color-status-success",
  "--color-status-warning",
  "--color-status-danger",
  "--color-status-info",
  "--color-shadow-soft",
  "--color-shadow-strong",
] as const;

const responsiveTokens = [
  "--layout-container-max",
  "--layout-gutter",
  "--layout-section-space",
  "--layout-content-gap",
  "--layout-hero-min-height",
  "--size-touch-target",
  "--size-brand-orbit",
  "--type-display",
  "--type-section-title",
  "--type-lead",
  "--type-leading-display",
  "--type-tracking-brand",
  "--type-tracking-display",
  "--type-tracking-heading",
  "--type-tracking-label",
  "--shape-control-radius",
  "--shape-pill-radius",
  "--elevation-surface",
  "--elevation-elevated",
] as const;

function declarationsFor(selector: string) {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const match = globalsCss.match(
    new RegExp(`${escapedSelector}\\s*\\{(?<body>[\\s\\S]*?)\\}`),
  );

  return match?.groups?.body ?? "";
}

function tokenValue(declarations: string, token: string) {
  const escapedToken = token.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  return declarations.match(new RegExp(`${escapedToken}:\\s*([^;]+);`))?.[1];
}

describe("design token contract", () => {
  it("defines the complete semantic color and responsive token API", () => {
    for (const token of [...semanticColorTokens, ...responsiveTokens]) {
      expect(globalsCss, `missing ${token}`).toContain(`${token}:`);
    }

    expect(globalsCss).toContain("clamp(");
  });

  it("supports automatic and explicit light and dark color schemes", () => {
    expect(globalsCss).toContain("@media (prefers-color-scheme: dark)");
    expect(globalsCss).toContain('[data-theme="light"]');
    expect(globalsCss).toContain('[data-theme="dark"]');

    const fallbackLightTheme = declarationsFor(":root");
    const lightTheme = declarationsFor('[data-theme="light"]');
    const systemDarkTheme = declarationsFor(':root:not([data-theme="light"])');
    const darkTheme = declarationsFor('[data-theme="dark"]');

    for (const token of semanticColorTokens) {
      expect(
        fallbackLightTheme,
        `fallback light theme missing ${token}`,
      ).toContain(`${token}:`);
      expect(lightTheme, `light theme missing ${token}`).toContain(`${token}:`);
      expect(systemDarkTheme, `system dark theme missing ${token}`).toContain(
        `${token}:`,
      );
      expect(darkTheme, `dark theme missing ${token}`).toContain(`${token}:`);
      expect(
        tokenValue(lightTheme, token),
        `${token} light override drifted`,
      ).toBe(tokenValue(fallbackLightTheme, token));
      expect(
        tokenValue(systemDarkTheme, token),
        `${token} system dark and explicit dark diverged`,
      ).toBe(tokenValue(darkTheme, token));
    }
  });

  it("maps Tailwind theme utilities to the shared semantic variables", () => {
    const theme = declarationsFor("@theme inline");

    expect(theme).toContain("--color-canvas: var(--color-bg-canvas)");
    expect(theme).toContain("--color-ink: var(--color-text-primary)");
    expect(theme).toContain("--spacing-page: var(--layout-gutter)");
    expect(theme).toContain("--spacing-section: var(--layout-section-space)");
    expect(theme).toContain("--container-content: var(--layout-container-max)");
    expect(theme).toContain("--text-display: var(--type-display)");
    expect(theme).toContain("--leading-display: var(--type-leading-display)");
    expect(theme).toContain("--tracking-brand: var(--type-tracking-brand)");
    expect(theme).toContain("--tracking-display: var(--type-tracking-display)");
    expect(theme).toContain("--tracking-heading: var(--type-tracking-heading)");
    expect(theme).toContain("--tracking-label: var(--type-tracking-label)");
    expect(theme).toContain("--radius-control: var(--shape-control-radius)");
    expect(theme).toContain("--shadow-elevated: var(--elevation-elevated)");
  });

  it("keeps raw colors and self references out of semantic tokens", () => {
    const declarations = [
      ...globalsCss.matchAll(/(--[a-z0-9-]+):\s*([^;]+);/gi),
    ];

    for (const [, token, value] of declarations) {
      expect(value, `${token} references itself`).not.toContain(
        `var(${token})`,
      );

      if (!token.startsWith("--palette-")) {
        expect(value, `${token} contains a raw color`).not.toMatch(
          /#[0-9a-f]{3,8}\b|(?:rgb|hsl|oklch)\(/i,
        );
      }
    }
  });

  it("keeps reusable design literals out of all frontend component markup", () => {
    for (const { path, source } of componentSources) {
      expect(source, path).not.toMatch(/#[0-9a-f]{3,8}/i);
      expect(source, path).not.toMatch(/(?:rgb|hsl|oklch)\(/i);
      expect(source, path).not.toMatch(/\b(?:text|bg)-(?:white|black)\b/);
      expect(source, path).not.toContain("shadow-[");
      expect(source, path).not.toContain("min-h-[");
      expect(source, path).not.toContain("tracking-[");
      expect(source, path).not.toContain("leading-[");
      expect(source, path).not.toContain("min-h-11");
      expect(source, path).not.toContain("rounded-full");
    }
  });
});
