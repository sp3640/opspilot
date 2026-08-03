import { z } from "zod";
import {
  CLUSTER_CONNECTION_TYPE_VALUES,
  CLUSTER_PROVIDER_VALUES,
  CLUSTER_STATUS_VALUES,
} from "@/lib/constants/cluster";

export const clusterFormSchema = z.object({
  project_id: z.string().uuid("Project is required."),
  name: z.string().trim().min(1, "Cluster name is required.").max(255, "Name must be 255 characters or less."),
  provider: z.enum(CLUSTER_PROVIDER_VALUES),
  status: z.enum(CLUSTER_STATUS_VALUES),
  connection_type: z.enum(CLUSTER_CONNECTION_TYPE_VALUES),
  kubeconfig_encrypted: z.string().max(50000, "Kubeconfig is too large."),
  api_endpoint: z.string().max(255, "API endpoint must be 255 characters or less."),
  region: z.string().max(128, "Region must be 128 characters or less."),
  version: z.string().max(64, "Version must be 64 characters or less."),
  validation_error: z.string(),
  metadataText: z.string(),
});

export type ClusterFormInput = z.infer<typeof clusterFormSchema>;
