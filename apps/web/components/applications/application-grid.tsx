"use client";

import { Rocket } from "lucide-react";

import { CardSkeleton, EmptyState } from "@/components/common";
import type { ApplicationResponse } from "@/types/application-api";

import { ApplicationCard } from "./application-card";

type ApplicationGridProps = {
  applications: ApplicationResponse[];
  loading?: boolean;
  onCardClick?: (application: ApplicationResponse) => void;
  onEdit?: (application: ApplicationResponse) => void;
  onDelete?: (application: ApplicationResponse) => void;
};

export function ApplicationGrid({ applications, loading, onCardClick, onEdit, onDelete }: ApplicationGridProps) {
  if (loading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {Array.from({ length: 6 }, (_, index) => (
          <CardSkeleton key={index} />
        ))}
      </div>
    );
  }

  if (applications.length === 0) {
    return (
      <EmptyState
        icon={Rocket}
        title="No applications yet"
        description="Applications you create for this project will appear here."
      />
    );
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      {applications.map((application) => (
        <ApplicationCard
          key={application.id}
          application={application}
          onClick={onCardClick ? () => onCardClick(application) : undefined}
          onEdit={onEdit ? () => onEdit(application) : undefined}
          onDelete={onDelete ? () => onDelete(application) : undefined}
        />
      ))}
    </div>
  );
}
