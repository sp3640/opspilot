export type AlertJson = Record<string, unknown>;
export type AlertResponse = { id: number; projectId: string; incidentId?: number; title: string; description: string; severity: string; status: string; source: string; resourceType: string; resourceId: string; fingerprint: string; occurrenceCount: number; labels: AlertJson; metadata: AlertJson; firstSeenAt: string; lastSeenAt: string; acknowledgedAt?: string; resolvedAt?: string; createdBy: number; createdAt: string; updatedAt: string };
export type AlertListResponse = { items: AlertResponse[]; page: number; limit: number; total: number; totalPages: number };
export type AlertPayload = { project_id: string; incident_id?: number; title: string; description: string; severity: string; status: string; source: string; resource_type: string; resource_id: string; fingerprint: string; occurrence_count: number; labels?: AlertJson; metadata?: AlertJson; first_seen_at: string; last_seen_at: string; acknowledged_at?: string; resolved_at?: string };
export type CreateAlertRequest = AlertPayload;
export type UpdateAlertRequest = AlertPayload;
export type AlertQueryParams = { page?: number; limit?: number; search?: string; sort?: "created_at" | "updated_at" | "severity" | "status" | "last_seen_at"; order?: "asc" | "desc"; projectId?: string; severity?: string; status?: string; source?: string };
