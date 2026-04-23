import { useMutation } from '@tanstack/react-query';
import { AxiosError } from 'axios';
import { login, register } from '@/api/auth';
import { useAuthStore } from '@/app/store/auth-store';

export const useLogin = () => {
  const setSession = useAuthStore((s) => s.setSession);

  return useMutation({
    mutationFn: login,
    onSuccess: ({ token, user }) => {
      setSession(token, user);
    }
  });
};

export const useRegister = () =>
  useMutation({
    mutationFn: register
  });

export const getApiErrorMessage = (error: unknown) => {
  if (error instanceof AxiosError) {
    return ((error.response?.data as { detail?: string; error?: string } | undefined)?.detail ?? (error.response?.data as { detail?: string; error?: string } | undefined)?.error ?? error.message);
  }

  return 'Unexpected error. Please try again.';
};
