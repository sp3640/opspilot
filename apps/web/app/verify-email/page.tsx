'use client';

import Link from 'next/link';
import { Suspense, useCallback, useEffect, useRef, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { CheckCircle2, Clock, MailX, ServerCrash } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { AuthShell } from '@/components/auth/auth-shell';
import { LoadingScreen } from '@/components/auth/loading-screen';
import { ResendVerification } from '@/components/auth/resend-verification';
import { AUTH_ERROR_CODES, getApiErrorCode } from '@/lib/auth-errors';
import { authService } from '@/services/auth-service';

type VerificationState = 'loading' | 'verified' | 'already_verified' | 'expired' | 'invalid' | 'error';

interface Presentation {
  icon: LucideIcon;
  tone: 'success' | 'danger' | 'warning';
  title: string;
  description: string;
}

const presentations: Record<Exclude<VerificationState, 'loading'>, Presentation> = {
  verified: {
    icon: CheckCircle2,
    tone: 'success',
    title: 'Email verified successfully.',
    description: 'Your account is active. You can now sign in.',
  },
  already_verified: {
    icon: CheckCircle2,
    tone: 'success',
    title: 'This email address is already verified.',
    description: 'Nothing more to do - you can sign in.',
  },
  expired: {
    icon: Clock,
    tone: 'warning',
    title: 'This verification link has expired.',
    description: 'Verification links only work for a limited time. Request a new one to continue.',
  },
  invalid: {
    icon: MailX,
    tone: 'danger',
    title: 'This verification link is invalid.',
    description: 'It may be incomplete, or replaced by a newer email. Request a new link to continue.',
  },
  error: {
    icon: ServerCrash,
    tone: 'danger',
    title: "We couldn't verify your email right now.",
    description: 'Something went wrong on our side or with your connection. Your link is still valid - please try again.',
  },
};

function VerifyEmailContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  // The token is read exactly once and kept in state: it is removed from the
  // address bar below (so it does not linger in history or get copied along
  // with the URL), which changes what useSearchParams returns afterwards.
  const [token] = useState(() => (searchParams.get('token') ?? '').trim());
  const [state, setState] = useState<VerificationState>(token ? 'loading' : 'invalid');
  const started = useRef(false);

  const verify = useCallback(async () => {
    setState('loading');
    try {
      const response = await authService.verifyEmail({ token });
      setState(response.data.status === 'already_verified' ? 'already_verified' : 'verified');
    } catch (error) {
      const code = getApiErrorCode(error);
      if (code === AUTH_ERROR_CODES.verificationTokenExpired) {
        setState('expired');
      } else if (code === AUTH_ERROR_CODES.verificationTokenInvalid) {
        setState('invalid');
      } else {
        setState('error');
      }
    }
  }, [token]);

  useEffect(() => {
    if (started.current) {
      return;
    }
    started.current = true;

    // Drop the one-time token from the URL right away.
    if (typeof window !== 'undefined' && window.location.search) {
      window.history.replaceState(null, '', window.location.pathname);
    }
    if (token) {
      void verify();
    }
  }, [token, verify]);

  if (state === 'loading') {
    return (
      <AuthShell>
        <div className="flex flex-col items-center py-6 text-center" role="status">
          <div
            className="mb-4 inline-flex h-10 w-10 animate-spin rounded-full border-4"
            style={{ borderColor: 'var(--border)', borderTopColor: 'var(--primary)' }}
            aria-hidden="true"
          />
          <h1 className="text-lg font-semibold tracking-tight">Verifying your email…</h1>
        </div>
      </AuthShell>
    );
  }

  const { icon: Icon, tone, title, description } = presentations[state];

  return (
    <AuthShell>
      <div className="text-center">
        <div
          className="mx-auto mb-4 inline-flex rounded-2xl border p-3"
          style={{
            color: `var(--${tone}-text)`,
            backgroundColor: `var(--${tone}-bg)`,
            borderColor: `var(--${tone}-border)`,
          }}
        >
          <Icon aria-hidden="true" className="h-6 w-6" />
        </div>
        <h1 className="text-xl font-semibold tracking-tight">{title}</h1>
        <p className="mt-2 text-sm" style={{ color: 'var(--muted-foreground)' }}>
          {description}
        </p>
      </div>

      <div className="mt-6 space-y-4">
        {(state === 'verified' || state === 'already_verified') && (
          <Button type="button" variant="primary" className="w-full" onClick={() => router.push('/login')}>
            Continue to login
          </Button>
        )}

        {state === 'error' && (
          <Button
            type="button"
            variant="primary"
            className="w-full"
            onClick={() => {
              void verify();
            }}
          >
            Try again
          </Button>
        )}

        {(state === 'expired' || state === 'invalid') && (
          <>
            <ResendVerification />
            <p className="text-center text-sm">
              <Link href="/login" className="font-medium" style={{ color: 'var(--primary)' }}>
                Back to login
              </Link>
            </p>
          </>
        )}
      </div>
    </AuthShell>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense fallback={<LoadingScreen />}>
      <VerifyEmailContent />
    </Suspense>
  );
}
