export type ConfigMapBinaryDataMetadataResponse = {
  key: string;
  sizeBytes: number;
};

export type ConfigMapResponse = {
  name: string;
  namespace: string;
  dataKeyCount: number;
  labels: Record<string, string>;
  annotations: Record<string, string>;
  age: string;
  creationTimestamp: string;
  data?: Record<string, string>;
  binaryData?: ConfigMapBinaryDataMetadataResponse[];
  immutable?: boolean;
  uid?: string;
  resourceVersion?: string;
};

export type ConfigMapListResponse = {
  items: ConfigMapResponse[];
  total: number;
};

export type ConfigMapQueryParams = {
  namespace?: string;
};
