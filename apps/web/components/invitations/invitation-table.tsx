"use client";

import { Link2, Mail, Trash2 } from "lucide-react";

import { StatusBadge } from "@/components/common";
import { Button } from "@/components/ui/button";

import { copyInvitationLink, formatDate, getStatusVariant } from "./invitation-card";
import type { Invitation } from "./types";

type InvitationTableProps = {
  invitations: Invitation[];
  onRevoke?: (invitation: Invitation) => void;
};

/** Dense, keyboard-friendly representation of API-backed invitations. */
export function InvitationTable({ invitations, onRevoke }: InvitationTableProps) {
  return (
    <div className="overflow-x-auto rounded-2xl border" style={{ borderColor: "var(--border)" }}>
      <table className="w-full min-w-[900px] text-left text-sm">
        <caption className="sr-only">Invitations list with role, status, and expiry</caption>
        <thead
          className="text-xs uppercase tracking-[0.1em]"
          style={{
            color: "var(--muted-foreground)",
            backgroundColor: "color-mix(in srgb, var(--muted) 55%, transparent)",
          }}
        >
          <tr>
            <th scope="col" className="px-5 py-3 font-semibold">Email</th>
            <th scope="col" className="px-5 py-3 font-semibold">Role</th>
            <th scope="col" className="px-5 py-3 font-semibold">Status</th>
            <th scope="col" className="px-5 py-3 font-semibold">Invited</th>
            <th scope="col" className="px-5 py-3 font-semibold">Expires</th>
            <th scope="col" className="px-5 py-3 font-semibold">Updated</th>
            <th scope="col" className="w-12 px-5 py-3"><span className="sr-only">Actions</span></th>
          </tr>
        </thead>
        <tbody>
          {invitations.map((invitation) => {
            const canRevoke = invitation.status === "Pending";
            const canCopyLink = invitation.status === "Pending" && Boolean(invitation.token);

            return (
              <tr key={invitation.id} className="border-t" style={{ borderColor: "var(--border)" }}>
                <td className="px-5 py-4">
                  <div className="flex items-center gap-3">
                    <span
                      className="rounded-xl p-2"
                      style={{
                        color: "var(--primary)",
                        backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)",
                      }}
                    >
                      <Mail aria-hidden="true" className="h-4 w-4" />
                    </span>
                    <p className="font-medium">{invitation.email}</p>
                  </div>
                </td>
                <td className="px-5 py-4">{invitation.role}</td>
                <td className="px-5 py-4">
                  <StatusBadge variant={getStatusVariant(invitation.status)}>{invitation.status}</StatusBadge>
                </td>
                <td className="px-5 py-4" style={{ color: "var(--muted-foreground)" }}>
                  {formatDate(invitation.createdAt)}
                </td>
                <td className="px-5 py-4" style={{ color: "var(--muted-foreground)" }}>
                  {formatDate(invitation.expiresAt)}
                </td>
                <td className="px-5 py-4" style={{ color: "var(--muted-foreground)" }}>
                  {formatDate(invitation.updatedAt)}
                </td>
                <td className="px-5 py-4">
                  <div className="flex items-center gap-1">
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
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
