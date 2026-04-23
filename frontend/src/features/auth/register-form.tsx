import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { register as registerApi } from '@/api/auth';

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(8)
});

type FormValues = z.infer<typeof schema>;

export const RegisterForm = () => {
  const { register, handleSubmit } = useForm<FormValues>({ resolver: zodResolver(schema) });

  return (
    <form className="w-full space-y-4 rounded border border-slate-200 bg-white p-6" onSubmit={handleSubmit((v) => void registerApi(v))}>
      <h1 className="text-2xl font-semibold">Register</h1>
      <input placeholder="Email" className="w-full rounded border px-3 py-2" {...register('email')} />
      <input type="password" placeholder="Password" className="w-full rounded border px-3 py-2" {...register('password')} />
      <button className="w-full rounded bg-slate-900 py-2 text-white">Create account</button>
    </form>
  );
};
