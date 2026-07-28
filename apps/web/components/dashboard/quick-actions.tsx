import { FileDown, FolderPlus, Siren, UserPlus } from "lucide-react";

import { Button } from "@/components/ui/button";

import { SectionCard } from "./section-card";

const actions = [
  { label: "Create project", description: "Set up a new service workspace", icon: FolderPlus },
  { label: "Create incident", description: "Start an incident response", icon: Siren },
  { label: "Export audit logs", description: "Download a compliant activity report", icon: FileDown },
  { label: "Invite team member", description: "Give a teammate secure access", icon: UserPlus },
];

/** Frequently used operator workflows, kept alongside current activity. */
export function QuickActions() {
  return (
    <SectionCard title="Quick actions" description="Common workflows for your platform">
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
        {actions.map(({ label, description, icon: Icon }) => (
          <Button key={label} type="button" variant="secondary" className="h-auto w-full justify-start rounded-2xl px-4 py-3.5 text-left hover:-translate-y-0.5">
            <span className="rounded-xl p-2" style={{ backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)", color: "var(--primary)" }}><Icon aria-hidden="true" className="h-4 w-4" /></span>
            <span className="min-w-0"><span className="block text-sm font-semibold">{label}</span><span className="mt-0.5 block text-xs font-normal" style={{ color: "var(--muted-foreground)" }}>{description}</span></span>
          </Button>
        ))}
      </div>
    </SectionCard>
  );
}
