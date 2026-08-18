import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
    input: "../backend/cmd/docs/swagger.json",
    output: {
        path: "src/openapi-ts_beta_api", // TODO: Change directory name when migration complete
        entryFile: false,
    },
});
