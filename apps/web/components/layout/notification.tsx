"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Bell, TriangleAlert } from "lucide-react";

import { useIncidents } from "@/hooks/use-incidents";
import { useDismissable } from "@/hooks/use-dismissable";
import { StatusBadge } from "@/components/common/status-badge";
import { INCIDENT_STATUS_VARIANTS, type IncidentStatus } from "@/lib/constants/incident";

const LAST_VIEWED_STORAGE_KEY = "opspilot:notifications:lastViewedAt";
const RECENT_INCIDENT_LIMIT = 10;

/**
 * There is no persisted, per-user notification/read-state backend capability
 * in this application - "notifications" in the backend only refers to
 * outbound alert-routing channel configuration (Slack/Email/Teams/Webhook),
 * not a per-user inbox. This panel is therefore built from real, existing
 * data (open incidents, via the same /incidents endpoint the Incidents page
 * uses) rather than any invented or fake notification feed. "Unread" is a
 * client-only, per-browser approximation (a localStorage timestamp of when
 * the panel was last opened, compared against each incident's updatedAt) -
 * not a real server-persisted read state. This limitation is intentional
 * and documented, not hidden.
 */
function getLastViewedAt(): number {
  if (typeof window === "undefined") return 0;
  const stored = window.localStorage.getItem(LAST_VIEWED_STORAGE_KEY);
  return stored ? Number(stored) || 0 : 0;
}

function setLastViewedAt(timestamp: number) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(LAST_VIEWED_STORAGE_KEY, String(timestamp));
}

export function NotificationButton() {
  const router = useRouter();
  const containerRef = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const [lastViewedAt, setLastViewedAtState] = useState(0);

  useEffect(() => {
    setLastViewedAtState(getLastViewedAt());
  }, []);

  const { data, isLoading, isError, refetch } = useIncidents({
    limit: RECENT_INCIDENT_LIMIT,
    sort: "updated_at",
    order: "desc",
  });

  const openIncidents = (data?.items ?? []).filter((incident) => incident.status !== "RESOLVED");
  const unreadCount = openIncidents.filter((incident) => new Date(incident.updatedAt).getTime() > lastViewedAt).length;

  useDismissable(containerRef, open, () => setOpen(false));

  const handleToggle = () => {
    setOpen((value) => {
      const next = !value;
      if (next) {
        const now = Date.now();
        setLastViewedAtState(now);
        setLastViewedAt(now);
      }
      return next;
    });
  };

  const handleSelect = (incidentId: number) => {
    setOpen(false);
    router.push(`/incidents/${incidentId}`);
  };

  return (
    <div className="relative" ref={containerRef}>
      <button
        type="button"
        onClick={handleToggle}
        aria-label={unreadCount > 0 ? `Notifications, ${unreadCount} unread` : "Notifications"}
        aria-haspopup="true"
        aria-expanded={open}
        className="
          relative
          flex
          h-11
          w-11
          items-center
          justify-center
          rounded-2xl
          border
          transition-all
          duration-300
          hover:-translate-y-0.5
          hover:border-[var(--primary-hover)]/40
        "
        style={{
          background: "var(--surface-hover)",
          borderColor: "var(--border)",
          color: "var(--surface-foreground)",
        }}
      >
        <Bell aria-hidden="true" className="h-5 w-5" />
        {unreadCount > 0 && (
          <span
            aria-hidden="true"
            className="absolute right-2.5 top-2.5 flex h-2.5 w-2.5 items-center justify-center rounded-full"
            style={{ backgroundColor: "var(--danger)" }}
          />
        )}
      </button>

      {open && (
        <div
          role="dialog"
          aria-label="Notifications"
          className="absolute right-0 z-40 mt-2 w-80 overflow-hidden rounded-2xl border"
          style={{ backgroundColor: "var(--card)", borderColor: "var(--border)", boxShadow: "var(--shadow-lg)", color: "var(--card-foreground)" }}
        >
          <div className="flex items-center justify-between border-b px-4 py-3" style={{ borderColor: "var(--border)" }}>
            <span className="text-sm font-semibold">Open incidents</span>
            {unreadCount > 0 && (
              <StatusBadge variant="critical" className="text-[10px]">
                {`${unreadCount} new`}
              </StatusBadge>
            )}
          </div>

          <div className="max-h-96 overflow-y-auto">
            {isLoading ? (
              <p className="px-4 py-8 text-center text-sm" style={{ color: "var(--muted-foreground)" }}>
                Loading…
              </p>
            ) : isError ? (
              <div className="px-4 py-8 text-center">
                <p className="text-sm" style={{ color: "var(--danger-text)" }}>
                  Unable to load incidents.
                </p>
                <button
                  type="button"
                  onClick={() => void refetch()}
                  className="mt-2 text-xs font-medium underline"
                  style={{ color: "var(--primary)" }}
                >
                  Try again
                </button>
              </div>
            ) : openIncidents.length === 0 ? (
              <p className="px-4 py-8 text-center text-sm" style={{ color: "var(--muted-foreground)" }}>
                No open incidents right now.
              </p>
            ) : (
              <ul>
                {openIncidents.map((incident) => {
                  const isUnread = new Date(incident.updatedAt).getTime() > lastViewedAt;
                  return (
                    <li key={incident.id}>
                      <button
                        type="button"
                        onClick={() => handleSelect(incident.id)}
                        className="flex w-full items-start gap-3 border-b px-4 py-3 text-left transition-colors last:border-b-0 hover:bg-[var(--muted)]"
                        style={{ borderColor: "var(--border)" }}
                      >
                        <span
                          className="mt-0.5 shrink-0 rounded-lg p-1.5"
                          style={{ color: "var(--danger-text)", backgroundColor: "var(--danger-bg)" }}
                        >
                          <TriangleAlert aria-hidden="true" className="h-3.5 w-3.5" />
                        </span>
                        <span className="min-w-0 flex-1">
                          <span className="flex items-center gap-2">
                            <span className="truncate text-sm font-medium">{incident.title}</span>
                            {isUnread && (
                              <span aria-label="Unread" className="h-1.5 w-1.5 shrink-0 rounded-full" style={{ backgroundColor: "var(--primary)" }} />
                            )}
                          </span>
                          <span className="mt-1 flex items-center gap-2">
                            <StatusBadge variant={INCIDENT_STATUS_VARIANTS[incident.status as IncidentStatus] ?? "info"} className="text-[10px]">
                              {incident.status}
                            </StatusBadge>
                            <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>
                              {incident.severity}
                            </span>
                          </span>
                        </span>
                      </button>
                    </li>
                  );
                })}
              </ul>
            )}
          </div>

          <div className="border-t px-4 py-2.5" style={{ borderColor: "var(--border)" }}>
            <button
              type="button"
              onClick={() => {
                setOpen(false);
                router.push("/incidents");
              }}
              className="text-xs font-medium"
              style={{ color: "var(--primary)" }}
            >
              View all incidents
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
