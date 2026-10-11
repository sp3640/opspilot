import Link from 'next/link';
import { MailCheck } from 'lucide-react';

import { ResendVerification } from './resend-verification';

interface CheckEmailPanelProps {
  email: string;
  /** Where "Back to login" goes (carries any `redirect` the user arrived with). */
  loginHref: string;
}

/**
 * Shown after registration. The account exists but cannot sign in until the
 * emailed link is followed. Nothing here (or anywhere) displays the token.
 */
export function CheckEmailPanel({ email, loginHref }: CheckEmailPanelProps) {
  return (
    <div className="text-center">
      <div
        className="mx-auto mb-4 inline-flex rounded-2xl border p-3"
        style={{
          color: 'var(--primary)',
          backgroundColor: 'color-mix(in srgb, var(--primary) 10%, transparent)',
          borderColor: 'color-mix(in srgb, var(--primary) 25%, var(--border))',
        }}
      >
        <MailCheck aria-hidden="true" className="h-6 w-6" />
      </div>

      <h1 className="text-2xl font-semibold tracking-tight">Check your email</h1>
      <p className="mt-2 text-sm" style={{ color: 'var(--muted-foreground)' }}>
        We&apos;ve sent a verification link to <span className="font-medium" style={{ color: 'var(--foreground)' }}>{email}</span>.
        Open it to activate your account, then sign in.
      </p>
      <p className="mt-2 text-xs" style={{ color: 'var(--muted-foreground)' }}>
        Can&apos;t find it? Check your spam folder, or request a new link below.
      </p>

      <div className="mt-6 text-left">
        <ResendVerification email={email} />
      </div>

      <p className="mt-6 text-center text-sm">
        <Link href={loginHref} className="font-medium" style={{ color: 'var(--primary)' }}>
          Back to login
        </Link>
      </p>
    </div>
  );
}
