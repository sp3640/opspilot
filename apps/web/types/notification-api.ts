export type NotificationChannelType = "EMAIL" | "SLACK" | "TEAMS" | "WEBHOOK";

export type NotificationEventType =
  | "CRITICAL_ALERT"
  | "SEV1_INCIDENT"
  | "INCIDENT_ASSIGNED"
  | "DEPLOYMENT_FAILED"
  | "DEPLOYMENT_RECOVERED";

export type NotificationChannelResponse = {
  id: string;
  organizationId: string;
  teamId?: string;
  name: string;
  type: NotificationChannelType;
  maskedTarget: string;
  events: NotificationEventType[];
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
};

export type NotificationChannelListResponse = {
  items: NotificationChannelResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type CreateNotificationChannelRequest = {
  name: string;
  type: NotificationChannelType;
  target: string;
  team_id?: string;
  events: NotificationEventType[];
};

export type UpdateNotificationChannelRequest = {
  name: string;
  target?: string;
  events: NotificationEventType[];
  enabled: boolean;
};

export const NOTIFICATION_CHANNEL_TYPES: NotificationChannelType[] = ["EMAIL", "SLACK", "TEAMS", "WEBHOOK"];

export const NOTIFICATION_EVENT_TYPES: NotificationEventType[] = [
  "CRITICAL_ALERT",
  "SEV1_INCIDENT",
  "INCIDENT_ASSIGNED",
  "DEPLOYMENT_FAILED",
  "DEPLOYMENT_RECOVERED",
];

export const NOTIFICATION_EVENT_LABELS: Record<NotificationEventType, string> = {
  CRITICAL_ALERT: "Critical alert",
  SEV1_INCIDENT: "SEV-1 incident",
  INCIDENT_ASSIGNED: "Incident assigned",
  DEPLOYMENT_FAILED: "Deployment failed",
  DEPLOYMENT_RECOVERED: "Deployment recovered",
};
