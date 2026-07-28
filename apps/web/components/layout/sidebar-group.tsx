type SidebarGroupProps = {
  title: string;
  children: React.ReactNode;
};

export function SidebarGroup({
  title,
  children,
}: SidebarGroupProps) {
  return (
    <div className="space-y-3">
      <h3 className="px-4 text-[11px] font-semibold uppercase tracking-[0.2em] text-zinc-500">
        {title}
      </h3>

      <div className="space-y-1">{children}</div>
    </div>
  );
}