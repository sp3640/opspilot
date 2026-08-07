"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { eventService } from "@/services/event-service";
import type { EventQueryParams } from "@/types/event-api";

const eventKeys = {
  all: ["events"] as const,
  list: (applicationId: string, params: EventQueryParams) =>
    ["events", "list", applicationId, params] as const,
};

export function useEventsByApplication(
  applicationId: string | null,
  params: EventQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: eventKeys.list(applicationId ?? "", params),
    queryFn: () => eventService.listEventsByApplication(applicationId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
  });
}
