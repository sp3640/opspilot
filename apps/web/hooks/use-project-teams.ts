"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { projectTeamService } from "@/services/project-team-service";
import type { AssignTeamRequest } from "@/types/project-team-api";

const projectTeamKeys = {
  all: ["project-teams"] as const,
  list: (projectId: string) => ["project-teams", "list", projectId] as const,
  teamList: (teamId: string) => ["project-teams", "team-list", teamId] as const,
};

export function useProjectTeams(projectId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: projectTeamKeys.list(projectId ?? ""),
    queryFn: () => projectTeamService.listProjectTeams(projectId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(projectId),
  });
}

export function useTeamProjects(teamId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: projectTeamKeys.teamList(teamId ?? ""),
    queryFn: () => projectTeamService.listTeamProjects(teamId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(teamId),
  });
}

export function useAssignProjectTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ projectId, payload }: { projectId: string; payload: AssignTeamRequest }) =>
      projectTeamService.assignTeam(projectId, payload),
    onSuccess: (_result, variables) => {
      toast.success("Team assigned to project successfully.");
      void queryClient.invalidateQueries({ queryKey: projectTeamKeys.teamList(variables.payload.teamId) });
      return queryClient.invalidateQueries({ queryKey: projectTeamKeys.list(variables.projectId) });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to assign team to project."));
    },
  });
}

export function useRemoveProjectTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ projectId, teamId }: { projectId: string; teamId: string }) =>
      projectTeamService.removeTeam(projectId, teamId),
    onSuccess: (_result, variables) => {
      toast.success("Team removed from project successfully.");
      void queryClient.invalidateQueries({ queryKey: projectTeamKeys.teamList(variables.teamId) });
      return queryClient.invalidateQueries({ queryKey: projectTeamKeys.list(variables.projectId) });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to remove team from project."));
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
