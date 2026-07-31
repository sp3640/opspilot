import { useEffect, useRef } from 'react';
import { useAuthStore } from '@/store/auth-store';
import { authService } from '@/services/auth-service';

export const useAuthBootstrap = () => {
  const { setAccessToken, setCurrentUser, setBootstrapping, logout } = useAuthStore();
  const hasBootstrappedRef = useRef(false);

  useEffect(() => {
    if (hasBootstrappedRef.current) {
      return;
    }

    hasBootstrappedRef.current = true;

    const bootstrap = async () => {
      try {
        const token = localStorage.getItem('accessToken');

        if (token) {
          setAccessToken(token);
          const response = await authService.me();
          setCurrentUser(response.data);
        }
      } catch {
        logout();
      } finally {
        setBootstrapping(false);
      }
    };

    bootstrap();
  }, [setAccessToken, setCurrentUser, setBootstrapping, logout]);
};


