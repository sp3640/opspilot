"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { teamService } from "@/services/team-service";
import type {
  AddTeamMemberRequest,
  CreateTeamRequest,
  TeamQueryParams,
  UpdateTeamRequest,
} from "@/types/team-api";

const teamKeys = {
  all: ["teams"] as const,
  list: (params: TeamQueryParams) => ["teams", "list", params] as const,
  detail: (id: string) => ["teams", "detail", id] as const,
  members: (teamId: string) => ["teams", "members", teamId] as const,
};

export function useTeams(params: TeamQueryParams, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: teamKeys.list(params),
    queryFn: () => teamService.listTeams(params),
    enabled: queryEnabled && Boolean(accessToken),
  });
}

export function useTeam(id: string | null) {
  return useQuery({
    queryKey: teamKeys.detail(id ?? ""),
    queryFn: () => teamService.getTeam(id ?? ""),
    enabled: Boolean(id),
  });
}

export function useTeamMembers(teamId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: teamKeys.members(teamId ?? ""),
    queryFn: () => teamService.listTeamMembers(teamId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(teamId),
  });
}

export function useCreateTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateTeamRequest) => teamService.createTeam(payload),
    onSuccess: () => {
      toast.success("Team created successfully.");
      return queryClient.invalidateQueries({ queryKey: teamKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to create team."));
    },
  });
}

export function useUpdateTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateTeamRequest }) =>
      teamService.updateTeam(id, payload),
    onSuccess: (team) => {
      toast.success("Team updated successfully.");
      queryClient.setQueryData(teamKeys.detail(team.id), team);
      return queryClient.invalidateQueries({ queryKey: teamKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to update team."));
    },
  });
}

export function useDeleteTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => teamService.deleteTeam(id),
    onSuccess: () => {
      toast.success("Team deleted successfully.");
      return queryClient.invalidateQueries({ queryKey: teamKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to delete team."));
    },
  });
}

export function useAddTeamMember() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ teamId, payload }: { teamId: string; payload: AddTeamMemberRequest }) =>
      teamService.addTeamMember(teamId, payload),
    onSuccess: (_member, variables) => {
      toast.success("Team member added successfully.");
      return queryClient.invalidateQueries({ queryKey: teamKeys.members(variables.teamId) });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to add team member."));
    },
  });
}

export function useRemoveTeamMember() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ teamId, userId }: { teamId: string; userId: number }) =>
      teamService.removeTeamMember(teamId, userId),
    onSuccess: (_result, variables) => {
      toast.success("Team member removed successfully.");
      return queryClient.invalidateQueries({ queryKey: teamKeys.members(variables.teamId) });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to remove team member."));
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
