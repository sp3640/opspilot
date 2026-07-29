'use client';

import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { useAuthStore, useIsAuthenticated } from '@/store/auth-store';
import { LoadingScreen } from './loading-screen';

interface AuthGuardProps {
  children: React.ReactNode;
}

export function AuthGuard({ children }: AuthGuardProps) {
  const router = useRouter();
  const isBootstrapping = useAuthStore((state) => state.isBootstrapping);
  const isAuthenticated = useIsAuthenticated();

  useEffect(() => {
    if (isBootstrapping) {
      return;
    }

    if (!isAuthenticated) {
      router.replace('/login');
    }
  }, [isBootstrapping, isAuthenticated, router]);

  if (isBootstrapping) {
    return <LoadingScreen />;
  }

  if (!isAuthenticated) {
    return null;
  }

  return children;
}
