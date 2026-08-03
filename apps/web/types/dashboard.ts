export interface DashboardOverview {
  greeting: string;
  environment: string;
  lastDeployment: string;
  uptime: string; // e.g., "99.9%"
  totalProjects: number;
  totalIncidents: number;
  openIncidents: number;
  criticalIncidents: number;
  resolvedIncidents: number;
}

export interface IncidentMetrics {
  total: number;
  open: number;
  critical: number;
  resolved: number;
}

export interface DashboardMetrics {
  projects: number;
  incidents: IncidentMetrics;
  auditLogs: number;
  users: number;
  projectsDeploying?: number;
  incidentsRequiringAttention?: number;
  newUsersThisWeek?: number;
}

export interface DashboardActivity {
  id: number;
  title: string;
  description: string;
  type: string; // e.g., "project", "incident", "user"
  status: "success" | "warning" | "info" | "error";
  timestamp: string;
}

export interface DashboardAlert {
  id: number;
  title: string;
  severity: string;
  status: string;
  createdAt: string;
}

export interface DashboardServiceHealth {
  name: string;
  status: "healthy" | "degraded" | "unhealthy";
  responseTime: number;
  health: number;
}

// Backward-compatible aliases used by existing dashboard components.
export type DashboardSummary = DashboardOverview;
export type DashboardStats = DashboardMetrics;
export type ServiceHealth = DashboardServiceHealth;

export interface DashboardData {
  summary: DashboardOverview;
  stats: DashboardMetrics;
  activity: DashboardActivity[];
  services: DashboardServiceHealth[];
}
