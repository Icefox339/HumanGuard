import { useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { getCurrentUser } from '@/api/auth';
import { getUsers, updateUser, UserDetails } from '@/api/users';

type DraftMap = Record<string, { name: string; role: 'user' | 'admin' }>;

const getError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return {
    status: err.response?.status,
    message: err.response?.data?.error ?? err.message ?? 'Unknown error'
  };
};

const toAdminRole = (role: string): 'user' | 'admin' => (role === 'admin' ? 'admin' : 'user');

export const UsersTable = () => {
  const [users, setUsers] = useState<UserDetails[]>([]);
  const [drafts, setDrafts] = useState<DraftMap>({});
  const [loading, setLoading] = useState(false);
  const [savingId, setSavingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const initDrafts = (data: UserDetails[]) => {
    const nextDrafts: DraftMap = {};
    data.forEach((user) => {
      nextDrafts[user.id] = {
        name: user.name ?? '',
        role: toAdminRole(user.role)
      };
    });
    setDrafts(nextDrafts);
  };

  const loadUsers = async () => {
    setLoading(true);
    setError(null);

    try {
      const data = await getUsers();
      setUsers(data);
      initDrafts(data);
      return;
    } catch (e) {
      const err = getError(e);

      if (err.status === 405) {
        setError('Текущий бэкенд не поддерживает GET /api/users (возвращает 405). Показываю только текущего пользователя.');
        try {
          const me = await getCurrentUser();
          const data: UserDetails[] = [
            {
              id: me.id,
              email: me.email,
              name: 'Текущий пользователь',
              role: me.role,
              created_at: undefined,
              updated_at: undefined,
              last_login: undefined
            }
          ];
          setUsers(data);
          initDrafts(data);
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

  const onSaveUser = async (userId: string) => {
    const draft = drafts[userId];
    if (!draft) {
      return;
    }

    setSavingId(userId);
    setError(null);
    setSuccess(null);

    try {
      await updateUser(userId, {
        name: draft.name,
        role: draft.role
      });

      setUsers((prev) => prev.map((user) => (user.id === userId ? { ...user, name: draft.name, role: draft.role } : user)));
      setSuccess('Данные пользователя успешно обновлены.');
    } catch (e) {
      const err = getError(e);
      setError(err.message);
    } finally {
      setSavingId(null);
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
      {success && <p className="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-800">{success}</p>}

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
              {users.map((user) => {
                const draft = drafts[user.id] ?? { name: user.name ?? '', role: toAdminRole(user.role) };
                const isSaving = savingId === user.id;

                return (
                  <tr key={user.id} className="border-t border-[rgb(var(--border))] align-top hover:bg-[rgb(var(--bg-main))]">
                    <td className="max-w-56 truncate px-3 py-2" title={user.id}>{user.id}</td>
                    <td className="px-3 py-2">{user.email}</td>
                    <td className="px-3 py-2">
                      <input
                        className="form-input w-full min-w-40 rounded-lg px-2 py-1"
                        value={draft.name}
                        onChange={(e) => setDrafts((prev) => ({ ...prev, [user.id]: { ...draft, name: e.target.value } }))}
                      />
                    </td>
                    <td className="px-3 py-2">
                      <select
                        className="form-input rounded-lg px-2 py-1"
                        value={draft.role}
                        onChange={(e) => setDrafts((prev) => ({ ...prev, [user.id]: { ...draft, role: e.target.value as 'user' | 'admin' } }))}
                      >
                        <option value="user">user</option>
                        <option value="admin">admin</option>
                      </select>
                    </td>
                    <td className="px-3 py-2">{user.created_at ? new Date(user.created_at).toLocaleString() : '—'}</td>
                    <td className="px-3 py-2">{user.last_login ? new Date(user.last_login).toLocaleString() : '—'}</td>
                    <td className="px-3 py-2">
                      <button
                        className="interactive-chip rounded-lg border border-[rgb(var(--border))] px-3 py-1.5 text-xs font-medium text-[rgb(var(--text-primary))] disabled:opacity-60"
                        disabled={isSaving}
                        onClick={() => void onSaveUser(user.id)}
                      >
                        {isSaving ? 'Сохраняем...' : 'Сохранить'}
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
};
