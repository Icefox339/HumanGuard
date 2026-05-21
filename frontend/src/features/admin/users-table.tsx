import { useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { getCurrentUser } from '@/api/auth';
import { getUsers, updateUser, UserDetails } from '@/api/users';

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
  const [updatingUserId, setUpdatingUserId] = useState<string | null>(null);
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

  const promoteToAdmin = async (user: UserDetails) => {
    setUpdatingUserId(user.id);
    setError(null);

    try {
      await updateUser(user.id, { role: 'admin' });
      setUsers((prev) => prev.map((item) => (item.id === user.id ? { ...item, role: 'admin' } : item)));
    } catch (e) {
      const err = getError(e);
      setError(`Не удалось назначить пользователя администратором: ${err.message}`);
    } finally {
      setUpdatingUserId(null);
    }
  };

  return (
    <section className="theme-card space-y-4 rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-xl font-semibold text-[rgb(var(--text-primary))]">Пользователи системы</h2>
        <button
          className="interactive-chip rounded-lg border border-[rgb(var(--border))] px-3 py-1.5 text-sm font-medium text-[rgb(var(--text-primary))]"
          onClick={() => void loadUsers()}
        >
          Обновить
        </button>
      </div>

      {loading && <p className="text-sm text-[rgb(var(--text-secondary))]">Загрузка пользователей...</p>}
      {error && <p className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">{error}</p>}

      {!loading && users.length === 0 && !error && <p className="text-sm text-[rgb(var(--text-secondary))]">Пользователей пока нет.</p>}

      {users.length > 0 && (
        <div className="overflow-x-auto rounded-xl border border-[rgb(var(--border))]">
          <table className="min-w-full border-collapse text-sm">
            <thead className="bg-[rgb(var(--bg-main))]">
              <tr className="text-left text-[rgb(var(--text-secondary))]">
                <th className="px-3 py-2">ID</th>
                <th className="px-3 py-2">Email</th>
                <th className="px-3 py-2">Имя</th>
                <th className="px-3 py-2">Роль</th>
                <th className="px-3 py-2">Создан</th>
                <th className="px-3 py-2">Последний вход</th>
                <th className="px-3 py-2">Действия</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id} className="border-t border-[rgb(var(--border))] align-top hover:bg-[rgb(var(--bg-main))]">
                  <td className="max-w-56 truncate px-3 py-2" title={user.id}>{user.id}</td>
                  <td className="px-3 py-2">{user.email}</td>
                  <td className="px-3 py-2">{user.name || '—'}</td>
                  <td className="px-3 py-2">
                    <span className="rounded-full bg-[rgb(var(--bg-main))] px-2 py-0.5 text-xs font-medium text-[rgb(var(--text-secondary))]">{user.role}</span>
                  </td>
                  <td className="px-3 py-2">{user.created_at ? new Date(user.created_at).toLocaleString() : '—'}</td>
                  <td className="px-3 py-2">{user.last_login ? new Date(user.last_login).toLocaleString() : '—'}</td>
                  <td className="px-3 py-2">
                    {user.role === 'admin' ? (
                      <span className="text-xs text-[rgb(var(--text-secondary))]">Уже админ</span>
                    ) : (
                      <button
                        className="interactive-chip rounded-lg border border-[rgb(var(--border))] px-2 py-1 text-xs font-medium text-[rgb(var(--text-primary))] disabled:opacity-60"
                        disabled={updatingUserId === user.id}
                        onClick={() => void promoteToAdmin(user)}
                      >
                        {updatingUserId === user.id ? 'Назначаем...' : 'Сделать админом'}
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
};
