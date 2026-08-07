import { api } from "@/lib/api";
import type { EventListResponse, EventQueryParams } from "@/types/event-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const eventService = {
  async listEventsByApplication(
    applicationId: string,
    params: EventQueryParams
  ): Promise<EventListResponse> {
    const response = await api.get<APIResponse<EventListResponse>>(
      `/applications/${applicationId}/events`,
      { params }
    );
    return response.data.data;
  },
};
