export type NamespaceResponse = {
  name: string;
  phase: string;
  labels: Record<string, string>;
  annotations: Record<string, string>;
  age: string;
  creationTimestamp: string;
};

export type NamespaceListResponse = {
  items: NamespaceResponse[];
  total: number;
};
