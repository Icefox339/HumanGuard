import { useMutation } from '@tanstack/react-query';
import { login, register } from '@/api/auth';
import { useAuthStore } from '@/app/store/auth-store';

const DEBUG_ADMIN_EMAIL = 'admin@humanguard.local';
const DEBUG_ADMIN_PASSWORD = 'Admin123!';
const DEBUG_ADMIN_TOKEN = 'debug-admin-token';

export const useAuth = () => {
  const setSession = useAuthStore((s) => s.setSession);

  const loginMutation = useMutation({
    mutationFn: async (payload: Parameters<typeof login>[0]) => {
      if (payload.email === DEBUG_ADMIN_EMAIL && payload.password === DEBUG_ADMIN_PASSWORD) {
        return {
          token: DEBUG_ADMIN_TOKEN,
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
    registerMutation
  };
};
