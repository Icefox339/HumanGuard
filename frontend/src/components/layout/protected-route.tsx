import { Navigate } from 'react-router-dom';
import { PropsWithChildren } from 'react';
import { useAuthStore } from '@/app/store/auth-store';

export const ProtectedRoute = ({ children }: PropsWithChildren) => {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  return isAuthenticated ? children : <Navigate to="/auth/login" replace />;
};
