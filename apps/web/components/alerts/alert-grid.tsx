import { AlertCard } from "./alert-card";
import type { Alert } from "./types";
export function AlertGrid({ alerts, projectNameById, onOpen }: { alerts: Alert[]; projectNameById: Map<string, string>; onOpen: (id: number) => void }) { return <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">{alerts.map((alert) => <AlertCard key={alert.id} alert={alert} projectName={projectNameById.get(alert.projectId)} onOpen={onOpen} />)}</div>; }
