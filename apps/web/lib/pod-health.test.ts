import { describe, expect, it } from "vitest";

import { classifyPodHealth } from "./pod-health";
import type { ContainerStatusResponse, PodResponse } from "@/types/pod-api";

function container(overrides: Partial<ContainerStatusResponse> = {}): ContainerStatusResponse {
  return {
    name: "app",
    image: "ghcr.io/opspilot/app:v1",
    ready: true,
    started: true,
    restartCount: 0,
    state: "Running",
    hasReadinessProbe: false,
    hasLivenessProbe: false,
    ...overrides,
  };
}

function pod(overrides: Partial<PodResponse> = {}): Pick<PodResponse, "phase" | "reason" | "message" | "ready" | "containerStatuses"> {
  return {
    phase: "Running",
    ready: true,
    containerStatuses: [container()],
    ...overrides,
  };
}

describe("classifyPodHealth", () => {
  it("classifies a healthy running pod as success", () => {
    const result = classifyPodHealth(pod());
    expect(result).toEqual({ state: "Running", variant: "success" });
  });

  it("classifies a running pod with not-ready containers as a warning", () => {
    const result = classifyPodHealth(pod({ ready: false, containerStatuses: [container({ ready: false })] }));
    expect(result.state).toBe("Running");
    expect(result.variant).toBe("warning");
    expect(result.reason).toBeTruthy();
  });

  it("classifies a pending pod as a warning", () => {
    const result = classifyPodHealth(pod({ phase: "Pending", containerStatuses: [] }));
    expect(result.state).toBe("Pending");
    expect(result.variant).toBe("warning");
  });

  it("classifies an evicted pod distinctly from a generic failure", () => {
    const result = classifyPodHealth(pod({ phase: "Failed", reason: "Evicted", message: "Pod ephemeral local storage usage exceeds the total limit" }));
    expect(result.state).toBe("Evicted");
    expect(result.variant).toBe("critical");
    expect(result.reason).toContain("ephemeral");
  });

  it("classifies a generic failed pod as critical", () => {
    const result = classifyPodHealth(pod({ phase: "Failed", reason: "Error" }));
    expect(result.state).toBe("Failed");
    expect(result.variant).toBe("critical");
  });

  it("detects CrashLoopBackOff from container state even while phase is Running", () => {
    const result = classifyPodHealth(
      pod({
        phase: "Running",
        containerStatuses: [container({ ready: false, state: "CrashLoopBackOff", stateMessage: "back-off 5m0s restarting failed container" })],
      })
    );
    expect(result.state).toBe("CrashLoopBackOff");
    expect(result.variant).toBe("critical");
    expect(result.reason).toBe("back-off 5m0s restarting failed container");
  });

  it("detects ImagePullBackOff via stateReason even when the aggregate state string differs", () => {
    const result = classifyPodHealth(
      pod({
        containerStatuses: [container({ ready: false, state: "Waiting", stateReason: "ImagePullBackOff" })],
      })
    );
    expect(result.state).toBe("ImagePullBackOff");
    expect(result.variant).toBe("critical");
  });

  it("detects ErrImagePull as an image pull failure", () => {
    const result = classifyPodHealth(
      pod({ containerStatuses: [container({ ready: false, state: "ErrImagePull" })] })
    );
    expect(result.state).toBe("ImagePullBackOff");
  });

  it("detects OOMKilled from the last termination reason even when currently ready", () => {
    const result = classifyPodHealth(
      pod({
        containerStatuses: [container({ ready: true, restartCount: 2, lastTerminationReason: "OOMKilled" })],
      })
    );
    expect(result.state).toBe("OOMKilled");
    expect(result.variant).toBe("critical");
  });

  it("classifies a succeeded pod as success", () => {
    const result = classifyPodHealth(pod({ phase: "Succeeded", containerStatuses: [] }));
    expect(result.state).toBe("Succeeded");
    expect(result.variant).toBe("success");
  });

  it("falls back to Unknown for an unrecognized phase", () => {
    const result = classifyPodHealth(pod({ phase: "", containerStatuses: [] }));
    expect(result.state).toBe("Unknown");
    expect(result.variant).toBe("info");
  });

  it("prioritizes crash-loop detection over a generically failed phase", () => {
    const result = classifyPodHealth(
      pod({
        phase: "Failed",
        containerStatuses: [container({ ready: false, state: "CrashLoopBackOff" })],
      })
    );
    expect(result.state).toBe("CrashLoopBackOff");
  });
});
