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
import { Eye, EyeOff } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Logo } from '@/components/layout/logo';
import { cn } from '@/lib/utils';
import { authService } from '@/services/auth-service';
import { useAuthStore } from '@/store/auth-store';
import { GuestGuard } from '@/components/auth/guest-guard';
import { LoadingScreen } from '@/components/auth/loading-screen';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(1, 'Password is required'),
});

type LoginFormData = z.infer<typeof loginSchema>;

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

function LoginPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const redirectTo = searchParams.get('redirect') || '/';
  const sessionExpired = searchParams.get('sessionExpired') === '1';
  const { setAccessToken, setCurrentUser } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginFormData) => {
    setIsLoading(true);
    setSubmitError(null);

    try {
      // Login and get access token
      const loginResponse = await authService.login(data);
      setAccessToken(loginResponse.data.access_token);

      // Fetch and store current user
      const userResponse = await authService.me();
      setCurrentUser(userResponse.data);

      toast.success('Welcome back!');
      router.replace(redirectTo);
    } catch (error) {
      let message = 'Login failed. Please try again.';

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
        <h1 className="text-2xl font-semibold tracking-tight">Sign in to OpsPilot</h1>
        <p className="mt-2 text-sm" style={{ color: 'var(--muted-foreground)' }}>
          Manage your infrastructure and operations
        </p>
      </div>

      {sessionExpired && !submitError && (
        <div
          role="status"
          className="mb-6 rounded-2xl border p-3 text-sm"
          style={{
            backgroundColor: 'color-mix(in srgb, var(--warning) 12%, transparent)',
            borderColor: 'var(--warning)',
            color: 'var(--warning)',
          }}
        >
          Your session has expired. Please sign in again.
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
          <label htmlFor="email" className="text-sm font-medium">
            Email
          </label>
          <input
            {...register('email')}
            type="email"
            id="email"
            placeholder="you@example.com"
            autoComplete="email"
            autoFocus
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
              autoComplete="current-password"
              disabled={isLoading}
              className={cn(inputClass, 'pr-11')}
              style={{ borderColor: errors.password ? 'var(--danger)' : 'var(--border)' }}
              aria-invalid={Boolean(errors.password)}
              aria-describedby={errors.password ? 'password-error' : undefined}
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
          {errors.password && (
            <p id="password-error" className="mt-1.5 text-xs" style={{ color: 'var(--danger)' }}>
              {errors.password.message}
            </p>
          )}
        </div>

        <Button type="submit" variant="primary" className="w-full" loading={isLoading} disabled={isLoading}>
          Sign in
        </Button>
      </form>

      <p className="mt-6 text-center text-sm" style={{ color: 'var(--muted-foreground)' }}>
        Don&apos;t have an account?{' '}
        <Link
          href={redirectTo === '/' ? '/register' : `/register?redirect=${encodeURIComponent(redirectTo)}`}
          className="font-medium"
          style={{ color: 'var(--primary)' }}
        >
          Create one
        </Link>
      </p>
    </AuthShell>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={<LoadingScreen />}>
      <GuestGuard>
        <LoginPageContent />
      </GuestGuard>
    </Suspense>
  );
}
