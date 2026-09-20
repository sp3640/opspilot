"use client";

import { useEffect, type RefObject } from "react";

/**
 * Closes an open popover/dropdown when the user clicks outside `ref`'s
 * element or presses Escape - the standard dismissal behavior for any
 * header popover (notifications, user menu). Only attaches listeners while
 * `open` is true, so it costs nothing when closed.
 */
export function useDismissable(ref: RefObject<HTMLElement | null>, open: boolean, onDismiss: () => void) {
  useEffect(() => {
    if (!open) return;

    const handlePointerDown = (event: MouseEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        onDismiss();
      }
    };
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onDismiss();
      }
    };

    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [open, ref, onDismiss]);
}
