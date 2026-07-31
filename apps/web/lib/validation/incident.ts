import { z } from "zod";
import {
  INCIDENT_DEFAULT_STATUS,
  INCIDENT_SEVERITY_VALUES,
  INCIDENT_STATUS_VALUES,
} from "@/lib/constants";

const incidentBaseSchema = {
  title: z
    .string()
    .trim()
    .min(3, "Title must contain at least 3 characters.")
    .max(255, "Title must be 255 characters or fewer."),
  description: z
    .string()
    .trim()
    .max(1000, "Description must be 1000 characters or fewer."),
  severity: z.enum(INCIDENT_SEVERITY_VALUES, {
    message: "Please select a valid severity level.",
  }),
  project_id: z.string().trim().min(1, "Please select a project."),
};

export const createIncidentSchema = z.object({
  ...incidentBaseSchema,
  status: z.literal(INCIDENT_DEFAULT_STATUS),
});

export const editIncidentSchema = z.object({
  ...incidentBaseSchema,
  status: z.enum(INCIDENT_STATUS_VALUES, {
    message: "Please select a valid status.",
  }),
});
