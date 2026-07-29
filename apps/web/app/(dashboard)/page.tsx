'use client';

import {
  DashboardHero,
  MetricsGrid,
  QuickActions,
  RecentActivity,
  SystemHealth,
} from "@/components/dashboard";

export default function DashboardPage() {
  return (
    <div className="mx-auto max-w-[1600px] space-y-6 lg:space-y-8">
      <DashboardHero />
      <MetricsGrid />

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.45fr)_minmax(340px,0.85fr)]">
        <RecentActivity />
        <QuickActions />
      </div>

      <SystemHealth />
    </div>
  );
}
