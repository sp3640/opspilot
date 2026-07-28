import { SearchBar } from "./search-bar";
import { NotificationButton } from "./notification";
import { UserMenu } from "./user-menu";

export function Header() {
  return (
    <header className="sticky top-0 z-50 px-8 pt-6">
      <div
        className="
          flex
          items-center
          justify-between
          rounded-3xl
          border
          px-6
          py-4
          backdrop-blur-xl
        "
        style={{
          background: "rgba(24,24,27,0.75)",
          borderColor: "var(--border)",
          boxShadow: "var(--shadow-md)",
        }}
      >
        <SearchBar />

        <div className="flex items-center gap-4">
          <NotificationButton />
          <UserMenu />
        </div>
      </div>
    </header>
  );
}