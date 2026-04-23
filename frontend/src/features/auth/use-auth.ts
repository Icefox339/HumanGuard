import { useMutation } from '@tanstack/react-query';
import { login, register } from '@/api/auth';
import { useAuthStore } from '@/app/store/auth-store';

export const useAuth = () => {
  const setSession = useAuthStore((s) => s.setSession);

  const loginMutation = useMutation({
    mutationFn: login,
    onSuccess: ({ token, user }) => setSession(token, user)
  });

  const registerMutation = useMutation({
    mutationFn: register
  });

  return {
    loginMutation,
    registerMutation
  };
};
