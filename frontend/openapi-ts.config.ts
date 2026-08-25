import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../openapi/juntly-api.v1.yaml",
  output: "src/shared/api/generated",
  plugins: ["@hey-api/typescript", "@hey-api/sdk", "@hey-api/client-fetch"],
});
