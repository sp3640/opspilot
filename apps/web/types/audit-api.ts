export type AuditAction = "CREATE" | "UPDATE" | "DELETE" | "LOGIN";

export type AuditResult = "SUCCESS" | "FAILURE";

export type AuditLogResponse = {
  id: number;
  user_id: number;
  project_id?: string | null;
  application_id?: string | null;
  incident_id?: number | null;
  entity_type: string;
  entity_id: string;
  action: AuditAction;
  result: AuditResult;
  field_name: string;
  old_value: string;
  new_value: string;
  before_state?: string;
  after_state?: string;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
};

export type AuditLogListResponse = {
  items: AuditLogResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type AuditScope = "organization" | "project" | "incident" | "alert" | "deployment";

export type AuditQueryParams = {
  page?: number;
  limit?: number;
  search?: string;
  sort?: "created_at" | "entity_type" | "action";
  order?: "asc" | "desc";
  action?: AuditAction;
  entityType?: string;
  userId?: number;
  applicationId?: string;
  result?: AuditResult;
  dateFrom?: string;
  dateTo?: string;
};
