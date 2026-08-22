export type ProjectTeamResponse = {
  id: string;
  organizationId: string;
  projectId: string;
  teamId: string;
  createdAt: string;
};

export type ProjectTeamListResponse = {
  items: ProjectTeamResponse[];
  total: number;
};

export type AssignTeamRequest = {
  teamId: string;
};
