export type CommentResponse = {
  id: number;
  content: string;
  incident_id: number;
  user_id: number;
  created_at: string;
  updated_at: string;
};

export type CommentListResponse = {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  items: CommentResponse[];
};

export type CreateCommentRequest = {
  content: string;
};

export type CommentQueryParams = {
  page?: number;
  limit?: number;
  search?: string;
  sort?: "created_at" | "updated_at";
  order?: "asc" | "desc";
};
