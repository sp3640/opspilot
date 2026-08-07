export type PodLogResponse = {
  pod: string;
  container: string;
  namespace: string;
  log: string;
  retrievedAt: string;
};

export type PodLogQueryParams = {
  applicationId: string;
  container?: string;
};
