export type SecretResponse = {
  name: string;
  namespace: string;
  type: string;
  dataKeyCount: number;
  labels: Record<string, string>;
  annotations: Record<string, string>;
  age: string;
  creationTimestamp: string;
};

export type SecretListResponse = {
  items: SecretResponse[];
  total: number;
};

export type SecretQueryParams = {
  namespace?: string;
};
