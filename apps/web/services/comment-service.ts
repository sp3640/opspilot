import { api } from "@/lib/api";
import type {
  CommentListResponse,
  CommentQueryParams,
  CommentResponse,
  CreateCommentRequest,
} from "@/types/comment-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const commentService = {
  async listComments(incidentId: number, params: CommentQueryParams): Promise<CommentListResponse> {
    const response = await api.get<APIResponse<CommentListResponse>>(
      `/incidents/${incidentId}/comments`,
      { params }
    );
    return response.data.data;
  },

  async createComment(incidentId: number, payload: CreateCommentRequest): Promise<CommentResponse> {
    const response = await api.post<APIResponse<CommentResponse>>(
      `/incidents/${incidentId}/comments`,
      payload
    );
    return response.data.data;
  },
};
