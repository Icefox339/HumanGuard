import { useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { getUsers, UserDetails } from '@/api/users';

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
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
    } catch (e) {
      setError(parseError(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadUsers();
  }, []);

  return (
    <section className="space-y-4 rounded border border-slate-200 bg-white p-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Пользователи системы</h2>
        <button className="rounded border px-3 py-1 text-sm" onClick={() => void loadUsers()}>
          Обновить
        </button>
      </div>

      {loading && <p className="text-sm text-slate-600">Загрузка пользователей...</p>}
      {error && <p className="text-sm text-red-600">{error}</p>}

      {!loading && users.length === 0 && <p className="text-sm text-slate-600">Пользователей пока нет.</p>}

      {users.length > 0 && (
        <div className="overflow-x-auto">
          <table className="min-w-full border-collapse text-sm">
            <thead>
              <tr className="border-b border-slate-200 text-left">
                <th className="px-2 py-2">ID</th>
                <th className="px-2 py-2">Email</th>
                <th className="px-2 py-2">Имя</th>
                <th className="px-2 py-2">Роль</th>
                <th className="px-2 py-2">Создан</th>
                <th className="px-2 py-2">Последний вход</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id} className="border-b border-slate-100 align-top">
                  <td className="max-w-56 truncate px-2 py-2" title={user.id}>{user.id}</td>
                  <td className="px-2 py-2">{user.email}</td>
                  <td className="px-2 py-2">{user.name || '—'}</td>
                  <td className="px-2 py-2">{user.role}</td>
                  <td className="px-2 py-2">{user.created_at ? new Date(user.created_at).toLocaleString() : '—'}</td>
                  <td className="px-2 py-2">{user.last_login ? new Date(user.last_login).toLocaleString() : '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
};
