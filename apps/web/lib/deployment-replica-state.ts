import type { DeploymentResponse } from "@/types/deployment-api";
import type { RuntimeDeploymentResponse } from "@/types/runtime-deployment-api";

/**
 * Correlates a DB deployment record with its live k8s Deployment object so
 * real replica state (ready/available/updated) can be shown alongside it.
 * The backend names every k8s Deployment object `{app-slug}-{first 8 chars
 * of the deployment's own UUID}` (one k8s object per DB deployment record,
 * never reused across revisions) - so matching on namespace + that suffix
 * is precise, not a guess. Returns null rather than a best-effort pick when
 * the match isn't exactly one candidate, since a wrong replica-state
 * association would be worse than showing "not available".
 */
export function matchRuntimeDeployment(
  deployment: Pick<DeploymentResponse, "id" | "namespace">,
  runtimeDeployments: RuntimeDeploymentResponse[]
): RuntimeDeploymentResponse | null {
  const suffix = `-${deployment.id.slice(0, 8).toLowerCase()}`;
  const candidates = runtimeDeployments.filter(
    (runtime) => runtime.namespace === deployment.namespace && runtime.name.toLowerCase().endsWith(suffix)
  );

  return candidates.length === 1 ? (candidates[0] ?? null) : null;
}
