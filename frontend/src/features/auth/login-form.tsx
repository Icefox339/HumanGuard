import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useAuth } from '@/features/auth/use-auth';

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
  otp: z.string().optional()
});

type FormValues = z.infer<typeof schema>;

export const LoginForm = () => {
  const { mutate, isPending } = useAuth();
  const { register, handleSubmit, formState: { errors } } = useForm<FormValues>({ resolver: zodResolver(schema) });

  return (
    <form className="w-full space-y-4 rounded border border-slate-200 bg-white p-6" onSubmit={handleSubmit((v) => mutate(v))}>
      <h1 className="text-2xl font-semibold">Login</h1>
      <input placeholder="Email" className="w-full rounded border px-3 py-2" {...register('email')} />
      <input type="password" placeholder="Password" className="w-full rounded border px-3 py-2" {...register('password')} />
      <input placeholder="OTP (optional)" className="w-full rounded border px-3 py-2" {...register('otp')} />
      {errors.email && <p className="text-sm text-red-600">{errors.email.message}</p>}
      <button disabled={isPending} className="w-full rounded bg-slate-900 py-2 text-white">Sign in</button>
    </form>
  );
};
