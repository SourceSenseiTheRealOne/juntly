import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

type PackageManifest = {
  scripts?: Record<string, string>;
};

function readPackageManifest(): PackageManifest {
  return JSON.parse(
    readFileSync(resolve(process.cwd(), "package.json"), "utf8"),
  ) as PackageManifest;
}

describe("local browser origin", () => {
  it("provides localhost-bound development and production launch scripts", () => {
    const manifest = readPackageManifest();

    expect(manifest.scripts?.["dev:local"]).toBe(
      "next dev --hostname localhost --port 4200",
    );
    expect(manifest.scripts?.["start:local"]).toBe(
      "next start --hostname localhost --port 4200",
    );
  });
});
