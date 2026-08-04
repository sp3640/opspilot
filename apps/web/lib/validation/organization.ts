import { z } from "zod";

const slugPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

export const organizationFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(3, "Organization name must contain at least 3 characters.")
    .max(100, "Organization name must be 100 characters or fewer."),
  slug: z
    .string()
    .trim()
    .max(120, "Slug must be 120 characters or fewer.")
    .refine((value) => value === "" || slugPattern.test(value), {
      message: "Slug must use lowercase letters, numbers, and hyphens.",
    }),
  description: z
    .string()
    .trim()
    .max(500, "Description must be 500 characters or fewer."),
});

export type OrganizationFormValues = z.infer<typeof organizationFormSchema>;
