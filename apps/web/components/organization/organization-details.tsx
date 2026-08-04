"use client";

import { Building2, Trash2 } from "lucide-react";

import { SectionCard } from "@/components/dashboard";
import { Button } from "@/components/ui/button";
import { type OrganizationFormValues } from "@/lib/validation/organization";
import type { OrganizationResponse } from "@/types/organization-api";

import { OrganizationForm } from "./organization-form";

type OrganizationDetailsProps = {
  organization: OrganizationResponse;
  onSubmit: (values: OrganizationFormValues) => void;
  isSaving: boolean;
  onDelete: () => void;
  canDelete: boolean;
};

export function OrganizationDetails({
  organization,
  onSubmit,
  isSaving,
  onDelete,
  canDelete,
}: OrganizationDetailsProps) {
  return (
    <SectionCard
      title="Organization details"
      description="Update organization name, slug, and description used across your workspace."
      action={
        <Button type="button" variant="danger" onClick={onDelete} disabled={!canDelete}>
          <Trash2 aria-hidden={true} className="h-4 w-4" />
          Delete
        </Button>
      }
    >
      <div className="mb-5 rounded-2xl border p-4" style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}>
        <div className="flex items-center gap-2">
          <Building2 aria-hidden={true} className="h-4 w-4" />
          <p className="text-sm font-medium">Primary workspace organization</p>
        </div>
        <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
          Changes here update organization metadata for all members.
        </p>
      </div>

      <OrganizationForm organization={organization} onSubmit={onSubmit} isSubmitting={isSaving} />
    </SectionCard>
  );
}
