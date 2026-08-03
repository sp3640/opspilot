import {
  Cloud,
  Container,
  Cpu,
  Database,
  Globe,
  Server,
  type LucideIcon,
} from "lucide-react";

import { StatusBadge } from "@/components/common";
import {
  CLUSTER_PROVIDER_LABELS,
  CLUSTER_STATUS_LABELS,
  type ClusterProvider,
  type ClusterStatus,
} from "@/lib/constants/cluster";

const statusVariantByValue: Record<ClusterStatus, "success" | "warning" | "critical" | "info"> = {
  CONNECTED: "success",
  DISCONNECTED: "warning",
  FAILED: "critical",
  PENDING: "info",
};

const providerIconByValue: Record<ClusterProvider, LucideIcon> = {
  KUBERNETES: Server,
  DOCKER: Container,
  AZURE: Cloud,
  VM: Cpu,
  AWS: Globe,
  CUSTOM: Database,
};

export function ClusterStatusBadge({ status }: { status: string }) {
  const statusValue = (status.toUpperCase() as ClusterStatus) || "PENDING";
  const label = CLUSTER_STATUS_LABELS[statusValue] ?? status;
  const variant = statusVariantByValue[statusValue] ?? "info";

  return <StatusBadge variant={variant}>{label}</StatusBadge>;
}

export function ClusterProviderBadge({ provider }: { provider: string }) {
  const providerValue = (provider.toUpperCase() as ClusterProvider) || "CUSTOM";
  const Icon = providerIconByValue[providerValue] ?? Database;
  const label = CLUSTER_PROVIDER_LABELS[providerValue] ?? provider;

  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold"
      style={{
        color: "var(--primary)",
        backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)",
      }}
    >
      <Icon aria-hidden="true" className="h-3.5 w-3.5" />
      {label}
    </span>
  );
}
