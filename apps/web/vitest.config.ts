import path from "node:path";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "."),
    },
  },
  test: {
    // Most existing tests are pure logic (lib/*.test.ts) and default to
    // "node". Component tests need a DOM, opted into per-file via a
    // `// @vitest-environment jsdom` pragma rather than switching every
    // test to jsdom globally.
    environment: "node",
    setupFiles: ["./vitest.setup.ts"],
  },
});
