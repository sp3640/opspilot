export type ReplicaSetResponse = {
  name: string;
  namespace: string;
  desiredReplicas: number;
  currentReplicas: number;
  readyReplicas: number;
  availableReplicas: number;
  status: string;
  age: string;
  createdAt: string;
};

export type ReplicaSetListResponse = {
  items: ReplicaSetResponse[];
  total: number;
};

export type ReplicaSetQueryParams = {
  namespace?: string;
};
