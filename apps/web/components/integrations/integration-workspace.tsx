"use client";

import { useEffect, useState } from "react";
import { GitBranch, Plug, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { EmptyState, ErrorState, StatusBadge, TableSkeleton } from "@/components/common";
import { SectionCard } from "@/components/dashboard";
import { Button } from "@/components/ui/button";
import { useHasPermission } from "@/store/auth-store";
import {
  useCheckIntegration,
  useCreateIntegration,
  useDeleteIntegration,
  useIntegrations,
  useTestIntegration,
} from "@/hooks/use-integrations";
import { useStartGitHubOAuth } from "@/hooks/use-github";
import {
  INTEGRATION_TYPES,
  INTEGRATION_TYPE_LABELS,
  type IntegrationResponse,
  type IntegrationStatus,
  type IntegrationType,
} from "@/types/integration-api";

import { GitHubRepositoryPanel } from "./github-repository-panel";
import { PrometheusQueryPanel } from "./prometheus-query-panel";

/**
 * Organization-level integration management. GitHub (Sprint 28) and
 * Prometheus (Sprint 30) have real connectors - GitHub via the recommended
 * "Connect with GitHub" OAuth flow, Prometheus via an API URL + optional
 * bearer token entered in the generic form below. Every other type (Slack/
 * Email/Loki/OpenTelemetry/Azure/Teams/Webhook) still has no connector
 * implemented (Sprint 27 foundation only) and is honestly labeled "Not yet
 * supported". Read is available to every organization member; create/
 * update/delete/test/check/sync require organization:manage
 * (Platform-Admin-only), reusing the existing organization permission
 * rather than a new one.
 */
export function IntegrationWorkspace() {
  const canManage = useHasPermission("organization:manage");
  const { data, error, isError, isLoading, refetch } = useIntegrations();
  const items = data?.items ?? [];

  const [isFormOpen, setFormOpen] = useState(false);
  const startOAuth = useStartGitHubOAuth();

  useEffect(() => {
    if (typeof window === "undefined") return;
    const params = new URLSearchParams(window.location.search);
    const githubResult = params.get("github");
    if (!githubResult) return;

    if (githubResult === "connected") {
      toast.success("GitHub connected successfully.");
    } else if (githubResult === "error") {
      toast.error("Failed to connect GitHub. Please try again.");
    }
    params.delete("github");
    params.delete("reason");
    const query = params.toString();
    window.history.replaceState({}, "", window.location.pathname + (query ? `?${query}` : ""));
  }, []);

  const handleConnectGitHub = () => {
    void startOAuth.mutateAsync(undefined, {
      onSuccess: (result) => {
        window.location.href = result.authorizeUrl;
      },
    });
  };

  return (
    <SectionCard
      title="Integrations"
      description="Connect external systems (GitHub, Slack, Email, Prometheus, Loki, OpenTelemetry, Azure, Teams, Webhook)."
      action={
        canManage ? (
          <div className="flex items-center gap-2">
            <Button type="button" variant="secondary" loading={startOAuth.isPending} onClick={handleConnectGitHub}>
              <GitBranch aria-hidden="true" className="h-4 w-4" />
              Connect with GitHub
            </Button>
            {!isFormOpen && (
              <Button type="button" variant="ghost" onClick={() => setFormOpen(true)}>
                Add integration
              </Button>
            )}
          </div>
        ) : undefined
      }
    >
      {isFormOpen && (
        <div className="mb-6">
          <IntegrationForm onClose={() => setFormOpen(false)} />
        </div>
      )}

      {isError ? (
        <ErrorState
          description={error instanceof Error ? error.message : "Unable to load integrations."}
          onRetry={() => {
            void refetch();
          }}
        />
      ) : isLoading ? (
        <TableSkeleton columns={4} rows={3} />
      ) : items.length === 0 ? (
        <EmptyState
          icon={Plug}
          title="No integrations configured"
          description="Add an integration to connect an external system to this organization."
        />
      ) : (
        <ul className="space-y-3">
          {items.map((integration) => (
            <IntegrationRow key={integration.id} integration={integration} canManage={canManage} />
          ))}
        </ul>
      )}
    </SectionCard>
  );
}

function IntegrationRow({
  integration,
  canManage,
}: {
  integration: IntegrationResponse;
  canManage: boolean;
}) {
  const deleteIntegration = useDeleteIntegration();
  const testIntegration = useTestIntegration();
  const checkIntegration = useCheckIntegration();

  return (
    <li
      className="rounded-2xl border p-4"
      style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-semibold">{integration.name}</span>
            <StatusBadge variant="info">{INTEGRATION_TYPE_LABELS[integration.type]}</StatusBadge>
            <StatusBadge variant={statusVariant(integration.status)}>{integration.status}</StatusBadge>
            {!integration.connectorImplemented && (
              <StatusBadge variant="archived">Not yet supported</StatusBadge>
            )}
          </div>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
            {integration.connectorImplemented
              ? integration.connectorDescription
              : integration.connectorDescription || "Connector implementation is coming in a future sprint."}
          </p>
          <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
            {integration.hasCredentials ? "Credentials configured" : "No credentials configured"}
            {integration.lastCheckedAt
              ? ` · Last checked ${new Date(integration.lastCheckedAt).toLocaleString()}`
              : " · Never checked"}
          </p>
          {integration.lastError && (
            <p className="mt-1 text-xs" style={{ color: "var(--danger-text)" }}>
              {integration.lastError}
            </p>
          )}
        </div>

        {canManage && (
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="ghost"
              loading={testIntegration.isPending}
              onClick={() => void testIntegration.mutateAsync(integration.id)}
            >
              Test connection
            </Button>
            <Button
              type="button"
              variant="secondary"
              loading={checkIntegration.isPending}
              onClick={() => void checkIntegration.mutateAsync(integration.id)}
            >
              Check status
            </Button>
            <Button
              type="button"
              variant="ghost"
              className="text-[var(--danger-text)]"
              loading={deleteIntegration.isPending}
              onClick={() => void deleteIntegration.mutateAsync(integration.id)}
              aria-label={`Delete ${integration.name}`}
            >
              <Trash2 aria-hidden="true" className="h-4 w-4" />
            </Button>
          </div>
        )}
      </div>

      {integration.type === "github" && <GitHubRepositoryPanel integrationId={integration.id} canManage={canManage} />}
      {integration.type === "prometheus" && (
        <PrometheusQueryPanel integrationId={integration.id} integrationStatus={integration.status} />
      )}
    </li>
  );
}

function statusVariant(status: IntegrationStatus): "success" | "warning" | "critical" | "archived" {
  switch (status) {
    case "CONNECTED":
      return "success";
    case "ERROR":
      return "critical";
    case "DISCONNECTED":
      return "archived";
    default:
      return "warning";
  }
}

function IntegrationForm({ onClose }: { onClose: () => void }) {
  const createIntegration = useCreateIntegration();

  const [type, setType] = useState<IntegrationType>(INTEGRATION_TYPES[0] ?? "webhook");
  const [name, setName] = useState("");
  const [prometheusApiUrl, setPrometheusApiUrl] = useState("");
  const [prometheusBearerToken, setPrometheusBearerToken] = useState("");

  const isPrometheus = type === "prometheus";
  const canSubmit = name.trim().length > 0 && (!isPrometheus || prometheusApiUrl.trim().length > 0);

  const handleSubmit = () => {
    if (!canSubmit) return;

    const metadata = isPrometheus && prometheusApiUrl.trim() ? { api_url: prometheusApiUrl.trim() } : undefined;
    // Credentials are write-only from here on: this form never receives
    // them back (CreateIntegrationRequest's credentials field is the only
    // place a token this session knows about ever exists on the client).
    const credentials =
      isPrometheus && prometheusBearerToken.trim() ? { bearer_token: prometheusBearerToken.trim() } : undefined;

    void createIntegration.mutateAsync(
      { type, name: name.trim(), metadata, credentials },
      {
        onSuccess: () => {
          setName("");
          setPrometheusApiUrl("");
          setPrometheusBearerToken("");
          onClose();
        },
      }
    );
  };

  return (
    <div className="rounded-2xl border p-4" style={{ borderColor: "var(--border)" }}>
      <div className="grid gap-3 sm:grid-cols-2">
        <label className="flex flex-col gap-1 text-sm">
          <span style={{ color: "var(--muted-foreground)" }}>Type</span>
          <select
            value={type}
            onChange={(event) => setType(event.target.value as IntegrationType)}
            className="h-10 rounded-xl border bg-transparent px-3 text-sm outline-none focus:border-[var(--primary)]"
            style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
          >
            {INTEGRATION_TYPES.map((integrationType) => (
              <option key={integrationType} value={integrationType}>
                {INTEGRATION_TYPE_LABELS[integrationType]}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <span style={{ color: "var(--muted-foreground)" }}>Name</span>
          <input
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder={`My ${INTEGRATION_TYPE_LABELS[type]} integration`}
            className="h-10 rounded-xl border bg-transparent px-3 text-sm outline-none focus:border-[var(--primary)]"
            style={{ borderColor: "var(--border)" }}
          />
        </label>
      </div>

      {isPrometheus && (
        <div className="mt-3 grid gap-3 sm:grid-cols-2">
          <label className="flex flex-col gap-1 text-sm">
            <span style={{ color: "var(--muted-foreground)" }}>API URL</span>
            <input
              value={prometheusApiUrl}
              onChange={(event) => setPrometheusApiUrl(event.target.value)}
              placeholder="https://prometheus.example.com"
              autoComplete="off"
              className="h-10 rounded-xl border bg-transparent px-3 text-sm outline-none focus:border-[var(--primary)]"
              style={{ borderColor: "var(--border)" }}
            />
          </label>
          <label className="flex flex-col gap-1 text-sm">
            <span style={{ color: "var(--muted-foreground)" }}>Bearer token (optional)</span>
            <input
              type="password"
              value={prometheusBearerToken}
              onChange={(event) => setPrometheusBearerToken(event.target.value)}
              placeholder="Leave blank if Prometheus requires no auth"
              autoComplete="new-password"
              className="h-10 rounded-xl border bg-transparent px-3 text-sm outline-none focus:border-[var(--primary)]"
              style={{ borderColor: "var(--border)" }}
            />
          </label>
        </div>
      )}

      <div
        className="mt-3 rounded-xl border p-3 text-xs"
        style={{ borderColor: "var(--border)", color: "var(--muted-foreground)" }}
      >
        {type === "github" ? (
          <>
            GitHub has a real connector, but &quot;Connect with GitHub&quot; above (OAuth) is the recommended way to
            connect it - this generic form only creates a connection record with no credentials attached, which
            cannot do anything until credentials are added.
          </>
        ) : isPrometheus ? (
          <>
            Prometheus has a real, read-only connector. The API URL is required; the bearer token is only needed if
            your Prometheus API requires authentication. Once saved, the token is encrypted and never shown again -
            use &quot;Test connection&quot; below to verify Prometheus is reachable.
          </>
        ) : (
          <>
            The {INTEGRATION_TYPE_LABELS[type]} connector is not implemented yet - this creates a connection record
            only, in a Pending state. No credentials are collected here since they would have nothing to
            authenticate against yet; testing this connection will honestly report failure until a real connector
            is added.
          </>
        )}
      </div>

      <div className="mt-4 flex items-center gap-2">
        <Button type="button" loading={createIntegration.isPending} disabled={!canSubmit} onClick={handleSubmit}>
          Create integration
        </Button>
        <Button type="button" variant="ghost" onClick={onClose}>
          Cancel
        </Button>
      </div>
    </div>
  );
}
