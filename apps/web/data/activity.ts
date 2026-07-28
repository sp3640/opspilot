import type { Activity } from "@/types/activity";

export const activities: Activity[] = [
  {
    id: 1,
    title: "Project Alpha deployed",
    description: "Deployment completed successfully.",
    time: "2 min ago",
    status: "success",
  },
  {
    id: 2,
    title: "Incident #245 created",
    description: "High CPU usage detected.",
    time: "10 min ago",
    status: "warning",
  },
  {
    id: 3,
    title: "Audit logs exported",
    description: "CSV exported by Siddharth.",
    time: "30 min ago",
    status: "info",
  },
  {
    id: 4,
    title: "Database restarted",
    description: "Maintenance completed.",
    time: "1 hour ago",
    status: "error",
  },
];