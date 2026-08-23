/**
 * Incident Intelligence / Root Cause Analysis types, mirroring
 * apps/backend/internal/dto/rca.go exactly. Every field here is a
 * deterministic correlation over already-real evidence (see
 * apps/backend/internal/rca) - nothing is AI-generated. EvidenceLevel is
 * never "confirmed"; it only ever describes how well the available
 * evidence corroborates itself.
 */

export type RCAEvidenceLevel = "INSUFFICIENT" | "WEAK" | "MODERATE" | "STRONG";

export type RCATimelineEvent = {
  timestamp: string;
  type: "incident" | "alert" | "deployment" | "config_change" | "kubernetes_event";
  title: string;
  description?: string;
  source?: string;
};

export type RCAAlert = {
  id: number;
  title: string;
  severity: string;
  status: string;
  firstSeenAt: string;
  lastSeenAt: string;
  occurrenceCount: number;
};

export type RCAMetricSignal = {
  metricType: string;
  metricName: string;
  unit?: string;
  before: number;
  after: number;
  changePercent: number;
  direction: "increased" | "decreased" | "stable";
  sampleCount: number;
};

export type RCALogSignal = {
  podName: string;
  container?: string;
  line: string;
  keyword: string;
};

export type RCAKubernetesEvent = {
  reason: string;
  message: string;
  type: string;
  count: number;
  lastTimestamp?: string;
  involvedObject: string;
};

export type RCAPod = {
  name: string;
  ready: boolean;
  restartCount: number;
  phase?: string;
  reason?: string;
};

export type RCAConfigChange = {
  entityType: string;
  entityId: string;
  action: string;
  fieldName?: string;
  oldValue?: string;
  newValue?: string;
  changedAt: string;
};

export type RCAPossibleCause = {
  title: string;
  category: string;
  confidence: RCAEvidenceLevel;
  evidence: string[];
};

export type IncidentRCAResponse = {
  incidentId: number;
  summary: string;

  affectedApplicationId?: string;
  affectedApplicationName?: string;
  applicationKnown: boolean;

  windowStart: string;
  windowEnd: string;

  timeline: RCATimelineEvent[];
  recentChanges: RCAConfigChange[];
  correlatedAlerts: RCAAlert[];
  relevantMetrics: RCAMetricSignal[];
  relevantLogs: RCALogSignal[];
  kubernetesEvidence: RCAKubernetesEvent[];
  podEvidence: RCAPod[];

  possibleCauses: RCAPossibleCause[];
  recommendedInvestigationSteps: string[];
  recommendedRemediation: string[];

  evidenceLevel: RCAEvidenceLevel;
  evidenceLevelReason: string;

  generatedAt: string;
};
