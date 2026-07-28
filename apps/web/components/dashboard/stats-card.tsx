import { Card } from "@/components/ui/card";
import type { LucideIcon } from "lucide-react";

type StatsCardProps = {
  title: string;
  value: number;
  change: string;
  icon: LucideIcon;
};

export function StatsCard({
  title,
  value,
  change,
  icon: Icon,
}: StatsCardProps) {
  return (
    <Card className="group cursor-pointer overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-900/60 transition-all duration-300 hover:-translate-y-1 hover:border-blue-500 hover:shadow-2xl">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-zinc-400">
            {title}
          </p>

          <h2 className="mt-3 text-4xl font-bold tracking-tight">
            {value}
          </h2>

          <p className="mt-2 text-sm text-emerald-400">
            {change}
          </p>
        </div>

        <div className="rounded-2xl bg-blue-500/10 p-4 transition-colors group-hover:bg-blue-500/20">
          <Icon className="h-7 w-7 text-blue-400" />
        </div>
      </div>
    </Card>
  );
}