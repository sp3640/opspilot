"use client";

import { Boxes, CloudCog, Cpu, Database, FolderKanban, Globe2, Layers3, Network, ShieldCheck } from "lucide-react";

import type { ProjectIconName } from "./types";

const projectIcons = {
  boxes: Boxes,
  "cloud-cog": CloudCog,
  cpu: Cpu,
  database: Database,
  "folder-kanban": FolderKanban,
  globe: Globe2,
  layers: Layers3,
  network: Network,
  "shield-check": ShieldCheck,
} as const;

/** Resolves JSON project icon identifiers only after data reaches the client. */
export function getProjectIcon(iconName: ProjectIconName) {
  return projectIcons[iconName];
}
