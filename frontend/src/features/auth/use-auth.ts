import { useMutation } from '@tanstack/react-query';
import { login, register } from '@/api/auth';
import { useAuthStore } from '@/app/store/auth-store';

export const useAuth = () => {
  const setSession = useAuthStore((s) => s.setSession);

  const loginMutation = useMutation({
    mutationFn: login,
    onSuccess: (response) => {
      if ('token' in response) {
        setSession(response.token, response.user);
      }
    }
  });

  const registerMutation = useMutation({
    mutationFn: register
  });

  return {
    loginMutation,
    registerMutation
  };
};
