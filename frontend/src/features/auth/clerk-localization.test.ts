import { enUS, esES, ptPT } from "@clerk/localizations";
import { describe, expect, it } from "vitest";

import { getClerkLocalization } from "./clerk-localization";

describe("getClerkLocalization", () => {
  it("maps every supported Juntly locale to the matching Clerk localization", () => {
    expect(getClerkLocalization("pt-PT")).toBe(ptPT);
    expect(getClerkLocalization("en")).toBe(enUS);
    expect(getClerkLocalization("es")).toBe(esES);
  });
});
