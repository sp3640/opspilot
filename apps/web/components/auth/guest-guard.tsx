'use client';

import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { useAuthStore, useIsAuthenticated } from '@/store/auth-store';
import { LoadingScreen } from './loading-screen';

interface GuestGuardProps {
  children: React.ReactNode;
}

export function GuestGuard({ children }: GuestGuardProps) {
  const router = useRouter();
  const isBootstrapping = useAuthStore((state) => state.isBootstrapping);
  const isAuthenticated = useIsAuthenticated();

  useEffect(() => {
    if (isBootstrapping) {
      return;
    }

    if (isAuthenticated) {
      router.replace('/');
    }
  }, [isBootstrapping, isAuthenticated, router]);

  if (isBootstrapping) {
    return <LoadingScreen />;
  }

  if (isAuthenticated) {
    return null;
  }

  return children;
}
