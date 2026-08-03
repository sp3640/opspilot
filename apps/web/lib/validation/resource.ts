import { z } from "zod";
import { RESOURCE_HEALTH_VALUES, RESOURCE_KIND_VALUES, RESOURCE_STATUS_VALUES } from "@/lib/constants/resource";

export const resourceFormSchema = z.object({
  project_id: z.string().uuid("Project is required."), parent_resource_id: z.string().uuid().or(z.literal("")),
  kind: z.enum(RESOURCE_KIND_VALUES), name: z.string().trim().min(1, "Resource name is required.").max(255),
  display_name: z.string().max(255), external_id: z.string().max(255), provider: z.string().max(64), region: z.string().max(128), namespace: z.string().max(255), cluster: z.string().max(255),
  status: z.enum(RESOURCE_STATUS_VALUES), health: z.enum(RESOURCE_HEALTH_VALUES), labelsText: z.string(), annotationsText: z.string(), metadataText: z.string(),
});
export type ResourceFormInput = z.infer<typeof resourceFormSchema>;
