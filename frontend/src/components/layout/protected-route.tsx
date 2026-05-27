import { Navigate } from 'react-router-dom';
import { PropsWithChildren, useEffect, useState } from 'react';
import { useAuthStore } from '@/app/store/auth-store';
import { getCurrentUser } from '@/api/auth';

export const ProtectedRoute = ({ children }: PropsWithChildren) => {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const user = useAuthStore((s) => s.user);
  const clearSession = useAuthStore((s) => s.clearSession);
  const setUser = useAuthStore((s) => s.setUser);
  const [loading, setLoading] = useState(isAuthenticated && !user);

  useEffect(() => {
    if (!isAuthenticated || user) {
      setLoading(false);
      return;
    }

    let mounted = true;
    void getCurrentUser()
      .then((currentUser) => {
        if (mounted) {
          setUser({ id: currentUser.id, email: currentUser.email, role: currentUser.role as 'user' | 'admin' });
        }
      })
      .catch(() => {
        if (mounted) {
          clearSession();
        }
      })
      .finally(() => {
        if (mounted) {
          setLoading(false);
        }
      });

    return () => {
      mounted = false;
    };
  }, [isAuthenticated, user, setUser, clearSession]);

  if (loading) {
    return <div className="p-6 text-sm text-[rgb(var(--text-secondary))]">Проверяем сессию...</div>;
  }

  return isAuthenticated ? children : <Navigate to="/auth/login" replace />;
};
