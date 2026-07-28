import {
  LayoutDashboard,
  FolderKanban,
  ShieldAlert,
  ClipboardList,
  Settings,
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
        title: "Settings",
        href: "/settings",
        icon: Settings,
      },
    ],
  },
];