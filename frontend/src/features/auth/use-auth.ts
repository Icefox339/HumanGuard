import { useMutation } from '@tanstack/react-query';
import { login } from '@/api/auth';
import { useAuthStore } from '@/app/store/auth-store';

export const useAuth = () => {
  const setSession = useAuthStore((s) => s.setSession);
  return useMutation({
    mutationFn: login,
    onSuccess: ({ token, user }) => setSession(token, user)
  });
};
