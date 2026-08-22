import type { OwnerReferenceResponse } from "@/types/pod-api";

export type IngressPathResponse = {
  path: string;
  pathType?: string;
  backendService?: string;
  servicePort?: string;
};

export type IngressRuleResponse = {
  host: string;
  paths: IngressPathResponse[];
};

export type IngressTLSResponse = {
  hosts: string[];
  secretName?: string;
};

export type IngressResponse = {
  name: string;
  namespace: string;
  ingressClass?: string;
  hosts: string[];
  paths: string[];
  backendServices: string[];
  tlsEnabled: boolean;
  tlsSecret?: string;
  loadBalancerIP?: string;
  loadBalancerHostname?: string;
  labels: Record<string, string>;
  annotations: Record<string, string>;
  age: string;
  creationTimestamp: string;
  rules?: IngressRuleResponse[];
  httpPaths?: IngressPathResponse[];
  defaultBackend?: string;
  defaultBackendPort?: string;
  tlsConfig?: IngressTLSResponse[];
  status?: string;
  ownerReferences?: OwnerReferenceResponse[];
  uid?: string;
  resourceVersion?: string;
};

export type IngressListResponse = {
  items: IngressResponse[];
  total: number;
};

export type IngressQueryParams = {
  namespace?: string;
};
