"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { MessageSquare, Send } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useComments, useCreateComment } from "@/hooks/use-comments";
import { useHasPermission } from "@/store/auth-store";

const commentFormSchema = z.object({
  content: z
    .string()
    .trim()
    .min(1, "Comment cannot be empty.")
    .max(2000, "Comment must be 2000 characters or fewer."),
});

type CommentFormInput = z.infer<typeof commentFormSchema>;

const COMMENT_QUERY_PARAMS = { page: 1, limit: 50, sort: "created_at" as const, order: "asc" as const };

/** Incident comments thread, rendered only inside the Incident Details Drawer. */
export function IncidentComments({ incidentId }: { incidentId: number }) {
  const canPostComments = useHasPermission("incident:comment");
  const { data, error, isError, isLoading, refetch } = useComments(incidentId, COMMENT_QUERY_PARAMS);
  const createComment = useCreateComment();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CommentFormInput>({
    resolver: zodResolver(commentFormSchema),
    defaultValues: { content: "" },
  });

  const comments = data?.items ?? [];

  const submit = async (input: CommentFormInput) => {
    try {
      await createComment.mutateAsync({ incidentId, payload: { content: input.content } });
      reset();
    } catch {
      // Error is surfaced via the mutation's error toast.
    }
  };

  const inputClass =
    "w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

  return (
    <div className="space-y-5">
      {canPostComments ? (
        <form onSubmit={handleSubmit(submit)} className="space-y-2">
          <label htmlFor="incident-comment-content" className="text-sm font-medium">
            Add a comment
          </label>
          <textarea
            id="incident-comment-content"
            {...register("content")}
            rows={3}
            placeholder="Share an update or note about this incident."
            disabled={createComment.isPending}
            className={inputClass}
            style={{ borderColor: errors.content ? "var(--danger)" : "var(--border)" }}
            aria-invalid={Boolean(errors.content)}
            aria-describedby={errors.content ? "incident-comment-content-error" : undefined}
          />
          {errors.content && (
            <p id="incident-comment-content-error" className="text-xs" style={{ color: "var(--danger)" }}>
              {errors.content.message}
            </p>
          )}
          <div className="flex justify-end">
            <Button type="submit" loading={createComment.isPending} disabled={createComment.isPending}>
              <Send aria-hidden="true" className="h-4 w-4" />
              Post comment
            </Button>
          </div>
        </form>
      ) : null}

      {isError ? (
        <ErrorState
          description={error instanceof Error ? error.message : "Unable to load comments. Please try again."}
          onRetry={() => {
            void refetch();
          }}
        />
      ) : isLoading ? (
        <ListSkeleton rows={3} />
      ) : comments.length === 0 ? (
        <EmptyState
          icon={MessageSquare}
          title="No comments yet"
          description="Be the first to add an update on this incident."
        />
      ) : (
        <ul className="space-y-3">
          {comments.map((comment) => (
            <li key={comment.id} className="rounded-2xl border p-4" style={{ borderColor: "var(--border)" }}>
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm font-medium">User #{comment.user_id}</p>
                <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
                  {formatDateTime(comment.created_at)}
                </p>
              </div>
              <p className="mt-2 text-sm leading-6">{comment.content}</p>
              {comment.updated_at !== comment.created_at ? (
                <p className="mt-2 text-xs" style={{ color: "var(--muted-foreground)" }}>
                  Edited {formatDateTime(comment.updated_at)}
                </p>
              ) : null}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString();
}
