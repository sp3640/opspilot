"use client";

import { useMemo, useState } from "react";

import { ErrorState } from "@/components/common";
import { SectionCard } from "@/components/dashboard";
import { type OrganizationFormValues } from "@/lib/validation/organization";
import {
  useCreateOrganization,
  useDeleteOrganization,
  useOrganizations,
  useUpdateOrganization,
} from "@/hooks/use-organizations";

import { OrganizationCard } from "./organization-card";
import { OrganizationDeleteDialog } from "./organization-delete-dialog";
import { OrganizationDetails } from "./organization-details";
import { OrganizationEmpty } from "./organization-empty";
import { OrganizationForm } from "./organization-form";
import { OrganizationHeader } from "./organization-header";
import { OrganizationSkeleton } from "./organization-skeleton";

const listParams = {
  page: 1,
  limit: 1,
  sort: "updated_at" as const,
  order: "desc" as const,
};

export function OrganizationWorkspace() {
  const [isDeleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [isCreateMode, setCreateMode] = useState(false);

  const {
    data,
    isLoading,
    isError,
    error,
    refetch,
  } = useOrganizations(listParams);

  const updateMutation = useUpdateOrganization();
  const createMutation = useCreateOrganization();
  const deleteMutation = useDeleteOrganization();

  const organization = useMemo(() => {
    return data?.items?.[0] ?? null;
  }, [data]);

  const canDelete = false;

  const handleUpdate = (values: OrganizationFormValues) => {
    if (!organization) return;

    updateMutation.mutate({
      id: organization.id,
      payload: {
        name: values.name,
        slug: values.slug || undefined,
        description: values.description,
      },
    });
  };

  const handleDelete = () => {
    if (!organization || !canDelete) return;

    deleteMutation.mutate(organization.id, {
      onSuccess: () => {
        setDeleteDialogOpen(false);
      },
    });
  };

  const handleCreate = (values: OrganizationFormValues) => {
    createMutation.mutate(
      {
        name: values.name,
        slug: values.slug || undefined,
        description: values.description,
      },
      {
        onSuccess: () => {
          setCreateMode(false);
        },
      }
    );
  };

  return (
    <div className="mx-auto max-w-[1600px] space-y-6 lg:space-y-8">
      <OrganizationHeader />

      {isError ? (
        <ErrorState
          title="Unable to load organization"
          description={error instanceof Error ? error.message : "Failed to fetch organization details."}
          onRetry={() => {
            void refetch();
          }}
        />
      ) : isLoading ? (
        <OrganizationSkeleton />
      ) : !organization ? (
        isCreateMode ? (
          <SectionCard title="Create organization" description="Set up your workspace organization metadata.">
            <OrganizationForm organization={null} onSubmit={handleCreate} isSubmitting={createMutation.isPending} />
          </SectionCard>
        ) : (
          <OrganizationEmpty onCreate={() => setCreateMode(true)} />
        )
      ) : (
        <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
          <OrganizationDetails
            organization={organization}
            onSubmit={handleUpdate}
            isSaving={updateMutation.isPending}
            onDelete={() => setDeleteDialogOpen(true)}
            canDelete={canDelete}
          />
          <OrganizationCard organization={organization} />
        </div>
      )}

      {organization ? (
        <OrganizationDeleteDialog
          organization={organization}
          open={isDeleteDialogOpen}
          onClose={() => setDeleteDialogOpen(false)}
          onConfirm={handleDelete}
          isDeleting={deleteMutation.isPending}
        />
      ) : null}
    </div>
  );
}
