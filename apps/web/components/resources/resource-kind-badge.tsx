const labels: Record<string, string> = { PersistentVolumeClaim: "PVC", PersistentVolume: "PV" };
export function ResourceKindBadge({ kind }: { kind: string }) { return <span className="inline-flex rounded-full px-2.5 py-1 text-xs font-semibold" style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}>{labels[kind] ?? kind}</span>; }
