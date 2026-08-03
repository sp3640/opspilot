import type { Cluster } from "./types";
import { ClusterCard } from "./cluster-card";

export function ClusterGrid({
  clusters,
  onOpen,
}: {
  clusters: Cluster[];
  onOpen: (clusterID: string) => void;
}) {
  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      {clusters.map((cluster) => (
        <ClusterCard key={cluster.id} cluster={cluster} onOpen={onOpen} />
      ))}
    </div>
  );
}
