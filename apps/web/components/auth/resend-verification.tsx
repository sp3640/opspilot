'use client';

import { useEffect, useState } from 'react';

import { Button } from '@/components/ui/button';
import { getApiErrorMessage, getApiErrorStatus } from '@/lib/auth-errors';
import { authService } from '@/services/auth-service';

/** Matches the backend's per-account cooldown between verification emails. */
export const RESEND_COOLDOWN_SECONDS = 60;

const inputClass =
  'mt-2 h-12 w-full rounded-2xl border bg-transparent px-4 text-base outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)] disabled:cursor-not-allowed disabled:opacity-50';

interface ResendVerificationProps {
  /**
   * The address to send to. When omitted (the verification link was
   * invalid or expired, so the address is unknown) an email field is shown.
   */
  email?: string;
}

/**
 * "Resend verification email". The backend deliberately answers the same
 * way for every address, so success here only means the request was
 * accepted - the message is worded accordingly - and the button then stays
 * disabled for the cooldown so it cannot be hammered.
 */
export function ResendVerification({ email: knownEmail }: ResendVerificationProps) {
  const [typedEmail, setTypedEmail] = useState('');
  const [isSending, setIsSending] = useState(false);
  const [cooldown, setCooldown] = useState(0);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const email = (knownEmail ?? typedEmail).trim();

  useEffect(() => {
    if (cooldown <= 0) {
      return;
    }
    const timer = window.setTimeout(() => setCooldown((value) => value - 1), 1000);
    return () => window.clearTimeout(timer);
  }, [cooldown]);

  const handleResend = async () => {
    if (!email) {
      setError('Enter the email address you registered with.');
      return;
    }

    setIsSending(true);
    setError(null);
    try {
      await authService.resendVerification({ email });
      setSent(true);
      setCooldown(RESEND_COOLDOWN_SECONDS);
    } catch (caught) {
      setSent(false);
      const status = getApiErrorStatus(caught);
      if (status === 429) {
        setError('Too many requests. Please wait a few minutes before trying again.');
        setCooldown(RESEND_COOLDOWN_SECONDS);
      } else if (status === 400) {
        setError('Enter a valid email address.');
      } else {
        setError(getApiErrorMessage(caught, "We couldn't send the email right now. Please try again in a few minutes."));
      }
    } finally {
      setIsSending(false);
    }
  };

  return (
    <div className="space-y-3">
      {knownEmail === undefined && (
        <div>
          <label htmlFor="resend-email" className="text-sm font-medium">
            Email
          </label>
          <input
            id="resend-email"
            type="email"
            value={typedEmail}
            onChange={(event) => setTypedEmail(event.target.value)}
            placeholder="you@example.com"
            autoComplete="email"
            disabled={isSending}
            className={inputClass}
            style={{ borderColor: 'var(--border)' }}
          />
        </div>
      )}

      <Button
        type="button"
        variant="secondary"
        className="w-full"
        loading={isSending}
        disabled={isSending || cooldown > 0}
        onClick={() => {
          void handleResend();
        }}
      >
        {cooldown > 0 ? `Resend available in ${cooldown}s` : 'Resend verification email'}
      </Button>

      {sent && (
        <p role="status" className="text-sm" style={{ color: 'var(--success-text)' }}>
          If that account is awaiting verification, a new link is on its way. Check your inbox and spam folder.
        </p>
      )}
      {error && (
        <p role="alert" className="text-sm" style={{ color: 'var(--danger-text)' }}>
          {error}
        </p>
      )}
    </div>
  );
}
