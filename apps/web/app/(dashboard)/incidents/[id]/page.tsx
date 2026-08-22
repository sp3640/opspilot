import { notFound } from "next/navigation";

import { IncidentConsole } from "@/components/incidents/incident-console";

export default async function IncidentConsolePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const incidentID = Number(id);

  if (!Number.isFinite(incidentID) || incidentID <= 0) {
    notFound();
  }

  return <IncidentConsole incidentID={incidentID} />;
}
