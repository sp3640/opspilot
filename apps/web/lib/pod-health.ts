import type { ContainerStatusResponse, PodResponse } from "@/types/pod-api";

export type PodHealthState =
  | "Running"
  | "Pending"
  | "Succeeded"
  | "CrashLoopBackOff"
  | "ImagePullBackOff"
  | "OOMKilled"
  | "Evicted"
  | "Failed"
  | "Unknown";

export type PodHealthVariant = "success" | "warning" | "critical" | "info";

export type PodHealthClassification = {
  state: PodHealthState;
  variant: PodHealthVariant;
  reason?: string;
};

const IMAGE_PULL_REASONS = new Set(["ImagePullBackOff", "ErrImagePull", "InvalidImageName"]);

type ClassifiablePod = Pick<PodResponse, "phase" | "reason" | "message" | "ready" | "containerStatuses">;

/**
 * Classifies a pod into one of the "important states" OpsPilot surfaces
 * consistently across pod lists and pod details: Running, Pending,
 * CrashLoopBackOff, ImagePullBackOff, OOMKilled, Evicted, Failed, Unknown
 * (plus Succeeded, for completed Job pods).
 *
 * Container-level crash reasons are checked before the pod's own phase,
 * since a pod can stay "Running" while one of its containers is stuck in
 * CrashLoopBackOff/ImagePullBackOff.
 */
export function classifyPodHealth(pod: ClassifiablePod): PodHealthClassification {
  const statuses = pod.containerStatuses ?? [];

  if (pod.phase === "Failed" && pod.reason === "Evicted") {
    return { state: "Evicted", variant: "critical", reason: pod.message || pod.reason };
  }

  const crashLooping = findContainer(statuses, (c) => c.state === "CrashLoopBackOff" || c.stateReason === "CrashLoopBackOff");
  if (crashLooping) {
    return {
      state: "CrashLoopBackOff",
      variant: "critical",
      reason: crashLooping.stateMessage || "Container is repeatedly crashing.",
    };
  }

  const imagePullFailing = findContainer(
    statuses,
    (c) => IMAGE_PULL_REASONS.has(c.state) || (c.stateReason !== undefined && IMAGE_PULL_REASONS.has(c.stateReason))
  );
  if (imagePullFailing) {
    return {
      state: "ImagePullBackOff",
      variant: "critical",
      reason: imagePullFailing.stateMessage || "Unable to pull the container image.",
    };
  }

  const oomKilled = findContainer(statuses, (c) => c.state === "OOMKilled" || c.lastTerminationReason === "OOMKilled");
  if (oomKilled) {
    return {
      state: "OOMKilled",
      variant: "critical",
      reason: "Container was killed for exceeding its memory limit.",
    };
  }

  if (pod.phase === "Failed") {
    return { state: "Failed", variant: "critical", reason: pod.reason || pod.message };
  }

  if (pod.phase === "Pending") {
    return { state: "Pending", variant: "warning", reason: pod.reason || pod.message };
  }

  if (pod.phase === "Succeeded") {
    return { state: "Succeeded", variant: "success" };
  }

  if (pod.phase === "Running") {
    return pod.ready
      ? { state: "Running", variant: "success" }
      : { state: "Running", variant: "warning", reason: "Not all containers are ready." };
  }

  return { state: "Unknown", variant: "info" };
}

function findContainer(
  statuses: ContainerStatusResponse[],
  predicate: (container: ContainerStatusResponse) => boolean
): ContainerStatusResponse | undefined {
  return statuses.find(predicate);
}
