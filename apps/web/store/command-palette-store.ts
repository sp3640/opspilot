import { create } from "zustand";

/** Shared open state so the header's search control and the global Cmd+K
 * shortcut both drive the same CommandPalette dialog instance. */
interface CommandPaletteStore {
  isOpen: boolean;
  open: () => void;
  close: () => void;
}

export const useCommandPaletteStore = create<CommandPaletteStore>((set) => ({
  isOpen: false,
  open: () => set({ isOpen: true }),
  close: () => set({ isOpen: false }),
}));
