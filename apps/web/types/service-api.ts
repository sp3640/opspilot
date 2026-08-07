import type { OwnerReferenceResponse } from "@/types/pod-api";

export type ServicePortResponse = {
  name?: string;
  port: number;
  targetPort: string;
  protocol: string;
  nodePort?: number;
};

export type ServiceResponse = {
  name: string;
  namespace: string;
  type: string;
  clusterIP: string;
  externalIPs: string[];
  loadBalancerIP?: string;
  ports: ServicePortResponse[];
  targetPorts: string[];
  protocol: string;
  selector: Record<string, string>;
  labels: Record<string, string>;
  annotations: Record<string, string>;
  age: string;
  creationTimestamp: string;
  sessionAffinity?: string;
  internalTrafficPolicy?: string;
  externalTrafficPolicy?: string;
  healthCheckNodePort?: number;
  publishNotReadyAddresses?: boolean;
  ipFamilies?: string[];
  ipFamilyPolicy?: string;
  ownerReferences?: OwnerReferenceResponse[];
  resourceVersion?: string;
  uid?: string;
};

export type ServiceListResponse = {
  items: ServiceResponse[];
  total: number;
};

export type ServiceQueryParams = {
  namespace?: string;
};
