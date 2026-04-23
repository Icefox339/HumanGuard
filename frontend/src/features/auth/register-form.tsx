import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Link, useNavigate } from 'react-router-dom';
import { getApiErrorMessage, useRegister } from '@/features/auth/use-auth';

const schema = z
  .object({
    email: z.string().email('Enter a valid email'),
    password: z.string().min(8, 'Minimum 8 characters'),
    confirmPassword: z.string().min(8, 'Minimum 8 characters')
  })
  .refine((data) => data.password === data.confirmPassword, {
    path: ['confirmPassword'],
    message: 'Passwords do not match'
  });

type FormValues = z.infer<typeof schema>;

export const RegisterForm = () => {
  const navigate = useNavigate();
  const { mutate, isPending, error } = useRegister();
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
      onSubmit={handleSubmit(({ email, password }) =>
        mutate(
          { email, password },
          {
            onSuccess: () => navigate('/auth/login')
          }
        )
      )}
    >
      <h1 className="text-2xl font-semibold">Create account</h1>

      <input placeholder="Email" className="w-full rounded border px-3 py-2" {...register('email')} />
      {errors.email && <p className="text-sm text-red-600">{errors.email.message}</p>}

      <input type="password" placeholder="Password" className="w-full rounded border px-3 py-2" {...register('password')} />
      {errors.password && <p className="text-sm text-red-600">{errors.password.message}</p>}

      <input
        type="password"
        placeholder="Confirm password"
        className="w-full rounded border px-3 py-2"
        {...register('confirmPassword')}
      />
      {errors.confirmPassword && <p className="text-sm text-red-600">{errors.confirmPassword.message}</p>}

      {error && <p className="text-sm text-red-600">{getApiErrorMessage(error)}</p>}

      <button disabled={isPending} className="w-full rounded bg-slate-900 py-2 text-white disabled:opacity-60">
        {isPending ? 'Creating...' : 'Create account'}
      </button>

      <p className="text-sm text-slate-600">
        Already have an account?{' '}
        <Link to="/auth/login" className="font-medium text-slate-900 underline">
          Sign in
        </Link>
      </p>
    </form>
  );
};
