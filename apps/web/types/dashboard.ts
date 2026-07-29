export interface DashboardSummary {
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

export interface DashboardStats {
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

export interface ServiceHealth {
  name: string;
  status: "healthy" | "degraded" | "unhealthy";
  responseTime: number;
  health: number;
}

export interface DashboardData {
  summary: DashboardSummary;
  stats: DashboardStats;
  activity: DashboardActivity[];
  services: ServiceHealth[];
}
