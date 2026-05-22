import { Navigate } from 'react-router-dom';
import { PropsWithChildren } from 'react';
import { useAuthStore } from '@/app/store/auth-store';

type RoleRouteProps = PropsWithChildren<{
  requiredRole: 'user' | 'admin';
}>;

export const RoleRoute = ({ children, requiredRole }: RoleRouteProps) => {
  const user = useAuthStore((s) => s.user);

  if (!user) {
    return <Navigate to="/auth/login" replace />;
  }

  return user.role === requiredRole ? children : <Navigate to="/dashboard" replace />;
};
