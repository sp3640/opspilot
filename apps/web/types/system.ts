export type BackendHealthResponse = {
  status: string;
  service?: string;
  environment?: string;
  version?: string;
  database?: string;
  uptime_seconds?: number;
  timestamp?: string;
};
