export const DEPLOYMENT_ENVIRONMENT_VALUES = ["Development", "Staging", "Production"] as const;
export const DEPLOYMENT_STRATEGY_VALUES = ["RollingUpdate", "Recreate"] as const;

export type DeploymentEnvironment = (typeof DEPLOYMENT_ENVIRONMENT_VALUES)[number];
export type DeploymentStrategy = (typeof DEPLOYMENT_STRATEGY_VALUES)[number];

export const DEPLOYMENT_DEFAULT_ENVIRONMENT: DeploymentEnvironment = "Development";
export const DEPLOYMENT_DEFAULT_STRATEGY: DeploymentStrategy = "RollingUpdate";

export const DEPLOYMENT_ENVIRONMENT_OPTIONS: ReadonlyArray<{ value: DeploymentEnvironment; label: string }> = [
  { value: "Development", label: "Development" },
  { value: "Staging", label: "Staging" },
  { value: "Production", label: "Production" },
];

export const DEPLOYMENT_STRATEGY_OPTIONS: ReadonlyArray<{ value: DeploymentStrategy; label: string }> = [
  { value: "RollingUpdate", label: "Rolling update" },
  { value: "Recreate", label: "Recreate" },
];
