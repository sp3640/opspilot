export type ApplicationTeamResponse = {
  id: string;
  organizationId: string;
  applicationId: string;
  teamId: string;
  createdAt: string;
};

export type ApplicationTeamListResponse = {
  items: ApplicationTeamResponse[];
  total: number;
};

export type AssignApplicationTeamRequest = {
  teamId: string;
};
