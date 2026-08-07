export type ContainerStatusResponse = {
  name: string;
  image: string;
  ready: boolean;
  restartCount: number;
  state: string;
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
  ready: boolean;
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
};

export type PodListResponse = {
  items: PodResponse[];
  total: number;
};

export type PodQueryParams = {
  namespace?: string;
};
