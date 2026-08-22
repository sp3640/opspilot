import {
  Boxes,
  Building2,
  Database,
  LayoutDashboard,
  FolderKanban,
  ShieldAlert,
  BellRing,
  ChartNoAxesCombined,
  ClipboardList,
  Settings,
  Users,
} from "lucide-react";

export const navigationGroups = [
  {
    title: "Overview",
    items: [
      {
        title: "Dashboard",
        href: "/",
        icon: LayoutDashboard,
      },
      {
        title: "Projects",
        href: "/projects",
        icon: FolderKanban,
      },
      {
        title: "Incidents",
        href: "/incidents",
        icon: ShieldAlert,
      },
      {
        title: "Clusters",
        href: "/clusters",
        icon: Boxes,
      },
      {
        title: "Resources",
        href: "/resources",
        icon: Database,
      },
      {
        title: "Alerts",
        href: "/alerts",
        icon: BellRing,
      },
      {
        title: "Metrics",
        href: "/metrics",
        icon: ChartNoAxesCombined,
      },
    ],
  },

  {
    title: "Security",
    items: [
      {
        title: "Audit Logs",
        href: "/audit",
        icon: ClipboardList,
      },
      {
        title: "Organization",
        href: "/organization",
        icon: Building2,
      },
      {
        title: "Teams",
        href: "/teams",
        icon: Users,
      },
      {
        title: "Settings",
        href: "/settings",
        icon: Settings,
      },
    ],
  },
];
