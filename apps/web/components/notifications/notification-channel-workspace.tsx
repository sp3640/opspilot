"use client";

import { useState } from "react";
import { Bell, Trash2 } from "lucide-react";

import { EmptyState, ErrorState, StatusBadge, TableSkeleton } from "@/components/common";
import { SectionCard } from "@/components/dashboard";
import { Button } from "@/components/ui/button";
import { useHasPermission } from "@/store/auth-store";
import {
  useCreateNotificationChannel,
  useDeleteNotificationChannel,
  useNotificationChannels,
  useTestNotificationChannel,
  useUpdateNotificationChannel,
} from "@/hooks/use-notifications";
import {
  NOTIFICATION_CHANNEL_TYPES,
  NOTIFICATION_EVENT_LABELS,
  NOTIFICATION_EVENT_TYPES,
  type NotificationChannelResponse,
  type NotificationChannelType,
  type NotificationEventType,
} from "@/types/notification-api";

const TARGET_PLACEHOLDER: Record<NotificationChannelType, string> = {
  EMAIL: "oncall@example.com",
  SLACK: "https://hooks.slack.com/services/...",
  TEAMS: "https://outlook.office.com/webhook/...",
  WEBHOOK: "https://example.com/opspilot-webhook",
};

/**
 * Organization-level notification channel management (Phase 24): create,
 * edit, enable/disable, delete, and test-send Email/Slack/Teams/Webhook
 * channels, each subscribed to a set of operational events. Read is
 * available to every organization member; create/update/delete/test are
 * gated by the notification:manage permission (DevOps Engineer/Platform
 * Admin), mirroring InvitationWorkspace's admin-only mutation pattern.
 */
export function NotificationChannelWorkspace() {
  const canManage = useHasPermission("notification:manage");
  const { data, error, isError, isLoading, refetch } = useNotificationChannels();
  const items = data?.items ?? [];

  const [isFormOpen, setFormOpen] = useState(false);

  return (
    <SectionCard
      title="Notification Channels"
      description="Route Critical Alert, SEV-1 Incident, Incident Assigned, Deployment Failed, and Deployment Recovered events to Email, Slack, Teams, or a generic Webhook."
      action={
        canManage && !isFormOpen ? (
          <Button type="button" variant="secondary" onClick={() => setFormOpen(true)}>
            Add channel
          </Button>
        ) : undefined
      }
    >
      {isFormOpen && (
        <div className="mb-6">
          <NotificationChannelForm onClose={() => setFormOpen(false)} />
        </div>
      )}

      {isError ? (
        <ErrorState
          description={error instanceof Error ? error.message : "Unable to load notification channels."}
          onRetry={() => {
            void refetch();
          }}
        />
      ) : isLoading ? (
        <TableSkeleton columns={4} rows={3} />
      ) : items.length === 0 ? (
        <EmptyState
          icon={Bell}
          title="No notification channels configured"
          description="Add a channel so critical alerts and incidents reach your team."
        />
      ) : (
        <ul className="space-y-3">
          {items.map((channel) => (
            <NotificationChannelRow key={channel.id} channel={channel} canManage={canManage} />
          ))}
        </ul>
      )}
    </SectionCard>
  );
}

function NotificationChannelRow({
  channel,
  canManage,
}: {
  channel: NotificationChannelResponse;
  canManage: boolean;
}) {
  const updateChannel = useUpdateNotificationChannel();
  const deleteChannel = useDeleteNotificationChannel();
  const testChannel = useTestNotificationChannel();

  return (
    <li
      className="rounded-2xl border p-4"
      style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="flex items-center gap-2">
            <span className="font-semibold">{channel.name}</span>
            <StatusBadge variant="info">{channel.type}</StatusBadge>
            <StatusBadge variant={channel.enabled ? "success" : "archived"}>
              {channel.enabled ? "Enabled" : "Disabled"}
            </StatusBadge>
          </div>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
            {channel.maskedTarget}
          </p>
          <div className="mt-2 flex flex-wrap gap-1.5">
            {channel.events.map((event) => (
              <span
                key={event}
                className="rounded-full px-2 py-0.5 text-xs"
                style={{ backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)", color: "var(--primary)" }}
              >
                {NOTIFICATION_EVENT_LABELS[event] ?? event}
              </span>
            ))}
          </div>
        </div>

        {canManage && (
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="ghost"
              loading={testChannel.isPending}
              onClick={() => void testChannel.mutateAsync(channel.id)}
            >
              Send test
            </Button>
            <Button
              type="button"
              variant="secondary"
              loading={updateChannel.isPending}
              onClick={() =>
                void updateChannel.mutateAsync({
                  id: channel.id,
                  payload: { name: channel.name, events: channel.events, enabled: !channel.enabled },
                })
              }
            >
              {channel.enabled ? "Disable" : "Enable"}
            </Button>
            <Button
              type="button"
              variant="ghost"
              className="text-[var(--danger)]"
              loading={deleteChannel.isPending}
              onClick={() => void deleteChannel.mutateAsync(channel.id)}
              aria-label={`Delete ${channel.name}`}
            >
              <Trash2 aria-hidden="true" className="h-4 w-4" />
            </Button>
          </div>
        )}
      </div>
    </li>
  );
}

function NotificationChannelForm({ onClose }: { onClose: () => void }) {
  const createChannel = useCreateNotificationChannel();

  const [name, setName] = useState("");
  const [type, setType] = useState<NotificationChannelType>("SLACK");
  const [target, setTarget] = useState("");
  const [events, setEvents] = useState<NotificationEventType[]>([]);

  const toggleEvent = (event: NotificationEventType) => {
    setEvents((current) => (current.includes(event) ? current.filter((e) => e !== event) : [...current, event]));
  };

  const handleSubmit = () => {
    if (!name.trim() || !target.trim() || events.length === 0) return;

    void createChannel.mutateAsync(
      { name: name.trim(), type, target: target.trim(), events },
      {
        onSuccess: () => {
          setName("");
          setTarget("");
          setEvents([]);
          onClose();
        },
      }
    );
  };

  return (
    <div className="rounded-2xl border p-4" style={{ borderColor: "var(--border)" }}>
      <div className="grid gap-3 sm:grid-cols-2">
        <label className="flex flex-col gap-1 text-sm">
          <span style={{ color: "var(--muted-foreground)" }}>Name</span>
          <input
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="Ops Slack channel"
            className="h-10 rounded-xl border bg-transparent px-3 text-sm outline-none focus:border-[var(--primary)]"
            style={{ borderColor: "var(--border)" }}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span style={{ color: "var(--muted-foreground)" }}>Type</span>
          <select
            value={type}
            onChange={(event) => setType(event.target.value as NotificationChannelType)}
            className="h-10 rounded-xl border bg-transparent px-3 text-sm outline-none focus:border-[var(--primary)]"
            style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
          >
            {NOTIFICATION_CHANNEL_TYPES.map((channelType) => (
              <option key={channelType} value={channelType}>
                {channelType}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-sm sm:col-span-2">
          <span style={{ color: "var(--muted-foreground)" }}>
            {type === "EMAIL" ? "Email address" : "Webhook URL"}
          </span>
          <input
            value={target}
            onChange={(event) => setTarget(event.target.value)}
            placeholder={TARGET_PLACEHOLDER[type]}
            className="h-10 rounded-xl border bg-transparent px-3 text-sm outline-none focus:border-[var(--primary)]"
            style={{ borderColor: "var(--border)" }}
          />
        </label>
      </div>

      <div className="mt-3">
        <span className="text-sm" style={{ color: "var(--muted-foreground)" }}>
          Events
        </span>
        <div className="mt-2 flex flex-wrap gap-2">
          {NOTIFICATION_EVENT_TYPES.map((event) => {
            const selected = events.includes(event);
            return (
              <button
                key={event}
                type="button"
                onClick={() => toggleEvent(event)}
                className="rounded-full border px-3 py-1 text-xs transition-colors"
                style={{
                  borderColor: selected ? "var(--primary)" : "var(--border)",
                  backgroundColor: selected ? "color-mix(in srgb, var(--primary) 14%, transparent)" : "transparent",
                  color: selected ? "var(--primary)" : "var(--foreground)",
                }}
              >
                {NOTIFICATION_EVENT_LABELS[event]}
              </button>
            );
          })}
        </div>
      </div>

      <div className="mt-4 flex items-center gap-2">
        <Button type="button" loading={createChannel.isPending} onClick={handleSubmit}>
          Create channel
        </Button>
        <Button type="button" variant="ghost" onClick={onClose}>
          Cancel
        </Button>
      </div>
    </div>
  );
}
