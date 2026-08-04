import { AlertTriangle } from "lucide-react";

import { Button } from "@/components/ui/button";
import type { OrganizationResponse } from "@/types/organization-api";

type OrganizationDeleteDialogProps = {
  organization: OrganizationResponse;
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isDeleting: boolean;
};

export function OrganizationDeleteDialog({
  organization,
  open,
  onClose,
  onConfirm,
  isDeleting,
}: OrganizationDeleteDialogProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="organization-delete-title"
        className="w-full max-w-md rounded-3xl border p-6 shadow-[var(--shadow-lg)]"
        style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
      >
        <div className="flex items-start gap-3">
          <span className="rounded-full bg-red-500/15 p-2 text-red-500">
            <AlertTriangle aria-hidden={true} className="h-5 w-5" />
          </span>
          <div>
            <h2 id="organization-delete-title" className="text-lg font-semibold tracking-tight">
              Delete organization
            </h2>
            <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
              This action permanently deletes {organization.name}. This cannot be undone.
            </p>
          </div>
        </div>

        <div className="mt-6 flex items-center justify-end gap-3">
          <Button type="button" variant="ghost" onClick={onClose} disabled={isDeleting}>
            Cancel
          </Button>
          <Button type="button" variant="danger" onClick={onConfirm} loading={isDeleting}>
            Delete organization
          </Button>
        </div>
      </div>
    </div>
  );
}
