export type NodeResponse = {
  name: string;
  ready: boolean;
  unschedulable: boolean;
  kubeletVersion: string;
  operatingSystem: string;
  architecture: string;
  internalIP?: string;
  labels: Record<string, string>;
  annotations: Record<string, string>;
  age: string;
  creationTimestamp: string;
};

export type NodeListResponse = {
  items: NodeResponse[];
  total: number;
};
