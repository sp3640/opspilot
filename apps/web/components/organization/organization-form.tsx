"use client";

import { useEffect } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { organizationFormSchema, type OrganizationFormValues } from "@/lib/validation/organization";
import type { OrganizationResponse } from "@/types/organization-api";

type OrganizationFormProps = {
  organization: OrganizationResponse | null;
  onSubmit: (values: OrganizationFormValues) => void;
  isSubmitting: boolean;
};

export function OrganizationForm({ organization, onSubmit, isSubmitting }: OrganizationFormProps) {
  const form = useForm<OrganizationFormValues>({
    resolver: zodResolver(organizationFormSchema),
    defaultValues: {
      name: organization?.name ?? "",
      slug: organization?.slug ?? "",
      description: organization?.description ?? "",
    },
  });

  useEffect(() => {
    form.reset({
      name: organization?.name ?? "",
      slug: organization?.slug ?? "",
      description: organization?.description ?? "",
    });
  }, [form, organization]);

  return (
    <form
      onSubmit={form.handleSubmit(onSubmit)}
      className="space-y-4"
      noValidate
    >
      <Field
        id="organization-name"
        label="Name"
        error={form.formState.errors.name?.message}
      >
        <input
          id="organization-name"
          type="text"
          autoComplete="organization"
          placeholder="Acme Operations"
          {...form.register("name")}
          className="h-11 w-full rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]"
          style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
        />
      </Field>

      <Field
        id="organization-slug"
        label="Slug"
        helper="Lowercase, hyphenated identifier used in URLs."
        error={form.formState.errors.slug?.message}
      >
        <input
          id="organization-slug"
          type="text"
          placeholder="acme-ops"
          {...form.register("slug")}
          className="h-11 w-full rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]"
          style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
        />
      </Field>

      <Field
        id="organization-description"
        label="Description"
        error={form.formState.errors.description?.message}
      >
        <textarea
          id="organization-description"
          rows={4}
          placeholder="Briefly describe what this organization manages."
          {...form.register("description")}
          className="w-full rounded-2xl border bg-transparent px-3 py-2 text-sm outline-none transition-colors focus:border-[var(--primary)]"
          style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
        />
      </Field>

      <div className="flex justify-end pt-2">
        <Button type="submit" loading={isSubmitting}>
          Save changes
        </Button>
      </div>
    </form>
  );
}

type FieldProps = {
  id: string;
  label: string;
  children: React.ReactNode;
  helper?: string;
  error?: string;
};

function Field({ id, label, children, helper, error }: FieldProps) {
  return (
    <div className="space-y-1.5">
      <label htmlFor={id} className="text-sm font-medium">
        {label}
      </label>
      {children}
      {error ? (
        <p className="text-xs text-red-500">{error}</p>
      ) : helper ? (
        <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
          {helper}
        </p>
      ) : null}
    </div>
  );
}
