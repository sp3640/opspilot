"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Copy, FileText, RefreshCw } from "lucide-react";
import { toast } from "sonner";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { Button } from "@/components/ui/button";
import { usePodLogs } from "@/hooks/use-pod-logs";
import type { PodResponse } from "@/types/pod-api";

export function PodLogs({ applicationId, pod }: { applicationId: string; pod: PodResponse }) {
  const containers = useMemo(
    () => (pod.containerStatuses ?? []).map((container) => container.name),
    [pod.containerStatuses]
  );

  const [container, setContainer] = useState(() => containers[0] ?? "");
  const [autoScroll, setAutoScroll] = useState(true);

  const params = useMemo(
    () => ({ applicationId, container: container || undefined }),
    [applicationId, container]
  );
  const { data, error, isError, isLoading, isFetching, refetch } = usePodLogs(pod.namespace, pod.name, params);

  const logText = data?.log ?? "";
  const viewerRef = useRef<HTMLPreElement>(null);

  useEffect(() => {
    if (autoScroll && viewerRef.current) {
      viewerRef.current.scrollTop = viewerRef.current.scrollHeight;
    }
  }, [logText, autoScroll]);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(logText);
      toast.success("Logs copied to clipboard.");
    } catch {
      toast.error("Unable to copy logs.");
    }
  };

  const selectClass =
    "h-11 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]";

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-3">
          {containers.length > 1 ? (
            <>
              <label htmlFor="pod-logs-container" className="sr-only">Container</label>
              <select
                id="pod-logs-container"
                value={container}
                onChange={(event) => setContainer(event.target.value)}
                className={selectClass}
                style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
              >
                {containers.map((name) => (
                  <option key={name} value={name}>{name}</option>
                ))}
              </select>
            </>
          ) : null}

          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={autoScroll}
              onChange={(event) => setAutoScroll(event.target.checked)}
              className="h-4 w-4"
            />
            <span>Auto-scroll</span>
          </label>
        </div>

        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              void refetch();
            }}
            loading={isFetching}
            className="px-3"
            aria-label="Refresh logs"
          >
            <RefreshCw aria-hidden="true" className="h-4 w-4" />
            <span className="hidden sm:inline">Refresh</span>
          </Button>
          <Button
            type="button"
            variant="secondary"
            onClick={handleCopy}
            disabled={!logText}
            className="px-3"
            aria-label="Copy logs"
          >
            <Copy aria-hidden="true" className="h-4 w-4" />
            <span className="hidden sm:inline">Copy</span>
          </Button>
        </div>
      </div>

      {isError ? (
        <ErrorState
          description={error instanceof Error ? error.message : "Unable to load pod logs. Please try again."}
          onRetry={() => {
            void refetch();
          }}
        />
      ) : isLoading ? (
        <ListSkeleton rows={6} />
      ) : logText.trim() === "" ? (
        <EmptyState
          icon={FileText}
          title="No logs available"
          description="This container has not produced any log output yet."
        />
      ) : (
        <pre
          ref={viewerRef}
          className="max-h-[26rem] overflow-auto whitespace-pre-wrap break-words rounded-2xl border p-4 font-mono text-xs leading-5"
          style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}
        >
          {logText}
        </pre>
      )}
    </div>
  );
}
