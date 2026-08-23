import { Menu } from "lucide-react";

import { SearchBar } from "./search-bar";
import { NotificationButton } from "./notification";
import { UserMenu } from "./user-menu";

type HeaderProps = {
  /** Opens the mobile off-canvas sidebar. The trigger button is only
   * rendered below the `lg:` breakpoint, where the sidebar is hidden by
   * default (see components/layout/sidebar.tsx). */
  onMenuClick?: () => void;
};

export function Header({ onMenuClick }: HeaderProps) {
  return (
    <header className="sticky top-0 z-30 px-4 pt-6 sm:px-6 lg:px-8">
      <div
        className="
          flex
          items-center
          justify-between
          gap-3
          rounded-3xl
          border
          px-4
          py-4
          backdrop-blur-xl
          sm:px-6
        "
        style={{
          background: "rgba(24,24,27,0.75)",
          borderColor: "var(--border)",
          boxShadow: "var(--shadow-md)",
        }}
      >
        <button
          type="button"
          onClick={onMenuClick}
          aria-label="Open navigation menu"
          className="rounded-xl p-2 text-zinc-400 hover:text-white lg:hidden"
        >
          <Menu aria-hidden="true" className="h-5 w-5" />
        </button>

        <SearchBar />

        <div className="flex items-center gap-4">
          <NotificationButton />
          <UserMenu />
        </div>
      </div>
    </header>
  );
}