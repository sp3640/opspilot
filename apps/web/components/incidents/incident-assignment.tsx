"use client";

import { useState } from "react";
import { CheckCircle2, UserPlus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useAcknowledgeIncident, useAssignIncident } from "@/hooks/use-incidents";
import { useTeamMembers } from "@/hooks/use-teams";
import { useHasPermission } from "@/store/auth-store";
import type { IncidentResponse } from "@/types/incident-api";

/**
 * Incident assignment + acknowledgement (Phase 24 SRE capability): who owns
 * driving this incident to resolution, and whether/when it was first
 * acknowledged (the source data for the MTTA SRE metric). Candidate
 * assignees are drawn from the incident's owner team's membership - there is
 * no organization-wide user directory in this system, so assignment can
 * only be offered once an owner team is set (see EditIncidentModal).
 */
export function IncidentAssignment({ incident }: { incident: IncidentResponse }) {
  const canManage = useHasPermission("incident:manage");
  const assignIncident = useAssignIncident();
  const acknowledgeIncident = useAcknowledgeIncident();

  const { data: membersData, isLoading: isMembersLoading } = useTeamMembers(
    incident.ownerTeamId ?? null,
    Boolean(incident.ownerTeamId)
  );
  const members = membersData?.items ?? [];

  const [selectedUserId, setSelectedUserId] = useState("");
  const assignee = members.find((member) => member.userId === incident.assigneeId);

  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div className="space-y-1 text-sm">
        <p>
          <span style={{ color: "var(--muted-foreground)" }}>Assignee: </span>
          <span className="font-semibold">
            {assignee ? `${assignee.name} (${assignee.email})` : incident.assigneeId ? `User #${incident.assigneeId}` : "Unassigned"}
          </span>
        </p>
        <p>
          <span style={{ color: "var(--muted-foreground)" }}>Acknowledged: </span>
          <span className="font-semibold">
            {incident.acknowledgedAt ? new Date(incident.acknowledgedAt).toLocaleString() : "Not yet acknowledged"}
          </span>
        </p>
      </div>

      {canManage && (
        <div className="flex flex-wrap items-center gap-2">
          {!incident.ownerTeamId ? (
            <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
              Set an owning team to assign this incident to one of its members.
            </p>
          ) : (
            <>
              <select
                value={selectedUserId}
                onChange={(event) => setSelectedUserId(event.target.value)}
                disabled={isMembersLoading || members.length === 0}
                className="h-10 rounded-xl border bg-transparent px-3 text-sm outline-none focus:border-[var(--primary)] disabled:opacity-60"
                style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
              >
                <option value="">
                  {members.length === 0 ? "No team members" : "Select a member"}
                </option>
                {members.map((member) => (
                  <option key={member.userId} value={member.userId}>
                    {member.name} ({member.email})
                  </option>
                ))}
              </select>
              <Button
                type="button"
                variant="secondary"
                loading={assignIncident.isPending}
                disabled={!selectedUserId}
                onClick={() =>
                  void assignIncident.mutateAsync({
                    id: incident.id,
                    payload: { assignee_user_id: Number(selectedUserId) },
                  })
                }
              >
                <UserPlus aria-hidden="true" className="h-4 w-4" />
                Assign
              </Button>
            </>
          )}

          {!incident.acknowledgedAt && (
            <Button
              type="button"
              variant="secondary"
              loading={acknowledgeIncident.isPending}
              onClick={() => void acknowledgeIncident.mutateAsync(incident.id)}
            >
              <CheckCircle2 aria-hidden="true" className="h-4 w-4" />
              Acknowledge
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
