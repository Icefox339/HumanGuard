import { useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { getCurrentUser } from '@/api/auth';
import { getUsers, UserDetails } from '@/api/users';

const getError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return {
    status: err.response?.status,
    message: err.response?.data?.error ?? err.message ?? 'Unknown error'
  };
};

export const UsersTable = () => {
  const [users, setUsers] = useState<UserDetails[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadUsers = async () => {
    setLoading(true);
    setError(null);

    try {
      const data = await getUsers();
      setUsers(data);
      return;
    } catch (e) {
      const err = getError(e);

      if (err.status === 405) {
        setError('Текущий бэкенд не поддерживает GET /api/users (возвращает 405). Показываю только текущего пользователя.');
        try {
          const me = await getCurrentUser();
          setUsers([
            {
              id: me.id,
              email: me.email,
              name: 'Текущий пользователь',
              role: me.role,
              created_at: undefined,
              updated_at: undefined,
              last_login: undefined
            }
          ]);
        } catch {
          setUsers([]);
        }
      } else {
        setError(err.message);
        setUsers([]);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadUsers();
  }, []);

  return (
    <section className="space-y-4 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-xl font-semibold text-slate-900">Пользователи системы</h2>
        <button
          className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-50"
          onClick={() => void loadUsers()}
        >
          Обновить
        </button>
      </div>

      {loading && <p className="text-sm text-slate-600">Загрузка пользователей...</p>}
      {error && <p className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">{error}</p>}

      {!loading && users.length === 0 && !error && <p className="text-sm text-slate-600">Пользователей пока нет.</p>}

      {users.length > 0 && (
        <div className="overflow-x-auto rounded-xl border border-slate-200">
          <table className="min-w-full border-collapse text-sm">
            <thead className="bg-slate-50">
              <tr className="text-left text-slate-700">
                <th className="px-3 py-2">ID</th>
                <th className="px-3 py-2">Email</th>
                <th className="px-3 py-2">Имя</th>
                <th className="px-3 py-2">Роль</th>
                <th className="px-3 py-2">Создан</th>
                <th className="px-3 py-2">Последний вход</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id} className="border-t border-slate-100 align-top hover:bg-slate-50/60">
                  <td className="max-w-56 truncate px-3 py-2" title={user.id}>{user.id}</td>
                  <td className="px-3 py-2">{user.email}</td>
                  <td className="px-3 py-2">{user.name || '—'}</td>
                  <td className="px-3 py-2">
                    <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-700">{user.role}</span>
                  </td>
                  <td className="px-3 py-2">{user.created_at ? new Date(user.created_at).toLocaleString() : '—'}</td>
                  <td className="px-3 py-2">{user.last_login ? new Date(user.last_login).toLocaleString() : '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
};
