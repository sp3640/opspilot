import { api } from "@/lib/api";
import type { SecretListResponse, SecretQueryParams } from "@/types/secret-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const secretService = {
  async listSecretsByApplication(
    applicationId: string,
    params: SecretQueryParams
  ): Promise<SecretListResponse> {
    const response = await api.get<APIResponse<SecretListResponse>>(
      `/applications/${applicationId}/secrets`,
      { params }
    );
    return response.data.data;
  },
};
