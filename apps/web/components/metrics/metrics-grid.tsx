import { MetricsCard } from "./metrics-card";
import type { Metric } from "./types";
export function MetricsGrid({ metrics, projectNames, clusterNames, onOpen }: { metrics: Metric[]; projectNames: Map<string, string>; clusterNames: Map<string, string>; onOpen: (id: string) => void }) { return <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">{metrics.map((m) => <MetricsCard key={m.id} metric={m} projectName={projectNames.get(m.projectId)} clusterName={clusterNames.get(m.clusterId)} onOpen={onOpen} />)}</div>; }
