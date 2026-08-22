import { api } from "@/lib/api";
import type {
  AuditLogListResponse,
  AuditQueryParams,
  AuditScope,
} from "@/types/audit-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const auditService = {
  async listAuditLogs(
    scope: AuditScope,
    scopeID: string | number,
    params: AuditQueryParams
  ): Promise<AuditLogListResponse> {
    const endpoint =
      scope === "project"
        ? `/projects/${scopeID}/audit-logs`
        : scope === "incident"
          ? `/incidents/${scopeID}/audit-logs`
          : scope === "alert"
            ? `/alerts/${scopeID}/audit-logs`
            : `/deployments/${scopeID}/audit-logs`;

    const response = await api.get<APIResponse<AuditLogListResponse>>(endpoint, {
      params,
    });

    return response.data.data;
  },
};
