import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Link, useNavigate } from 'react-router-dom';
import { getApiErrorMessage, useLogin } from '@/features/auth/use-auth';

const schema = z.object({
  email: z.string().email('Enter a valid email'),
  password: z.string().min(8, 'Minimum 8 characters'),
  otp: z.string().min(6, 'Enter OTP from authenticator app')
});

type FormValues = z.infer<typeof schema>;

export const LoginForm = () => {
  const navigate = useNavigate();
  const { mutate, isPending, error } = useLogin();
  const {
    register,
    handleSubmit,
    formState: { errors }
  } = useForm<FormValues>({
    resolver: zodResolver(schema)
  });

  return (
    <form
      className="w-full space-y-4 rounded border border-slate-200 bg-white p-6"
      onSubmit={handleSubmit((values) => mutate(values, { onSuccess: () => navigate('/dashboard') }))}
    >
      <h1 className="text-2xl font-semibold">Login</h1>

      <input placeholder="Email" className="w-full rounded border px-3 py-2" {...register('email')} />
      {errors.email && <p className="text-sm text-red-600">{errors.email.message}</p>}

      <input type="password" placeholder="Password" className="w-full rounded border px-3 py-2" {...register('password')} />
      {errors.password && <p className="text-sm text-red-600">{errors.password.message}</p>}

      <input placeholder="OTP code" className="w-full rounded border px-3 py-2" {...register('otp')} />
      {errors.otp && <p className="text-sm text-red-600">{errors.otp.message}</p>}

      {error && <p className="text-sm text-red-600">{getApiErrorMessage(error)}</p>}

      <button disabled={isPending} className="w-full rounded bg-slate-900 py-2 text-white disabled:opacity-60">
        {isPending ? 'Signing in...' : 'Sign in'}
      </button>

      <Link
        to="/auth/register"
        className="block w-full rounded border border-slate-300 py-2 text-center font-medium text-slate-800 hover:bg-slate-100"
      >
        Sign up
      </Link>
    </form>
  );
};
