"use client";

import { useMemo } from "react";
import { ChevronRight } from "lucide-react";
import { useMetricAggregation } from "@/hooks/use-metrics";

import { format } from "./metrics-card";
import type { Metric } from "./types";

type MetricsTableProps = {
	metrics: Metric[];
	projectNames: Map<string, string>;
	clusterNames: Map<string, string>;
	timeRange: "24h" | "7d" | "30d";
	onOpen: (id: string) => void;
};

export function MetricsTable({
	metrics,
	projectNames,
	clusterNames,
	timeRange,
	onOpen,
}: MetricsTableProps) {
	return (
		<div className="overflow-x-auto rounded-2xl border" style={{ borderColor: "var(--border)" }}>
			<table className="w-full min-w-[1280px] text-left text-sm">
				<caption className="sr-only">Metrics list</caption>
				<thead
					className="text-xs uppercase tracking-[.1em]"
					style={{ color: "var(--muted-foreground)", backgroundColor: "var(--muted)" }}
				>
					<tr>
						{["Metric name", "Metric type", "Resource", "Cluster", "Project", "Current value", "Minimum", "Maximum", "Average", "Timestamp"].map((column) => (
							<th key={column} className="px-5 py-3 font-semibold">{column}</th>
						))}
						<th />
					</tr>
				</thead>
				<tbody>
					{metrics.map((metric) => (
						<MetricsTableRow
							key={metric.id}
							metric={metric}
							projectNames={projectNames}
							clusterNames={clusterNames}
							timeRange={timeRange}
							onOpen={onOpen}
						/>
					))}
				</tbody>
			</table>
		</div>
	);
}

function MetricsTableRow({
	metric,
	projectNames,
	clusterNames,
	timeRange,
	onOpen,
}: {
	metric: Metric;
	projectNames: Map<string, string>;
	clusterNames: Map<string, string>;
	timeRange: "24h" | "7d" | "30d";
	onOpen: (id: string) => void;
}) {
	const { start, end } = useMemo(() => getRangeBounds(timeRange), [timeRange]);
	const aggregate = useMetricAggregation({
		projectId: metric.projectId,
		metricType: metric.metricType,
		metricName: metric.metricName,
		start,
		end,
		interval: timeRange === "24h" ? "hour" : "day",
	});

	const minValue = aggregate.data ? `${format(aggregate.data.minimum)} ${aggregate.data.unit}` : aggregate.isLoading ? "Loading" : "Unavailable";
	const maxValue = aggregate.data ? `${format(aggregate.data.maximum)} ${aggregate.data.unit}` : aggregate.isLoading ? "Loading" : "Unavailable";
	const avgValue = aggregate.data ? `${format(aggregate.data.average)} ${aggregate.data.unit}` : aggregate.isLoading ? "Loading" : "Unavailable";

	return (
		<tr
			role="button"
			tabIndex={0}
			onClick={() => onOpen(metric.id)}
			onKeyDown={(event) => {
				if (event.key === "Enter" || event.key === " ") {
					event.preventDefault();
					onOpen(metric.id);
				}
			}}
			className="cursor-pointer border-t hover:bg-[var(--muted)]"
			style={{ borderColor: "var(--border)" }}
		>
			<td className="px-5 py-4 font-medium">{metric.metricName}</td>
			<td className="px-5 py-4">{metric.metricType}</td>
			<td className="max-w-44 truncate px-5 py-4">{metric.resourceKind}: {metric.resourceId}</td>
			<td className="max-w-36 truncate px-5 py-4">{clusterNames.get(metric.clusterId) || metric.clusterId}</td>
			<td className="max-w-36 truncate px-5 py-4">{projectNames.get(metric.projectId) || metric.projectId}</td>
			<td className="px-5 py-4">{format(metric.value)} {metric.unit}</td>
			<td className="px-5 py-4">{minValue}</td>
			<td className="px-5 py-4">{maxValue}</td>
			<td className="px-5 py-4">{avgValue}</td>
			<td className="whitespace-nowrap px-5 py-4">{new Date(metric.timestamp).toLocaleString()}</td>
			<td className="px-5 py-4"><ChevronRight className="h-4 w-4" /></td>
		</tr>
	);
}

function getRangeBounds(range: "24h" | "7d" | "30d") {
	const hours = range === "24h" ? 24 : range === "7d" ? 168 : 720;
	const end = new Date().toISOString();
	const start = new Date(Date.now() - hours * 3600_000).toISOString();

	return { start, end };
}
