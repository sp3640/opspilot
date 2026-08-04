import { SectionCard } from "@/components/dashboard";

export function OrganizationSkeleton() {
  return (
    <SectionCard title="Organization details" description="Loading current organization settings.">
      <div className="space-y-3">
        {Array.from({ length: 4 }, (_, index) => (
          <div
            key={index}
            className="h-16 animate-pulse rounded-2xl"
            style={{ backgroundColor: "color-mix(in srgb, var(--muted) 60%, transparent)" }}
          />
        ))}
      </div>
    </SectionCard>
  );
}
