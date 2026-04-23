import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Link } from 'react-router-dom';
import { RegisterResponse } from '@/api/auth';
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
  const { mutate, isPending, error } = useRegister();
  const [createdUser, setCreatedUser] = useState<RegisterResponse | null>(null);
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
            onSuccess: (result) => setCreatedUser(result)
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

      {createdUser && (
        <div className="rounded-md border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-800">
          <p className="font-medium">User created successfully.</p>
          {createdUser.totp_secret && <p>TOTP secret: {createdUser.totp_secret}</p>}
          {createdUser.qr_code_url && (
            <p>
              QR URL:{' '}
              <a href={createdUser.qr_code_url} target="_blank" rel="noreferrer" className="underline">
                Open QR code
              </a>
            </p>
          )}
          <p className="mt-2">Save this secret in authenticator app, then sign in.</p>
        </div>
      )}

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
