import { Navigate } from 'react-router-dom';
import { PropsWithChildren, useEffect, useState } from 'react';
import { useAuthStore } from '@/app/store/auth-store';
import { getCurrentUser } from '@/api/auth';

export const ProtectedRoute = ({ children }: PropsWithChildren) => {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const user = useAuthStore((s) => s.user);
  const setUser = useAuthStore((s) => s.setUser);
  const clearSession = useAuthStore((s) => s.clearSession);
  const [hydrating, setHydrating] = useState(false);

  useEffect(() => {
    const hydrateSession = async () => {
      if (!isAuthenticated || user) {
        return;
      }

      setHydrating(true);
      try {
        const me = await getCurrentUser();
        setUser(me);
      } catch {
        clearSession();
      } finally {
        setHydrating(false);
      }
    };

    void hydrateSession();
  }, [clearSession, isAuthenticated, setUser, user]);

  if (!isAuthenticated) {
    return <Navigate to="/auth/login" replace />;
  }

  if (hydrating) {
    return <div className="p-6 text-sm text-[rgb(var(--text-secondary))]">Проверяем сессию...</div>;
  }

  return children;
};

export const AdminRoute = ({ children }: PropsWithChildren) => {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const user = useAuthStore((s) => s.user);

  if (!isAuthenticated) {
    return <Navigate to="/auth/login" replace />;
  }

  if (user?.role !== 'admin') {
    return <Navigate to="/dashboard" replace />;
  }

  return children;
};
