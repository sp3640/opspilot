import type { Team } from "./types";
import { TeamCard } from "./team-card";

export function TeamGrid({
  teams,
  onOpen,
}: {
  teams: Team[];
  onOpen: (teamID: string) => void;
}) {
  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      {teams.map((team) => (
        <TeamCard key={team.id} team={team} onOpen={onOpen} />
      ))}
    </div>
  );
}
