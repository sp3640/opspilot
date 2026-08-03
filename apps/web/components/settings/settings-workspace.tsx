"use client";

import { useMemo } from "react";
import { useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import { useQuery } from "@tanstack/react-query";
import { LogOut, Server, UserCircle2 } from "lucide-react";
import { toast } from "sonner";

import { ErrorState, PageHeader, StatusBadge } from "@/components/common";
import { SectionCard } from "@/components/dashboard";
import { Button } from "@/components/ui/button";
import { env } from "@/lib/env";
import { authService } from "@/services/auth-service";
import { systemService } from "@/services/system-service";
import { useAuthStore } from "@/store/auth-store";

export function SettingsWorkspace() {
  const router = useRouter();
  const { resolvedTheme, setTheme } = useTheme();
  const { currentUser, logout, setCurrentUser } = useAuthStore();

  const {
    data: profile,
    isLoading: isProfileLoading,
    isError: isProfileError,
    error: profileError,
    refetch: refetchProfile,
  } = useQuery({
    queryKey: ["auth", "me"],
    queryFn: async () => {
      const response = await authService.me();
      return response.data;
    },
    initialData: currentUser ?? undefined,
  });

  const {
    data: backendHealth,
    isLoading: isHealthLoading,
    isError: isHealthError,
    error: healthError,
    refetch: refetchHealth,
  } = useQuery({
    queryKey: ["system", "health"],
    queryFn: systemService.getHealth,
    retry: 0,
  });

  const activeTheme = resolvedTheme ?? "dark";

  const backendStatus = useMemo(() => {
    if (!backendHealth) return "unknown";
    return backendHealth.status;
  }, [backendHealth]);

  const databaseStatus = backendHealth?.database ?? "unknown";
  const uptime = backendHealth?.uptime_seconds;

  const handleLogout = () => {
    logout();
    setCurrentUser(null);
    toast.success("Signed out successfully.");
    router.replace("/login");
  };

  return (
    <div className="mx-auto max-w-[1600px] space-y-6 lg:space-y-8">
      <PageHeader
        title="Settings"
        description="Manage your profile, sessions, and workspace configuration."
        breadcrumb={[{ label: "Overview", href: "/" }, { label: "Settings" }]}
      />

      <div className="grid gap-6 xl:grid-cols-2">
        <SectionCard title="User profile" description="Your account information from the authenticated session.">
          {isProfileError ? (
            <ErrorState
              title="Unable to load profile"
              description={profileError instanceof Error ? profileError.message : "Failed to fetch user profile."}
              onRetry={() => {
                void refetchProfile();
              }}
            />
          ) : isProfileLoading ? (
            <SettingsLoadingRows rows={3} />
          ) : (
            <div className="space-y-3">
              <InfoRow icon={UserCircle2} label="Name" value={profile?.name ?? "—"} />
              <InfoRow icon={UserCircle2} label="Email" value={profile?.email ?? "—"} />
              <InfoRow icon={UserCircle2} label="User ID" value={profile?.id ? String(profile.id) : "—"} />
            </div>
          )}
        </SectionCard>

        <SectionCard title="Theme settings" description="Choose how the interface is rendered.">
          <div className="space-y-4">
            <label htmlFor="theme-select" className="block text-sm font-medium">
              Active theme
            </label>
            <select
              id="theme-select"
              value={activeTheme}
              onChange={(event) => setTheme(event.target.value)}
              className="h-11 w-full max-w-xs rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]"
              style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
            >
              <option value="system">System</option>
              <option value="dark">Dark</option>
              <option value="light">Light</option>
            </select>
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
              Current resolved theme: {activeTheme}
            </p>
          </div>
        </SectionCard>

        <SectionCard title="Session" description="End your current authenticated session.">
          <div className="flex items-center justify-between gap-3 rounded-2xl border p-4" style={{ borderColor: "var(--border)" }}>
            <div>
              <p className="font-medium">Active session</p>
              <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
                Token-based session stored locally for this browser.
              </p>
            </div>
            <Button type="button" variant="danger" onClick={handleLogout}>
              <LogOut aria-hidden="true" className="h-4 w-4" />
              Logout
            </Button>
          </div>
        </SectionCard>

        <SectionCard title="Workspace" description="Current environment and backend connectivity.">
          <div className="space-y-3">
            <InfoRow icon={Server} label="Application" value={env.APP_NAME} />
            <InfoRow icon={Server} label="API URL" value={env.API_URL} />

            {isHealthError ? (
              <ErrorState
                title="Backend status unavailable"
                description={healthError instanceof Error ? healthError.message : "Health endpoint is not reachable."}
                onRetry={() => {
                  void refetchHealth();
                }}
              />
            ) : isHealthLoading ? (
              <SettingsLoadingRows rows={2} />
            ) : (
              <>
                <div className="flex items-center gap-2">
                  <StatusBadge variant={backendStatus === "ok" ? "success" : "warning"}>
                    {`Backend: ${backendStatus}`}
                  </StatusBadge>
                  <StatusBadge variant={databaseStatus === "ok" ? "success" : "critical"}>
                    {`Database: ${databaseStatus}`}
                  </StatusBadge>
                </div>
                <div className="rounded-2xl border p-4" style={{ borderColor: "var(--border)" }}>
                  <p className="text-sm font-medium">Service metadata</p>
                  <dl className="mt-2 space-y-2 text-sm">
                    <MetadataRow label="Service" value={backendHealth?.service || "—"} />
                    <MetadataRow label="Environment" value={backendHealth?.environment || "—"} />
                    <MetadataRow label="Version" value={backendHealth?.version || "—"} />
                    <MetadataRow label="Uptime" value={typeof uptime === "number" ? `${uptime}s` : "—"} />
                    <MetadataRow label="Timestamp" value={backendHealth?.timestamp || "—"} />
                  </dl>
                </div>
              </>
            )}
          </div>
        </SectionCard>
      </div>
    </div>
  );
}

function InfoRow({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ComponentType<{ className?: string; "aria-hidden"?: boolean }>;
  label: string;
  value: string;
}) {
  return (
    <div className="rounded-2xl border p-4" style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}>
      <div className="flex items-center gap-2">
        <Icon aria-hidden={true} className="h-4 w-4" />
        <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
          {label}
        </p>
      </div>
      <p className="mt-1 break-all font-medium">{value}</p>
    </div>
  );
}

function MetadataRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-3">
      <dt style={{ color: "var(--muted-foreground)" }}>{label}</dt>
      <dd className="text-right font-medium">{value}</dd>
    </div>
  );
}

function SettingsLoadingRows({ rows }: { rows: number }) {
  return (
    <div className="space-y-3">
      {Array.from({ length: rows }, (_, index) => (
        <div
          key={index}
          className="h-16 animate-pulse rounded-2xl"
          style={{ backgroundColor: "color-mix(in srgb, var(--muted) 60%, transparent)" }}
        />
      ))}
    </div>
  );
}
