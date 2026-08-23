"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { notificationService } from "@/services/notification-service";
import type { CreateNotificationChannelRequest, UpdateNotificationChannelRequest } from "@/types/notification-api";

const notificationKeys = {
  all: ["notification-channels"] as const,
  list: () => ["notification-channels", "list"] as const,
  detail: (id: string) => ["notification-channels", "detail", id] as const,
};

export function useNotificationChannels(queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: notificationKeys.list(),
    queryFn: () => notificationService.listChannels(),
    enabled: queryEnabled && Boolean(accessToken),
  });
}

export function useCreateNotificationChannel() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateNotificationChannelRequest) => notificationService.createChannel(payload),
    onSuccess: () => {
      toast.success("Notification channel created successfully.");
      return queryClient.invalidateQueries({ queryKey: notificationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to create notification channel."));
    },
  });
}

export function useUpdateNotificationChannel() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateNotificationChannelRequest }) =>
      notificationService.updateChannel(id, payload),
    onSuccess: (channel) => {
      toast.success("Notification channel updated successfully.");
      queryClient.setQueryData(notificationKeys.detail(channel.id), channel);
      return queryClient.invalidateQueries({ queryKey: notificationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to update notification channel."));
    },
  });
}

export function useDeleteNotificationChannel() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => notificationService.deleteChannel(id),
    onSuccess: () => {
      toast.success("Notification channel deleted successfully.");
      return queryClient.invalidateQueries({ queryKey: notificationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to delete notification channel."));
    },
  });
}

export function useTestNotificationChannel() {
  return useMutation({
    mutationFn: (id: string) => notificationService.testChannel(id),
    onSuccess: () => {
      toast.success("Test notification sent successfully.");
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to send test notification."));
    },
  });
}

function getMutationErrorMessage(error: unknown, fallback: string) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? fallback;
  }

  if (error instanceof Error && error.message) {
    return error.message;
  }

  return fallback;
}
