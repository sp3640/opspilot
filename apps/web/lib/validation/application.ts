import { z } from "zod";

const APPLICATION_RUNTIME_VALUES = [
  "NodeJS",
  "Go",
  "Java",
  "Python",
  "DotNet",
  "Static",
  "Docker",
] as const;

const APPLICATION_STATUS_VALUES = ["Draft", "Ready", "Archived"] as const;

export const applicationSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Application name is required.")
    .max(255, "Name must be 255 characters or less."),
  slug: z.string().trim().max(120, "Slug must be 120 characters or less."),
  description: z.string().max(500, "Description must be 500 characters or less."),
  repository_url: z.string().max(500, "Repository URL must be 500 characters or less."),
  default_branch: z.string().max(128, "Default branch must be 128 characters or less."),
  runtime: z.enum(APPLICATION_RUNTIME_VALUES, {
    message: "Please select a valid runtime.",
  }),
  build_command: z.string(),
  start_command: z.string(),
  port: z
    .number({ message: "Port is required." })
    .int("Port must be an integer.")
    .min(1, "Port must be between 1 and 65535.")
    .max(65535, "Port must be between 1 and 65535."),
  environment: z.string().max(255, "Environment must be 255 characters or less."),
  status: z.enum(APPLICATION_STATUS_VALUES, {
    message: "Please select a valid status.",
  }),
});

export type ApplicationFormValues = z.infer<typeof applicationSchema>;
