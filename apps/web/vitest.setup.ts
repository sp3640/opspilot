import { afterEach } from "vitest";
import { cleanup } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";

// vitest.config.ts doesn't enable `test.globals`, so
// @testing-library/react's own auto-cleanup (which detects a global
// `afterEach`) never registers - wire it up explicitly instead, or DOM from
// one component test leaks into the next.
afterEach(() => {
  cleanup();
});

// jsdom doesn't implement HTMLDialogElement.showModal()/close() (used by
// CommandPalette's native <dialog>) - polyfill just enough of the `open`
// attribute/state toggling for tests to exercise real open/close behavior,
// since this environment only has `HTMLDialogElement` when running with
// `// @vitest-environment jsdom`.
if (typeof HTMLDialogElement !== "undefined") {
  if (!HTMLDialogElement.prototype.showModal) {
    HTMLDialogElement.prototype.showModal = function showModal(this: HTMLDialogElement) {
      this.setAttribute("open", "");
    };
  }
  if (!HTMLDialogElement.prototype.close) {
    HTMLDialogElement.prototype.close = function close(this: HTMLDialogElement) {
      const wasOpen = this.hasAttribute("open");
      this.removeAttribute("open");
      if (wasOpen) {
        this.dispatchEvent(new Event("close"));
      }
    };
  }
}
