import { Navigate } from 'react-router-dom';
import { useAuthStore } from '@/app/store/auth-store';
import { RegisterForm } from '@/features/auth/register-form';

export const RegisterPage = () => {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  return <section><RegisterForm /></section>;
};
