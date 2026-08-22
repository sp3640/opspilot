"use client";

import { Link2, Mail, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { StatusBadge } from "@/components/common";
import { Button } from "@/components/ui/button";

import type { Invitation } from "./types";

type InvitationCardProps = {
  invitation: Invitation;
  onRevoke?: (invitation: Invitation) => void;
};

/** Builds the shareable acceptance link and copies it to the clipboard. */
export async function copyInvitationLink(token: string): Promise<void> {
  const link = `${window.location.origin}/accept-invitation?token=${encodeURIComponent(token)}`;
  try {
    await navigator.clipboard.writeText(link);
    toast.success("Invitation link copied to clipboard.");
  } catch {
    toast.error("Unable to copy invitation link. Please try again.");
  }
}

/** Scannable invitation summary using fields returned by the invitations API. */
export function InvitationCard({ invitation, onRevoke }: InvitationCardProps) {
  const canRevoke = invitation.status === "Pending";
  const canCopyLink = invitation.status === "Pending" && Boolean(invitation.token);

  return (
    <article
      className="rounded-3xl border p-5 shadow-[var(--shadow-sm)] transition-all duration-300 hover:shadow-[var(--shadow-md)]"
      style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <div
            className="rounded-2xl p-3"
            style={{
              color: "var(--primary)",
              backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)",
            }}
          >
            <Mail aria-hidden="true" className="h-5 w-5" />
          </div>
          <div className="min-w-0">
            <h2 className="truncate font-semibold tracking-tight">{invitation.email}</h2>
            <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
              Invited {formatDate(invitation.createdAt)}
            </p>
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-1">
          {canCopyLink ? (
            <Button
              type="button"
              variant="ghost"
              onClick={() => {
                void copyInvitationLink(invitation.token);
              }}
              className="h-9 w-9 rounded-xl p-0"
              aria-label={`Copy invitation link for ${invitation.email}`}
            >
              <Link2 aria-hidden="true" className="h-4 w-4" />
            </Button>
          ) : null}
          {canRevoke && onRevoke ? (
            <Button
              type="button"
              variant="ghost"
              onClick={() => onRevoke(invitation)}
              className="h-9 w-9 rounded-xl p-0 text-[var(--danger)]"
              aria-label={`Revoke invitation for ${invitation.email}`}
            >
              <Trash2 aria-hidden="true" className="h-4 w-4" />
            </Button>
          ) : null}
        </div>
      </div>

      <div className="mt-5 flex flex-wrap gap-2">
        <StatusBadge variant={getStatusVariant(invitation.status)}>{invitation.status}</StatusBadge>
        <span
          className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
          style={{
            color: "var(--muted-foreground)",
            backgroundColor: "color-mix(in srgb, var(--muted-foreground) 12%, transparent)",
          }}
        >
          {invitation.role}
        </span>
      </div>

      <dl className="mt-5 grid grid-cols-2 gap-3 border-y py-4 text-sm" style={{ borderColor: "var(--border)" }}>
        <div>
          <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Expires</dt>
          <dd className="mt-1 font-semibold">{formatDate(invitation.expiresAt)}</dd>
        </div>
        <div>
          <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Updated</dt>
          <dd className="mt-1 font-semibold">{formatDate(invitation.updatedAt)}</dd>
        </div>
      </dl>
    </article>
  );
}

export function getStatusVariant(status: string): "success" | "warning" | "critical" | "archived" | "info" {
  switch (status) {
    case "Accepted":
      return "success";
    case "Pending":
      return "warning";
    case "Revoked":
      return "critical";
    case "Expired":
      return "archived";
    default:
      return "info";
  }
}

export function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
