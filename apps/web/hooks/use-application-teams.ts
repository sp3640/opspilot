"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { applicationTeamService } from "@/services/application-team-service";
import type { AssignApplicationTeamRequest } from "@/types/application-team-api";

const applicationTeamKeys = {
  all: ["application-teams"] as const,
  list: (applicationId: string) => ["application-teams", "list", applicationId] as const,
  teamList: (teamId: string) => ["application-teams", "team-list", teamId] as const,
};

export function useApplicationTeams(applicationId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: applicationTeamKeys.list(applicationId ?? ""),
    queryFn: () => applicationTeamService.listApplicationTeams(applicationId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
  });
}

export function useTeamApplications(teamId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: applicationTeamKeys.teamList(teamId ?? ""),
    queryFn: () => applicationTeamService.listTeamApplications(teamId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(teamId),
  });
}

export function useAssignApplicationTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ applicationId, payload }: { applicationId: string; payload: AssignApplicationTeamRequest }) =>
      applicationTeamService.assignTeam(applicationId, payload),
    onSuccess: (_result, variables) => {
      toast.success("Team assigned to application successfully.");
      void queryClient.invalidateQueries({ queryKey: applicationTeamKeys.teamList(variables.payload.teamId) });
      return queryClient.invalidateQueries({ queryKey: applicationTeamKeys.list(variables.applicationId) });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to assign team to application."));
    },
  });
}

export function useRemoveApplicationTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ applicationId, teamId }: { applicationId: string; teamId: string }) =>
      applicationTeamService.removeTeam(applicationId, teamId),
    onSuccess: (_result, variables) => {
      toast.success("Team removed from application successfully.");
      void queryClient.invalidateQueries({ queryKey: applicationTeamKeys.teamList(variables.teamId) });
      return queryClient.invalidateQueries({ queryKey: applicationTeamKeys.list(variables.applicationId) });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to remove team from application."));
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
