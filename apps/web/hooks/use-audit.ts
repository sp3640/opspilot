"use client";

import { useQuery } from "@tanstack/react-query";
import { auditService } from "@/services/audit-service";
import type {
  AuditLogListResponse,
  AuditQueryParams,
  AuditScope,
} from "@/types/audit-api";

const auditKeys = {
  all: ["audit"] as const,
  list: (scope: AuditScope, scopeID: string | number | null, params: AuditQueryParams) =>
    ["audit", "list", scope, scopeID, params] as const,
};

export function useAuditLogs(
  scope: AuditScope,
  scopeID: string | number | null,
  params: AuditQueryParams
) {
  return useQuery<AuditLogListResponse>({
    queryKey: auditKeys.list(scope, scopeID, params),
    queryFn: () => auditService.listAuditLogs(scope, scopeID as string | number, params),
    enabled: scopeID !== null,
  });
}
