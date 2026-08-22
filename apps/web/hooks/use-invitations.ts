"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { invitationService } from "@/services/invitation-service";
import type {
  AcceptInvitationRequest,
  InvitationQueryParams,
  InviteRequest,
} from "@/types/invitation-api";

const invitationKeys = {
  all: ["invitations"] as const,
  list: (params: InvitationQueryParams) => ["invitations", "list", params] as const,
};

export function useInvitations(params: InvitationQueryParams, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: invitationKeys.list(params),
    queryFn: () => invitationService.listInvitations(params),
    enabled: queryEnabled && Boolean(accessToken),
  });
}

export function useInviteUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: InviteRequest) => invitationService.inviteUser(payload),
    onSuccess: () => {
      toast.success("Invitation sent successfully.");
      return queryClient.invalidateQueries({ queryKey: invitationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to send invitation."));
    },
  });
}

export function useValidateInvitation(token: string, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: ["invitations", "validate", token],
    queryFn: () => invitationService.validateInvitation(token),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(token),
    retry: false,
  });
}

export function useAcceptInvitation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: AcceptInvitationRequest) => invitationService.acceptInvitation(payload),
    onSuccess: () => {
      toast.success("Invitation accepted successfully.");
      return queryClient.invalidateQueries({ queryKey: invitationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to accept invitation."));
    },
  });
}

export function useRevokeInvitation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => invitationService.revokeInvitation(id),
    onSuccess: () => {
      toast.success("Invitation revoked successfully.");
      return queryClient.invalidateQueries({ queryKey: invitationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to revoke invitation."));
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
