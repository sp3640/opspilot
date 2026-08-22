import { z } from "zod";
import {
  CLUSTER_CONNECTION_TYPE_VALUES,
  CLUSTER_PROVIDER_VALUES,
} from "@/lib/constants/cluster";

// Status is server-owned (set by creation / real connectivity validation),
// so it's intentionally not part of either form schema.
const baseClusterFields = {
  project_id: z.string().uuid("Project is required."),
  name: z.string().trim().min(1, "Cluster name is required.").max(255, "Name must be 255 characters or less."),
  provider: z.enum(CLUSTER_PROVIDER_VALUES),
  connection_type: z.enum(CLUSTER_CONNECTION_TYPE_VALUES),
  api_endpoint: z.string().max(255, "API endpoint must be 255 characters or less."),
  region: z.string().max(128, "Region must be 128 characters or less."),
  metadataText: z.string(),
};

// Create: a kubeconfig is required — there is no existing credential to fall
// back to.
export const createClusterFormSchema = z.object({
  ...baseClusterFields,
  kubeconfig_encrypted: z
    .string()
    .trim()
    .min(1, "A kubeconfig is required to connect a cluster.")
    .max(50000, "Kubeconfig is too large."),
});

// Edit: the backend never returns the stored credential, so this field
// starts blank and stays optional — leaving it blank preserves the existing
// credential unchanged; only a non-empty value replaces it.
export const editClusterFormSchema = z.object({
  ...baseClusterFields,
  kubeconfig_encrypted: z.string().max(50000, "Kubeconfig is too large."),
});

// Both schemas share the exact same field shape (only the kubeconfig
// constraint differs), so a single input type covers both create and edit
// forms — this is what ClusterFormFields is typed against.
export type ClusterFormInput = z.infer<typeof editClusterFormSchema>;
