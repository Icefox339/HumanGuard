import { FormEvent, useMemo, useState } from 'react';
import { AxiosError } from 'axios';
import {
  changeUserPassword,
  checkEmailExists,
  createUser,
  deleteUser,
  getUserByEmail,
  getUserById,
  updateUser,
  UserDetails
} from '@/api/users';
import { login, LoginResponse } from '@/api/auth';
import { useAuthStore } from '@/app/store/auth-store';

type EndpointResult = {
  title: string;
  status: 'success' | 'error';
  payload: unknown;
};

const cardClass = 'rounded border border-slate-200 bg-white p-4 space-y-3';

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
};

const JsonBox = ({ value }: { value: unknown }) => (
  <pre className="max-h-72 overflow-auto rounded bg-slate-900 p-3 text-xs text-slate-100">{JSON.stringify(value, null, 2)}</pre>
);

export const UsersTable = () => {
  const [result, setResult] = useState<EndpointResult | null>(null);
  const setSession = useAuthStore((s) => s.setSession);

  const [createPayload, setCreatePayload] = useState({
    email: '',
    name: '',
    password_hash: '',
    role: 'user'
  });
  const [id, setId] = useState('');
  const [email, setEmail] = useState('');
  const [passwordPayload, setPasswordPayload] = useState({ old_password: '', new_password: '' });
  const [updatePayload, setUpdatePayload] = useState({ name: '', role: '' });
  const [loginPayload, setLoginPayload] = useState({ email: '', password: '', totp_code: '' });

  const onAction = async (title: string, action: () => Promise<unknown>) => {
    try {
      const payload = await action();
      setResult({ title, status: 'success', payload: payload ?? { ok: true } });
    } catch (error) {
      setResult({ title, status: 'error', payload: { message: parseError(error) } });
    }
  };

  const updateBody = useMemo(
    () => ({
      ...(updatePayload.name.trim() ? { name: updatePayload.name } : {}),
      ...(updatePayload.role.trim() ? { role: updatePayload.role } : {})
    }),
    [updatePayload]
  );

  const handleCreate = (e: FormEvent) => {
    e.preventDefault();
    void onAction('POST /api/users', () => createUser(createPayload));
  };

  return (
    <section className="space-y-4">
      <h1 className="text-2xl font-semibold">Функциональные страницы для user endpoints</h1>
      <p className="text-sm text-slate-600">Каждый блок соответствует одному endpoint из твоего списка.</p>

      <form className={cardClass} onSubmit={handleCreate}>
        <h2 className="font-semibold">POST /api/users — создание пользователя</h2>
        <input className="w-full rounded border px-3 py-2" placeholder="email" value={createPayload.email} onChange={(e) => setCreatePayload((p) => ({ ...p, email: e.target.value }))} required />
        <input className="w-full rounded border px-3 py-2" placeholder="name" value={createPayload.name} onChange={(e) => setCreatePayload((p) => ({ ...p, name: e.target.value }))} required />
        <input className="w-full rounded border px-3 py-2" placeholder="password_hash" value={createPayload.password_hash} onChange={(e) => setCreatePayload((p) => ({ ...p, password_hash: e.target.value }))} required />
        <input className="w-full rounded border px-3 py-2" placeholder="role" value={createPayload.role} onChange={(e) => setCreatePayload((p) => ({ ...p, role: e.target.value }))} required />
        <button className="rounded bg-slate-900 px-4 py-2 text-white">Создать</button>
      </form>

      <div className={cardClass}>
        <h2 className="font-semibold">GET /api/users/{'{id}'} — получить по ID</h2>
        <input className="w-full rounded border px-3 py-2" placeholder="user id" value={id} onChange={(e) => setId(e.target.value)} />
        <button className="rounded bg-slate-900 px-4 py-2 text-white" onClick={() => void onAction('GET /api/users/{id}', () => getUserById(id))}>Получить</button>
      </div>

      <div className={cardClass}>
        <h2 className="font-semibold">GET /api/users/email/{'{email}'} — получить по email</h2>
        <input className="w-full rounded border px-3 py-2" placeholder="email" value={email} onChange={(e) => setEmail(e.target.value)} />
        <button className="rounded bg-slate-900 px-4 py-2 text-white" onClick={() => void onAction('GET /api/users/email/{email}', () => getUserByEmail(email))}>Получить</button>
      </div>

      <div className={cardClass}>
        <h2 className="font-semibold">GET /api/users/exists?email=... — занят ли email</h2>
        <input className="w-full rounded border px-3 py-2" placeholder="email" value={email} onChange={(e) => setEmail(e.target.value)} />
        <button className="rounded bg-slate-900 px-4 py-2 text-white" onClick={() => void onAction('GET /api/users/exists', () => checkEmailExists(email))}>Проверить</button>
      </div>

      <form className={cardClass} onSubmit={(e) => { e.preventDefault(); void onAction('POST /api/users/{id}/password', () => changeUserPassword(id, passwordPayload)); }}>
        <h2 className="font-semibold">POST /api/users/{'{id}'}/password — смена пароля</h2>
        <input className="w-full rounded border px-3 py-2" placeholder="user id" value={id} onChange={(e) => setId(e.target.value)} required />
        <input className="w-full rounded border px-3 py-2" placeholder="old_password" value={passwordPayload.old_password} onChange={(e) => setPasswordPayload((p) => ({ ...p, old_password: e.target.value }))} required />
        <input className="w-full rounded border px-3 py-2" placeholder="new_password" value={passwordPayload.new_password} onChange={(e) => setPasswordPayload((p) => ({ ...p, new_password: e.target.value }))} required />
        <button className="rounded bg-slate-900 px-4 py-2 text-white">Сменить пароль</button>
      </form>

      <form className={cardClass} onSubmit={(e) => { e.preventDefault(); void onAction('PUT /api/users/{id}', () => updateUser(id, updateBody)); }}>
        <h2 className="font-semibold">PUT /api/users/{'{id}'} — обновление юзера</h2>
        <input className="w-full rounded border px-3 py-2" placeholder="user id" value={id} onChange={(e) => setId(e.target.value)} required />
        <input className="w-full rounded border px-3 py-2" placeholder="name (optional)" value={updatePayload.name} onChange={(e) => setUpdatePayload((p) => ({ ...p, name: e.target.value }))} />
        <input className="w-full rounded border px-3 py-2" placeholder="role (optional)" value={updatePayload.role} onChange={(e) => setUpdatePayload((p) => ({ ...p, role: e.target.value }))} />
        <button className="rounded bg-slate-900 px-4 py-2 text-white">Обновить</button>
      </form>

      <div className={cardClass}>
        <h2 className="font-semibold">DELETE /api/users/{'{id}'} — удаление</h2>
        <input className="w-full rounded border px-3 py-2" placeholder="user id" value={id} onChange={(e) => setId(e.target.value)} />
        <button className="rounded bg-red-700 px-4 py-2 text-white" onClick={() => void onAction('DELETE /api/users/{id}', () => deleteUser(id))}>Удалить</button>
      </div>

      <form
        className={cardClass}
        onSubmit={(e) => {
          e.preventDefault();
          void onAction('POST /api/login', async () => {
            const payload = await login({
              email: loginPayload.email,
              password: loginPayload.password,
              ...(loginPayload.totp_code.trim() ? { totp_code: loginPayload.totp_code } : {})
            });

            const loginResponse = payload as LoginResponse | UserDetails;
            if ('token' in loginResponse) {
              setSession(loginResponse.token, loginResponse.user);
              return { ...loginResponse, info: 'Токен сохранён, доступ в защищённые страницы открыт.' };
            }

            return loginResponse;
          });
        }}
      >
        <h2 className="font-semibold">POST /api/login — логин</h2>
        <input className="w-full rounded border px-3 py-2" placeholder="email" value={loginPayload.email} onChange={(e) => setLoginPayload((p) => ({ ...p, email: e.target.value }))} required />
        <input className="w-full rounded border px-3 py-2" placeholder="password" value={loginPayload.password} onChange={(e) => setLoginPayload((p) => ({ ...p, password: e.target.value }))} required />
        <input className="w-full rounded border px-3 py-2" placeholder="totp_code (если требуется бэком)" value={loginPayload.totp_code} onChange={(e) => setLoginPayload((p) => ({ ...p, totp_code: e.target.value }))} />
        <button className="rounded bg-emerald-700 px-4 py-2 text-white">Войти</button>
      </form>

      {result && (
        <section className={`rounded border p-4 ${result.status === 'success' ? 'border-emerald-200 bg-emerald-50' : 'border-red-200 bg-red-50'}`}>
          <h3 className="mb-2 font-semibold">Результат: {result.title}</h3>
          <JsonBox value={result.payload} />
        </section>
      )}
    </section>
  );
};
