import { z } from "zod";

export const projectFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(3, "Project name must contain at least 3 characters.")
    .max(100, "Project name must be 100 characters or fewer."),
  description: z
    .string()
    .trim()
    .max(300, "Description must be 300 characters or fewer."),
});
