'use client';

import Link from 'next/link';
import type { ReactNode } from 'react';
import { Suspense, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { toast } from 'sonner';
import axios from 'axios';
import { CheckCircle2, Eye, EyeOff, XCircle } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Logo } from '@/components/layout/logo';
import { cn } from '@/lib/utils';
import { authService } from '@/services/auth-service';
import { GuestGuard } from '@/components/auth/guest-guard';
import { LoadingScreen } from '@/components/auth/loading-screen';

const registerSchema = z
  .object({
    name: z.string().min(2, 'Name must be at least 2 characters'),
    email: z.string().email('Invalid email address'),
    password: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string().min(1, 'Please confirm your password'),
    organizationName: z.string().trim().max(100, 'Workspace name must be 100 characters or fewer').optional(),
  })
  .refine((values) => values.password === values.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  });

type RegisterFormData = z.infer<typeof registerSchema>;

const inputClass =
  'mt-2 h-12 w-full rounded-2xl border bg-transparent px-4 text-base outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)] disabled:cursor-not-allowed disabled:opacity-50';

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

function RegisterPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const redirectTo = searchParams.get('redirect') || '';
  const isInvitedFlow = redirectTo.startsWith('/accept-invitation');
  const loginHref = redirectTo ? `/login?redirect=${encodeURIComponent(redirectTo)}` : '/login';
  const [isLoading, setIsLoading] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  });

  const name = watch('name') ?? '';
  const password = watch('password') ?? '';
  const confirmPassword = watch('confirmPassword') ?? '';
  const showMatchIndicator = confirmPassword.length > 0;
  const passwordsMatch = showMatchIndicator && password === confirmPassword;
  const workspaceNamePlaceholder = name.trim() ? `${name.trim()}'s Workspace` : 'Acme Inc.';

  const onSubmit = async (data: RegisterFormData) => {
    setIsLoading(true);
    setSubmitError(null);

    try {
      await authService.register({
        name: data.name,
        email: data.email,
        password: data.password,
        // The invited flow's organization is determined by the invitation
        // itself, so a workspace name never applies there.
        organizationName: isInvitedFlow ? undefined : data.organizationName,
      });

      toast.success(
        isInvitedFlow
          ? 'Account created successfully. Please sign in to accept your invitation.'
          : 'Account created successfully. Please sign in.'
      );
      router.replace(loginHref);
    } catch (error) {
      let message = 'Registration failed. Please try again.';

      if (axios.isAxiosError(error) && error.response?.data?.message) {
        message = error.response.data.message;
      } else if (error instanceof Error) {
        message = error.message;
      }

      setSubmitError(message);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthShell>
      <div className="mb-8 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">
          {isInvitedFlow ? 'Create your account to join your team' : 'Create your OpsPilot account'}
        </h1>
        <p className="mt-2 text-sm" style={{ color: 'var(--muted-foreground)' }}>
          {isInvitedFlow
            ? "You're accepting an invitation — sign in once your account is ready to join your team's workspace."
            : 'Start managing your operations platform'}
        </p>
      </div>

      {isInvitedFlow && (
        <div
          className="mb-6 rounded-2xl border p-3 text-sm"
          style={{
            backgroundColor: 'color-mix(in srgb, var(--primary) 8%, transparent)',
            borderColor: 'color-mix(in srgb, var(--primary) 25%, var(--border))',
            color: 'var(--foreground)',
          }}
        >
          You&apos;re joining an existing workspace. There&apos;s no need to name one — your invitation already
          determines which organization and role you&apos;ll have.
        </div>
      )}

      {submitError && (
        <div
          role="alert"
          aria-live="assertive"
          className="mb-6 rounded-2xl border p-3 text-sm"
          style={{
            backgroundColor: 'color-mix(in srgb, var(--danger) 12%, transparent)',
            borderColor: 'var(--danger)',
            color: 'var(--danger)',
          }}
        >
          {submitError}
        </div>
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
        <div>
          <label htmlFor="name" className="text-sm font-medium">
            Name
          </label>
          <input
            {...register('name')}
            type="text"
            id="name"
            placeholder="Jane Doe"
            autoComplete="name"
            autoFocus
            disabled={isLoading}
            className={inputClass}
            style={{ borderColor: errors.name ? 'var(--danger)' : 'var(--border)' }}
            aria-invalid={Boolean(errors.name)}
            aria-describedby={errors.name ? 'name-error' : undefined}
          />
          {errors.name && (
            <p id="name-error" className="mt-1.5 text-xs" style={{ color: 'var(--danger)' }}>
              {errors.name.message}
            </p>
          )}
        </div>

        <div>
          <label htmlFor="email" className="text-sm font-medium">
            Email
          </label>
          <input
            {...register('email')}
            type="email"
            id="email"
            placeholder="you@example.com"
            autoComplete="email"
            disabled={isLoading}
            className={inputClass}
            style={{ borderColor: errors.email ? 'var(--danger)' : 'var(--border)' }}
            aria-invalid={Boolean(errors.email)}
            aria-describedby={errors.email ? 'email-error' : undefined}
          />
          {errors.email && (
            <p id="email-error" className="mt-1.5 text-xs" style={{ color: 'var(--danger)' }}>
              {errors.email.message}
            </p>
          )}
        </div>

        {!isInvitedFlow && (
          <div>
            <label htmlFor="organizationName" className="text-sm font-medium">
              Workspace name
            </label>
            <input
              {...register('organizationName')}
              type="text"
              id="organizationName"
              placeholder={workspaceNamePlaceholder}
              autoComplete="organization"
              disabled={isLoading}
              className={inputClass}
              style={{ borderColor: errors.organizationName ? 'var(--danger)' : 'var(--border)' }}
              aria-invalid={Boolean(errors.organizationName)}
              aria-describedby={errors.organizationName ? 'organization-name-error' : 'organization-name-hint'}
            />
            {errors.organizationName ? (
              <p id="organization-name-error" className="mt-1.5 text-xs" style={{ color: 'var(--danger)' }}>
                {errors.organizationName.message}
              </p>
            ) : (
              <p id="organization-name-hint" className="mt-1.5 text-xs" style={{ color: 'var(--muted-foreground)' }}>
                This creates your own new workspace, and you&apos;ll be its admin. Leave blank to use {workspaceNamePlaceholder}.
              </p>
            )}
          </div>
        )}

        <div>
          <label htmlFor="password" className="text-sm font-medium">
            Password
          </label>
          <div className="relative">
            <input
              {...register('password')}
              type={showPassword ? 'text' : 'password'}
              id="password"
              placeholder="••••••••"
              autoComplete="new-password"
              disabled={isLoading}
              className={cn(inputClass, 'pr-11')}
              style={{ borderColor: errors.password ? 'var(--danger)' : 'var(--border)' }}
              aria-invalid={Boolean(errors.password)}
              aria-describedby={errors.password ? 'password-error' : 'password-hint'}
            />
            <button
              type="button"
              onClick={() => setShowPassword((value) => !value)}
              aria-label={showPassword ? 'Hide password' : 'Show password'}
              aria-pressed={showPassword}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded-lg p-1.5 hover:bg-[var(--muted)]"
              style={{ color: 'var(--muted-foreground)' }}
            >
              {showPassword ? (
                <EyeOff aria-hidden="true" className="h-4 w-4" />
              ) : (
                <Eye aria-hidden="true" className="h-4 w-4" />
              )}
            </button>
          </div>
          {errors.password ? (
            <p id="password-error" className="mt-1.5 text-xs" style={{ color: 'var(--danger)' }}>
              {errors.password.message}
            </p>
          ) : (
            <p id="password-hint" className="mt-1.5 text-xs" style={{ color: 'var(--muted-foreground)' }}>
              At least 8 characters
            </p>
          )}
        </div>

        <div>
          <label htmlFor="confirmPassword" className="text-sm font-medium">
            Confirm password
          </label>
          <div className="relative">
            <input
              {...register('confirmPassword')}
              type={showConfirmPassword ? 'text' : 'password'}
              id="confirmPassword"
              placeholder="••••••••"
              autoComplete="new-password"
              disabled={isLoading}
              className={cn(inputClass, 'pr-16')}
              style={{ borderColor: errors.confirmPassword ? 'var(--danger)' : 'var(--border)' }}
              aria-invalid={Boolean(errors.confirmPassword)}
              aria-describedby={errors.confirmPassword ? 'confirm-password-error' : undefined}
            />
            {showMatchIndicator && (
              <span
                className="absolute right-10 top-1/2 -translate-y-1/2"
                style={{ color: passwordsMatch ? 'var(--success)' : 'var(--danger)' }}
              >
                {passwordsMatch ? (
                  <CheckCircle2 aria-hidden="true" className="h-4 w-4" />
                ) : (
                  <XCircle aria-hidden="true" className="h-4 w-4" />
                )}
                <span className="sr-only">{passwordsMatch ? 'Passwords match' : 'Passwords do not match'}</span>
              </span>
            )}
            <button
              type="button"
              onClick={() => setShowConfirmPassword((value) => !value)}
              aria-label={showConfirmPassword ? 'Hide password' : 'Show password'}
              aria-pressed={showConfirmPassword}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded-lg p-1.5 hover:bg-[var(--muted)]"
              style={{ color: 'var(--muted-foreground)' }}
            >
              {showConfirmPassword ? (
                <EyeOff aria-hidden="true" className="h-4 w-4" />
              ) : (
                <Eye aria-hidden="true" className="h-4 w-4" />
              )}
            </button>
          </div>
          {errors.confirmPassword && (
            <p id="confirm-password-error" className="mt-1.5 text-xs" style={{ color: 'var(--danger)' }}>
              {errors.confirmPassword.message}
            </p>
          )}
        </div>

        <Button type="submit" variant="primary" className="w-full" loading={isLoading} disabled={isLoading}>
          Create account
        </Button>
      </form>

      <p className="mt-6 text-center text-sm" style={{ color: 'var(--muted-foreground)' }}>
        Already have an account?{' '}
        <Link href={loginHref} className="font-medium" style={{ color: 'var(--primary)' }}>
          Sign in
        </Link>
      </p>
    </AuthShell>
  );
}

export default function RegisterPage() {
  return (
    <Suspense fallback={<LoadingScreen />}>
      <GuestGuard>
        <RegisterPageContent />
      </GuestGuard>
    </Suspense>
  );
}
