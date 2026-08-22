"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Copy, Download, FileText, Pause, Play, RefreshCw, Search } from "lucide-react";
import { toast } from "sonner";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { Button } from "@/components/ui/button";
import { usePodLogs } from "@/hooks/use-pod-logs";
import { filterLogLines, formatLogLine, LOG_TAIL_LINE_OPTIONS, LOG_TIME_RANGE_OPTIONS, splitLogLines } from "@/lib/log-view";
import type { PodResponse } from "@/types/pod-api";

const LIVE_TAIL_INTERVAL_MS = 5000;

const selectClass =
  "h-11 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]";

/**
 * Live Kubernetes Pod Logs: fetched directly from the Kubernetes API for a
 * running container. This is not a centralized log store - history is
 * limited to whatever the kubelet/container runtime still retains, and logs
 * are lost once the pod is deleted. See ApplicationLogs for the explicit
 * "Centralized Historical Logs" distinction.
 */
export function PodLogs({ applicationId, pod }: { applicationId: string; pod: PodResponse }) {
  const containerStatuses = useMemo(() => pod.containerStatuses ?? [], [pod.containerStatuses]);
  const containers = useMemo(() => containerStatuses.map((c) => c.name), [containerStatuses]);

  const [container, setContainer] = useState(() => containers[0] ?? "");
  const [search, setSearch] = useState("");
  const [timeRangeIndex, setTimeRangeIndex] = useState(LOG_TIME_RANGE_OPTIONS.length - 1);
  const [tailLines, setTailLines] = useState(1000);
  const [showTimestamps, setShowTimestamps] = useState(false);
  const [previous, setPrevious] = useState(false);
  const [autoScroll, setAutoScroll] = useState(true);
  const [live, setLive] = useState(false);

  const selectedContainerStatus = containerStatuses.find((c) => c.name === container);
  const canShowPrevious = (selectedContainerStatus?.restartCount ?? 0) > 0;

  useEffect(() => {
    if (!canShowPrevious && previous) setPrevious(false);
  }, [canShowPrevious, previous]);

  const params = useMemo(
    () => ({
      applicationId,
      container: container || undefined,
      tailLines,
      sinceSeconds: LOG_TIME_RANGE_OPTIONS[timeRangeIndex]?.sinceSeconds,
      timestamps: true,
      previous,
    }),
    [applicationId, container, tailLines, timeRangeIndex, previous]
  );

  const { data, error, isError, isLoading, isFetching, refetch } = usePodLogs(pod.namespace, pod.name, params, {
    refetchIntervalMs: live ? LIVE_TAIL_INTERVAL_MS : false,
  });

  const rawLog = data?.log ?? "";
  const lines = useMemo(() => splitLogLines(rawLog), [rawLog]);
  const filteredLines = useMemo(() => filterLogLines(lines, search), [lines, search]);
  const displayText = useMemo(
    () => filteredLines.map((line) => formatLogLine(line, showTimestamps)).join("\n"),
    [filteredLines, showTimestamps]
  );

  const viewerRef = useRef<HTMLPreElement>(null);
  useEffect(() => {
    if (autoScroll && viewerRef.current) {
      viewerRef.current.scrollTop = viewerRef.current.scrollHeight;
    }
  }, [displayText, autoScroll]);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(displayText);
      toast.success("Logs copied to clipboard.");
    } catch {
      toast.error("Unable to copy logs.");
    }
  };

  const handleDownload = () => {
    const blob = new Blob([rawLog], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `${pod.namespace}-${pod.name}-${container || "container"}${previous ? "-previous" : ""}.log`;
    link.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="space-y-4">
      <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
        Live Kubernetes Pod Logs — fetched directly from the cluster for this running pod. Not a centralized log
        store; history is limited to what the container runtime still retains.
      </p>

      <div className="flex flex-wrap items-center gap-3">
        {containers.length > 1 ? (
          <select
            aria-label="Container"
            value={container}
            onChange={(event) => setContainer(event.target.value)}
            className={selectClass}
            style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
          >
            {containers.map((name) => (
              <option key={name} value={name}>{name}</option>
            ))}
          </select>
        ) : null}

        <select
          aria-label="Time range"
          value={timeRangeIndex}
          onChange={(event) => setTimeRangeIndex(Number(event.target.value))}
          className={selectClass}
          style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
        >
          {LOG_TIME_RANGE_OPTIONS.map((option, index) => (
            <option key={option.label} value={index}>{option.label}</option>
          ))}
        </select>

        <select
          aria-label="Lines"
          value={tailLines}
          onChange={(event) => setTailLines(Number(event.target.value))}
          className={selectClass}
          style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
        >
          {LOG_TAIL_LINE_OPTIONS.map((count) => (
            <option key={count} value={count}>Last {count} lines</option>
          ))}
        </select>

        <div className="relative">
          <Search
            aria-hidden="true"
            className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2"
            style={{ color: "var(--muted-foreground)" }}
          />
          <input
            type="search"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search logs"
            aria-label="Search logs"
            className={`${selectClass} w-48 pl-9`}
            style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
          />
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-4">
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={autoScroll} onChange={(event) => setAutoScroll(event.target.checked)} className="h-4 w-4" />
          <span>Auto-scroll</span>
        </label>
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={showTimestamps} onChange={(event) => setShowTimestamps(event.target.checked)} className="h-4 w-4" />
          <span>Show timestamps</span>
        </label>
        {canShowPrevious ? (
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={previous} onChange={(event) => setPrevious(event.target.checked)} className="h-4 w-4" />
            <span>Previous instance</span>
          </label>
        ) : null}

        <div className="ml-auto flex items-center gap-2">
          <Button
            type="button"
            variant={live ? "primary" : "ghost"}
            onClick={() => setLive((value) => !value)}
            className="px-3"
            aria-pressed={live}
            aria-label={live ? "Pause live tail" : "Resume live tail"}
          >
            {live ? <Pause aria-hidden="true" className="h-4 w-4" /> : <Play aria-hidden="true" className="h-4 w-4" />}
            <span className="hidden sm:inline">{live ? "Live (polling)" : "Live tail"}</span>
          </Button>
          <Button
            type="button"
            variant="ghost"
            onClick={() => { void refetch(); }}
            loading={isFetching}
            className="px-3"
            aria-label="Refresh logs"
          >
            <RefreshCw aria-hidden="true" className="h-4 w-4" />
            <span className="hidden sm:inline">Refresh</span>
          </Button>
          <Button type="button" variant="secondary" onClick={handleDownload} disabled={!rawLog} className="px-3" aria-label="Download logs">
            <Download aria-hidden="true" className="h-4 w-4" />
            <span className="hidden sm:inline">Download</span>
          </Button>
          <Button type="button" variant="secondary" onClick={handleCopy} disabled={!displayText} className="px-3" aria-label="Copy logs">
            <Copy aria-hidden="true" className="h-4 w-4" />
            <span className="hidden sm:inline">Copy</span>
          </Button>
        </div>
      </div>

      {isError ? (
        <ErrorState
          description={error instanceof Error ? error.message : "Unable to load pod logs. Please try again."}
          onRetry={() => { void refetch(); }}
        />
      ) : isLoading ? (
        <ListSkeleton rows={6} />
      ) : lines.length === 0 ? (
        <EmptyState
          icon={FileText}
          title="No logs available"
          description="This container has not produced any log output in the selected range yet."
        />
      ) : filteredLines.length === 0 ? (
        <EmptyState icon={Search} title="No matching lines" description="No log lines match your search." />
      ) : (
        <pre
          ref={viewerRef}
          className="max-h-[26rem] overflow-auto whitespace-pre-wrap break-words rounded-2xl border p-4 font-mono text-xs leading-5"
          style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}
        >
          {displayText}
        </pre>
      )}
    </div>
  );
}
