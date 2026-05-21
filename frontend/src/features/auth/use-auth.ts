import { useMutation } from '@tanstack/react-query';
import { LoginPayload, login, register } from '@/api/auth';
import { useAuthStore } from '@/app/store/auth-store';

const DEBUG_ADMIN_EMAIL = 'admin@humanguard.local';
const DEBUG_ADMIN_PASSWORD = 'Admin#2026';
const ENABLE_DEBUG_ADMIN = import.meta.env.DEV;

export const useAuth = () => {
  const setSession = useAuthStore((s) => s.setSession);

  const loginMutation = useMutation({
    mutationFn: async (payload: LoginPayload) => {
      if (ENABLE_DEBUG_ADMIN && payload.email === DEBUG_ADMIN_EMAIL && payload.password === DEBUG_ADMIN_PASSWORD) {
        return {
          token: 'debug-admin-token',
          user: {
            id: 'debug-admin',
            email: DEBUG_ADMIN_EMAIL,
            role: 'admin' as const
          }
        };
      }

      return login(payload);
    },
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
    registerMutation,
    debugAdmin: ENABLE_DEBUG_ADMIN
      ? {
        email: DEBUG_ADMIN_EMAIL,
        password: DEBUG_ADMIN_PASSWORD
      }
      : null
  };
};
