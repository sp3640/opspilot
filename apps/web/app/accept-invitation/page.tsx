'use client';

import type { ReactNode } from 'react';
import { Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import axios from 'axios';
import { Link2 } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Logo } from '@/components/layout/logo';
import { EmptyState, StatusBadge } from '@/components/common';
import { LoadingScreen } from '@/components/auth/loading-screen';
import { authService } from '@/services/auth-service';
import { useAcceptInvitation, useValidateInvitation } from '@/hooks/use-invitations';
import { useAuthStore, useIsAuthenticated } from '@/store/auth-store';

function AuthShell({ children }: { children: ReactNode }) {
  return (
    <div
      className="flex min-h-screen flex-col items-center justify-center px-4 py-12"
      style={{ backgroundColor: 'var(--background)', color: 'var(--foreground)' }}
    >
      <div className="mb-8">
        <Logo />
      </div>
      <div
        className="w-full max-w-md rounded-[var(--radius)] border p-6 sm:p-8"
        style={{ backgroundColor: 'var(--card)', borderColor: 'var(--border)', boxShadow: 'var(--shadow-lg)' }}
      >
        {children}
      </div>
    </div>
  );
}

function getInvitationErrorMessage(error: unknown): string {
  if (axios.isAxiosError<{ message?: string }>(error) && error.response?.data?.message) {
    return error.response.data.message;
  }
  return 'This invitation link is no longer valid.';
}

function AcceptInvitationContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = (searchParams.get('token') ?? '').trim();

  const isBootstrapping = useAuthStore((state) => state.isBootstrapping);
  const setAccessToken = useAuthStore((state) => state.setAccessToken);
  const setCurrentUser = useAuthStore((state) => state.setCurrentUser);
  const isAuthenticated = useIsAuthenticated();

  // Preserve the token through the auth pages without ever persisting it —
  // it lives only in this URL's query string.
  const returnTo = `/accept-invitation?token=${encodeURIComponent(token)}`;

  const {
    data: invitation,
    isLoading: isValidating,
    isError: isValidateError,
    error: validateError,
  } = useValidateInvitation(token, isAuthenticated && Boolean(token));

  const acceptInvitation = useAcceptInvitation();

  if (isBootstrapping) {
    return <LoadingScreen />;
  }

  if (!token) {
    return (
      <AuthShell>
        <EmptyState
          icon={Link2}
          title="Invitation link is missing"
          description="A valid invitation link is required to continue. Ask the person who invited you to send it again."
          action={
            <Button type="button" variant="secondary" onClick={() => router.push('/login')}>
              Back to login
            </Button>
          }
        />
      </AuthShell>
    );
  }

  if (!isAuthenticated) {
    return (
      <AuthShell>
        <div className="mb-6 text-center">
          <h1 className="text-2xl font-semibold tracking-tight">You&apos;ve been invited to OpsPilot</h1>
          <p className="mt-2 text-sm" style={{ color: 'var(--muted-foreground)' }}>
            Sign in or create an account to continue with your invitation.
          </p>
        </div>
        <div className="flex flex-col gap-3">
          <Button
            type="button"
            variant="primary"
            className="w-full"
            onClick={() => router.push(`/login?redirect=${encodeURIComponent(returnTo)}`)}
          >
            Sign in
          </Button>
          <Button
            type="button"
            variant="secondary"
            className="w-full"
            onClick={() => router.push(`/register?redirect=${encodeURIComponent(returnTo)}`)}
          >
            Create account
          </Button>
        </div>
      </AuthShell>
    );
  }

  if (isValidating) {
    return (
      <AuthShell>
        <div className="flex flex-col items-center py-6 text-center">
          <div
            className="mb-4 inline-flex h-10 w-10 animate-spin rounded-full border-4"
            style={{ borderColor: 'var(--border)', borderTopColor: 'var(--primary)' }}
            aria-hidden="true"
          />
          <h1 className="text-lg font-semibold tracking-tight">Checking invitation</h1>
        </div>
      </AuthShell>
    );
  }

  if (isValidateError || !invitation) {
    return (
      <AuthShell>
        <div className="text-center">
          <div
            className="mx-auto mb-4 inline-flex rounded-2xl border p-3"
            style={{
              color: 'var(--danger)',
              backgroundColor: 'color-mix(in srgb, var(--danger) 10%, transparent)',
              borderColor: 'color-mix(in srgb, var(--danger) 25%, var(--border))',
            }}
          >
            <Link2 aria-hidden="true" className="h-6 w-6" />
          </div>
          <h1 className="text-xl font-semibold tracking-tight">This invitation can&apos;t be accepted</h1>
          <p className="mt-2 text-sm" style={{ color: 'var(--muted-foreground)' }}>
            {getInvitationErrorMessage(validateError)}
          </p>
          <Button type="button" variant="secondary" className="mt-6" onClick={() => router.push('/login')}>
            Back to login
          </Button>
        </div>
      </AuthShell>
    );
  }

  const handleAccept = async () => {
    try {
      await acceptInvitation.mutateAsync({ token });

      // Acceptance updates the organization/role in the database, but the
      // current access token was minted before that change and still
      // carries the old claims. Reissue it so this session's authorization
      // reflects the new organization/role immediately, without requiring
      // a full re-login.
      const reissued = await authService.reissue();
      setAccessToken(reissued.data.access_token);

      const me = await authService.me();
      setCurrentUser(me.data);
      router.replace('/');
    } catch {
      // useAcceptInvitation already surfaces an error toast; stay on this
      // page so the user can see it and try again if appropriate.
    }
  };

  return (
    <AuthShell>
      <div className="mb-6 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">You&apos;re invited to join OpsPilot</h1>
      </div>

      <div
        className="mb-6 rounded-2xl border p-4"
        style={{ borderColor: 'var(--border)', backgroundColor: 'color-mix(in srgb, var(--muted) 45%, transparent)' }}
      >
        <p className="text-xs" style={{ color: 'var(--muted-foreground)' }}>
          You&apos;re joining
        </p>
        <p className="mt-1 text-lg font-semibold tracking-tight">{invitation.organizationName}</p>
        <div className="mt-3">
          <StatusBadge variant="info">{invitation.role}</StatusBadge>
        </div>
      </div>

      <Button
        type="button"
        variant="primary"
        className="w-full"
        loading={acceptInvitation.isPending}
        disabled={acceptInvitation.isPending}
        onClick={() => {
          void handleAccept();
        }}
      >
        Accept invitation
      </Button>
    </AuthShell>
  );
}

export default function AcceptInvitationPage() {
  return (
    <Suspense fallback={<LoadingScreen />}>
      <AcceptInvitationContent />
    </Suspense>
  );
}
