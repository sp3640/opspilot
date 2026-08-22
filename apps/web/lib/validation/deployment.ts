import { z } from "zod";
import { DEPLOYMENT_ENVIRONMENT_VALUES, DEPLOYMENT_STRATEGY_VALUES } from "@/lib/constants";

const namespacePattern = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/;

const deploymentBaseSchema = {
  image: z.string().trim().min(1, "Image is required.").max(500, "Image must be 500 characters or fewer."),
  imageTag: z.string().trim().max(128, "Image tag must be 128 characters or fewer.").optional(),
  environment: z.enum(DEPLOYMENT_ENVIRONMENT_VALUES, { message: "Please select a valid environment." }),
  namespace: z
    .string()
    .trim()
    .min(1, "Namespace is required.")
    .max(63, "Namespace must be 63 characters or fewer.")
    .regex(namespacePattern, "Namespace must be a valid Kubernetes namespace name."),
  replicaCount: z.coerce.number().int().min(1, "Replica count must be at least 1."),
  deploymentStrategy: z.enum(DEPLOYMENT_STRATEGY_VALUES, { message: "Please select a valid deployment strategy." }),
  targetClusterId: z.string().trim().min(1, "Please select a target cluster."),
  commitSha: z.string().trim().max(64, "Commit SHA must be 64 characters or fewer.").optional(),
  author: z.string().trim().max(255, "Author must be 255 characters or fewer.").optional(),
};

export const createDeploymentSchema = z.object(deploymentBaseSchema);

export const editDeploymentSchema = z.object(deploymentBaseSchema);
