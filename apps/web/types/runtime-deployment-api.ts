export type RuntimeDeploymentResponse = {
  name: string;
  namespace: string;
  replicas: number;
  readyReplicas: number;
  updatedReplicas: number;
  availableReplicas: number;
  unavailableReplicas: number;
  strategy: string;
  labels: Record<string, string>;
  age: string;
  status: string;
};

export type RuntimeDeploymentListResponse = {
  items: RuntimeDeploymentResponse[];
  total: number;
};

export type RuntimeDeploymentQueryParams = {
  namespace?: string;
};
