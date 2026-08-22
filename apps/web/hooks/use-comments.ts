"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { commentService } from "@/services/comment-service";
import type { CommentQueryParams, CreateCommentRequest } from "@/types/comment-api";

const commentKeys = {
  all: ["comments"] as const,
  byIncident: (incidentId: number) => ["comments", "list", incidentId] as const,
  list: (incidentId: number, params: CommentQueryParams) => ["comments", "list", incidentId, params] as const,
};

export function useComments(incidentId: number | null, params: CommentQueryParams, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: commentKeys.list(incidentId ?? 0, params),
    queryFn: () => commentService.listComments(incidentId ?? 0, params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(incidentId),
  });
}

export function useCreateComment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ incidentId, payload }: { incidentId: number; payload: CreateCommentRequest }) =>
      commentService.createComment(incidentId, payload),
    onSuccess: (_comment, variables) => {
      toast.success("Comment added successfully.");
      return queryClient.invalidateQueries({ queryKey: commentKeys.byIncident(variables.incidentId) });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to add comment."));
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
