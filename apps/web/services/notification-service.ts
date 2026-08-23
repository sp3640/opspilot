import { api } from "@/lib/api";
import type {
  CreateNotificationChannelRequest,
  NotificationChannelListResponse,
  NotificationChannelResponse,
  UpdateNotificationChannelRequest,
} from "@/types/notification-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const notificationService = {
  async listChannels(): Promise<NotificationChannelListResponse> {
    const response = await api.get<APIResponse<NotificationChannelListResponse>>("/notification-channels", {
      params: { limit: 100 },
    });
    return response.data.data;
  },

  async getChannel(id: string): Promise<NotificationChannelResponse> {
    const response = await api.get<APIResponse<NotificationChannelResponse>>(`/notification-channels/${id}`);
    return response.data.data;
  },

  async createChannel(payload: CreateNotificationChannelRequest): Promise<NotificationChannelResponse> {
    const response = await api.post<APIResponse<NotificationChannelResponse>>("/notification-channels", payload);
    return response.data.data;
  },

  async updateChannel(id: string, payload: UpdateNotificationChannelRequest): Promise<NotificationChannelResponse> {
    const response = await api.put<APIResponse<NotificationChannelResponse>>(`/notification-channels/${id}`, payload);
    return response.data.data;
  },

  async deleteChannel(id: string): Promise<void> {
    await api.delete<APIResponse<null>>(`/notification-channels/${id}`);
  },

  async testChannel(id: string): Promise<void> {
    await api.post<APIResponse<null>>(`/notification-channels/${id}/test`);
  },
};
