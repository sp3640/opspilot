/** Default window (minutes) compared on each side of a deployment's pivot time. */
export const CORRELATION_WINDOW_MINUTES = 30;

export type CorrelationVerdict = "potential_degradation" | "no_degradation_detected" | "insufficient_data";

export type TimeWindow = { start: string; end: string };

export type CountWindowSignal = {
  status: "available";
  before: number;
  after: number;
  degraded: boolean;
};

export type PodHealthSignal =
  | { status: "unavailable" }
  | {
      status: "post_only";
      totalPods: number;
      unreadyPods: number;
      restartedSinceDeployment: number;
      degraded: boolean;
    };

export type ClusterMetricSignal =
  | { status: "unavailable" }
  | {
      status: "available";
      scope: "cluster";
      unit: string;
      before: number | null;
      after: number | null;
      changePercent: number | null;
    };

export type UnavailableSignal = { status: "unavailable" };

export type DeploymentHealthCorrelationResult = {
  windowMinutes: number;
  beforeWindow: TimeWindow;
  afterWindow: TimeWindow;
  afterWindowElapsed: boolean;
  alerts: CountWindowSignal;
  incidents: CountWindowSignal;
  pods: PodHealthSignal;
  cpu: ClusterMetricSignal;
  memory: ClusterMetricSignal;
  errorRate: UnavailableSignal;
  latency: UnavailableSignal;
  requestRate: UnavailableSignal;
  verdict: CorrelationVerdict;
  reasons: string[];
};

type DeploymentPivotLike = { startedAt?: string; completedAt?: string; createdAt: string };

/**
 * The before/after windows to fetch data for, computed from the
 * deployment's own real timestamps only: `startedAt` (falling back to
 * `createdAt` if the deployment never started) marks when it began
 * affecting the system, `completedAt` (falling back to the start) marks
 * when it finished. Callers use this to know what `start`/`end` to query
 * before evaluateDeploymentHealthCorrelation is called with the results.
 */
export function computeDeploymentWindows(
  deployment: DeploymentPivotLike,
  windowMinutes: number = CORRELATION_WINDOW_MINUTES
): { beforeWindow: TimeWindow; afterWindow: TimeWindow } {
  const pivotStart = new Date(deployment.startedAt ?? deployment.createdAt).getTime();
  const pivotEnd = new Date(deployment.completedAt ?? deployment.startedAt ?? deployment.createdAt).getTime();
  const windowMs = windowMinutes * 60 * 1000;

  return {
    beforeWindow: { start: new Date(pivotStart - windowMs).toISOString(), end: new Date(pivotStart).toISOString() },
    afterWindow: { start: new Date(pivotEnd).toISOString(), end: new Date(pivotEnd + windowMs).toISOString() },
  };
}

type AlertLike = { firstSeenAt: string };
type IncidentLike = { createdAt: string };
type PodContainerLike = { lastTerminationFinishedAt?: string };
type PodLike = { ready: boolean; containerStatuses?: PodContainerLike[] };
type MetricAggregateLike = { count: number; average: number; unit: string };

function withinWindow(timestamp: string, window: TimeWindow): boolean {
  const time = new Date(timestamp).getTime();
  return time >= new Date(window.start).getTime() && time < new Date(window.end).getTime();
}

function countWindowSignal(timestamps: string[], beforeWindow: TimeWindow, afterWindow: TimeWindow): CountWindowSignal {
  const before = timestamps.filter((timestamp) => withinWindow(timestamp, beforeWindow)).length;
  const after = timestamps.filter((timestamp) => withinWindow(timestamp, afterWindow)).length;
  return { status: "available", before, after, degraded: after > before };
}

function clusterMetricSignal(
  before: MetricAggregateLike | null | undefined,
  after: MetricAggregateLike | null | undefined
): ClusterMetricSignal {
  if (!before && !after) return { status: "unavailable" };

  const beforeValue = before && before.count > 0 ? before.average : null;
  const afterValue = after && after.count > 0 ? after.average : null;
  const unit = before?.unit || after?.unit || "";
  const changePercent =
    beforeValue !== null && afterValue !== null && beforeValue !== 0
      ? ((afterValue - beforeValue) / Math.abs(beforeValue)) * 100
      : null;

  return { status: "available", scope: "cluster", unit, before: beforeValue, after: afterValue, changePercent };
}

/**
 * Evaluates whether a deployment is potentially correlated with a health
 * regression, from already-fetched real data only. Every input is optional
 * and independently omittable - a signal this function has no data for is
 * reported as "unavailable"/"post_only" rather than guessed at. Cluster-wide
 * CPU/memory are deliberately excluded from the verdict itself (they aren't
 * scoped to this application, so they're weak, easily-confounded evidence)
 * and are surfaced only as labeled supplementary context.
 */
export function evaluateDeploymentHealthCorrelation(input: {
  deployment: DeploymentPivotLike;
  windowMinutes?: number;
  alerts?: AlertLike[];
  incidents?: IncidentLike[];
  pods?: PodLike[];
  cpuBefore?: MetricAggregateLike | null;
  cpuAfter?: MetricAggregateLike | null;
  memoryBefore?: MetricAggregateLike | null;
  memoryAfter?: MetricAggregateLike | null;
}): DeploymentHealthCorrelationResult {
  const windowMinutes = input.windowMinutes ?? CORRELATION_WINDOW_MINUTES;
  const { beforeWindow, afterWindow } = computeDeploymentWindows(input.deployment, windowMinutes);
  const afterWindowElapsed = Date.now() >= new Date(afterWindow.end).getTime();

  const alerts = countWindowSignal((input.alerts ?? []).map((alert) => alert.firstSeenAt), beforeWindow, afterWindow);
  const incidents = countWindowSignal((input.incidents ?? []).map((incident) => incident.createdAt), beforeWindow, afterWindow);

  const pivotEndMs = new Date(input.deployment.completedAt ?? input.deployment.startedAt ?? input.deployment.createdAt).getTime();
  const pods: PodHealthSignal = input.pods
    ? (() => {
        const totalPods = input.pods!.length;
        const unreadyPods = input.pods!.filter((pod) => !pod.ready).length;
        const restartedSinceDeployment = input.pods!.filter((pod) =>
          (pod.containerStatuses ?? []).some(
            (container) => container.lastTerminationFinishedAt && new Date(container.lastTerminationFinishedAt).getTime() > pivotEndMs
          )
        ).length;
        return {
          status: "post_only" as const,
          totalPods,
          unreadyPods,
          restartedSinceDeployment,
          degraded: unreadyPods > 0 || restartedSinceDeployment > 0,
        };
      })()
    : { status: "unavailable" };

  const cpu = clusterMetricSignal(input.cpuBefore, input.cpuAfter);
  const memory = clusterMetricSignal(input.memoryBefore, input.memoryAfter);

  const reasons: string[] = [];
  if (alerts.degraded) {
    reasons.push(`${alerts.after} alert(s) fired in the ${windowMinutes} minutes after deployment, vs ${alerts.before} in the ${windowMinutes} minutes before.`);
  }
  if (incidents.degraded) {
    reasons.push(`${incidents.after} incident(s) opened in the ${windowMinutes} minutes after deployment, vs ${incidents.before} in the ${windowMinutes} minutes before.`);
  }
  if (pods.status === "post_only" && pods.degraded) {
    if (pods.restartedSinceDeployment > 0) {
      reasons.push(`${pods.restartedSinceDeployment} of ${pods.totalPods} pod(s) have restarted since this deployment.`);
    }
    if (pods.unreadyPods > 0) {
      reasons.push(`${pods.unreadyPods} of ${pods.totalPods} pod(s) are currently not ready.`);
    }
  }

  let verdict: CorrelationVerdict;
  if (reasons.length > 0) {
    verdict = "potential_degradation";
  } else if (!afterWindowElapsed) {
    verdict = "insufficient_data";
  } else {
    verdict = "no_degradation_detected";
  }

  return {
    windowMinutes,
    beforeWindow,
    afterWindow,
    afterWindowElapsed,
    alerts,
    incidents,
    pods,
    cpu,
    memory,
    errorRate: { status: "unavailable" },
    latency: { status: "unavailable" },
    requestRate: { status: "unavailable" },
    verdict,
    reasons,
  };
}
