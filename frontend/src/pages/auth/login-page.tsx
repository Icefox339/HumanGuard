import { Navigate } from 'react-router-dom';
import { useAuthStore } from '@/app/store/auth-store';
import { LoginForm } from '@/features/auth/login-form';

export const LoginPage = () => {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  return <section><LoginForm /></section>;
};
