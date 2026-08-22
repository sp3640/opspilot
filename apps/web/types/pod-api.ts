export type ContainerStatusResponse = {
  name: string;
  image: string;
  ready: boolean;
  started: boolean;
  restartCount: number;
  state: string;
  stateReason?: string;
  stateMessage?: string;
  exitCode?: number;
  lastTerminationReason?: string;
  lastTerminationExitCode?: number;
  lastTerminationFinishedAt?: string;
  hasReadinessProbe: boolean;
  hasLivenessProbe: boolean;
  cpuRequest?: string;
  cpuLimit?: string;
  memoryRequest?: string;
  memoryLimit?: string;
};

export type ContainerMetricsResponse = {
  name: string;
  cpu: string;
  memory: string;
};

/**
 * Present only when the target cluster has metrics-server installed.
 * Consumers must render "Not available" rather than fabricate a value
 * when this is undefined.
 */
export type PodMetricsResponse = {
  timestamp: string;
  window: string;
  containers: ContainerMetricsResponse[];
};

export type PodConditionResponse = {
  type: string;
  status: string;
  reason?: string;
  message?: string;
  lastTransitionTime?: string;
};

export type OwnerReferenceResponse = {
  apiVersion: string;
  kind: string;
  name: string;
  uid: string;
};

export type PodResponse = {
  name: string;
  namespace: string;
  phase: string;
  reason?: string;
  message?: string;
  ready: boolean;
  readyContainerCount: number;
  restartCount: number;
  nodeName: string;
  podIP: string;
  hostIP: string;
  creationTimestamp: string;
  age: string;
  containerImages: string[];
  containerState: string;
  containerCount: number;
  labels: Record<string, string>;
  ownerReferences: OwnerReferenceResponse[];
  conditions?: PodConditionResponse[];
  events?: string[];
  containerStatuses?: ContainerStatusResponse[];
  startTime?: string;
  qosClass?: string;
  volumes?: string[];
  serviceAccount?: string;
  metrics?: PodMetricsResponse;
};

export type PodListResponse = {
  items: PodResponse[];
  total: number;
};

export type PodQueryParams = {
  namespace?: string;
};
