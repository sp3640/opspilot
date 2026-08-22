export type PodLogResponse = {
  pod: string;
  container: string;
  namespace: string;
  log: string;
  previous: boolean;
  retrievedAt: string;
};

export type PodLogQueryParams = {
  applicationId: string;
  container?: string;
  tailLines?: number;
  sinceSeconds?: number;
  timestamps?: boolean;
  previous?: boolean;
};
