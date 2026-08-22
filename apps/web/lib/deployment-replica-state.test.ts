import { describe, expect, it } from "vitest";

import { matchRuntimeDeployment } from "./deployment-replica-state";
import type { RuntimeDeploymentResponse } from "@/types/runtime-deployment-api";

function runtimeDeployment(overrides: Partial<RuntimeDeploymentResponse> = {}): RuntimeDeploymentResponse {
  return {
    name: "web-12345678",
    namespace: "default",
    replicas: 3,
    readyReplicas: 3,
    updatedReplicas: 3,
    availableReplicas: 3,
    unavailableReplicas: 0,
    strategy: "RollingUpdate",
    labels: {},
    age: "2h",
    status: "Healthy",
    ...overrides,
  };
}

describe("matchRuntimeDeployment", () => {
  it("matches by namespace + the deployment id's first 8 characters as a name suffix", () => {
    const deployment = { id: "12345678-aaaa-bbbb-cccc-000000000000", namespace: "default" };
    const runtime = runtimeDeployment({ name: "web-12345678", namespace: "default" });

    expect(matchRuntimeDeployment(deployment, [runtime])).toBe(runtime);
  });

  it("is case-insensitive on the name suffix", () => {
    const deployment = { id: "ABCDEF12-aaaa-bbbb-cccc-000000000000", namespace: "default" };
    const runtime = runtimeDeployment({ name: "web-abcdef12", namespace: "default" });

    expect(matchRuntimeDeployment(deployment, [runtime])).toBe(runtime);
  });

  it("returns null when the namespace does not match", () => {
    const deployment = { id: "12345678-aaaa-bbbb-cccc-000000000000", namespace: "default" };
    const runtime = runtimeDeployment({ name: "web-12345678", namespace: "other" });

    expect(matchRuntimeDeployment(deployment, [runtime])).toBeNull();
  });

  it("returns null when no runtime deployment's name ends with the id suffix", () => {
    const deployment = { id: "12345678-aaaa-bbbb-cccc-000000000000", namespace: "default" };
    const runtime = runtimeDeployment({ name: "web-99999999", namespace: "default" });

    expect(matchRuntimeDeployment(deployment, [runtime])).toBeNull();
  });

  it("returns null when more than one candidate matches, rather than guessing", () => {
    const deployment = { id: "12345678-aaaa-bbbb-cccc-000000000000", namespace: "default" };
    const runtimeA = runtimeDeployment({ name: "web-12345678", namespace: "default" });
    const runtimeB = runtimeDeployment({ name: "api-12345678", namespace: "default" });

    expect(matchRuntimeDeployment(deployment, [runtimeA, runtimeB])).toBeNull();
  });

  it("returns null for an empty runtime deployment list", () => {
    const deployment = { id: "12345678-aaaa-bbbb-cccc-000000000000", namespace: "default" };
    expect(matchRuntimeDeployment(deployment, [])).toBeNull();
  });
});
