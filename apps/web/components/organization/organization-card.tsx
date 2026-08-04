import type { ComponentType } from "react";
import { Building2, CalendarClock, FolderKanban, Hash } from "lucide-react";

import { StatusBadge } from "@/components/common";
import { SectionCard } from "@/components/dashboard";
import type { OrganizationResponse } from "@/types/organization-api";

type OrganizationCardProps = {
  organization: OrganizationResponse;
};

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";

  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "2-digit",
    year: "numeric",
  }).format(date);
}

export function OrganizationCard({ organization }: OrganizationCardProps) {
  const projectCount = organization.projectCount ?? organization.projectsCount;

  return (
    <SectionCard title="Workspace snapshot" description="Current organization metadata and ownership information.">
      <div className="space-y-4">
        <div className="flex items-center justify-between gap-4 rounded-2xl border p-4" style={{ borderColor: "var(--border)" }}>
          <div className="flex items-center gap-2">
            <Building2 aria-hidden={true} className="h-4 w-4" />
            <span className="text-sm font-medium">Slug</span>
          </div>
          <StatusBadge variant="info">{organization.slug}</StatusBadge>
        </div>

        <CardRow icon={Hash} label="Organization ID" value={organization.id} />
        <CardRow icon={Hash} label="Owner ID" value={String(organization.ownerId)} />
        <CardRow icon={CalendarClock} label="Created" value={formatDate(organization.createdAt)} />
        <CardRow icon={CalendarClock} label="Updated" value={formatDate(organization.updatedAt)} />

        {typeof projectCount === "number" && (
          <CardRow icon={FolderKanban} label="Projects" value={String(projectCount)} />
        )}
      </div>
    </SectionCard>
  );
}

function CardRow({
  icon: Icon,
  label,
  value,
}: {
  icon: ComponentType<{ className?: string; "aria-hidden"?: boolean }>;
  label: string;
  value: string;
}) {
  return (
    <div
      className="rounded-2xl border p-4"
      style={{
        borderColor: "var(--border)",
        backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)",
      }}
    >
      <div className="flex items-center gap-2">
        <Icon aria-hidden={true} className="h-4 w-4" />
        <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
          {label}
        </p>
      </div>
      <p className="mt-1 break-all font-medium">{value}</p>
    </div>
  );
}
