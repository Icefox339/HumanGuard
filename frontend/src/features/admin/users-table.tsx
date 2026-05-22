import { useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { getCurrentUser } from '@/api/auth';
import { changeUserPassword, getUsers, updateUser, UserDetails } from '@/api/users';
import { AdminUserSession, deactivateUserSession, getAdminUserSessions } from '@/api/sessions';

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
  const [editingUserId, setEditingUserId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [passwordDraft, setPasswordDraft] = useState('');
  const [sessions, setSessions] = useState<AdminUserSession[]>([]);
  const [sessionsLoading, setSessionsLoading] = useState(false);
  const [deactivatingSessionId, setDeactivatingSessionId] = useState<string | null>(null);

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
    void loadSessions();
  }, []);

  const loadSessions = async () => {
    setSessionsLoading(true);

    try {
      const data = await getAdminUserSessions();
      setSessions(data);
    } catch (e) {
      const err = getError(e);
      setError(`Не удалось загрузить сессии пользователей: ${err.message}`);
      setSessions([]);
    } finally {
      setSessionsLoading(false);
    }
  };

  const saveUser = async (user: UserDetails) => {
    setUpdatingUserId(user.id);
    setError(null);

    try {
      const updated = await updateUser(user.id, {
        name: user.name,
        role: user.role,
        avatar_url: user.avatar_url ?? null
      });
      setUsers((prev) => prev.map((item) => (item.id === user.id ? updated : item)));
      setEditingUserId(null);
    } catch (e) {
      const err = getError(e);
      setError(`Не удалось обновить пользователя: ${err.message}`);
    } finally {
      setUpdatingUserId(null);
    }
  };

  const savePassword = async (user: UserDetails) => {
    if (!passwordDraft || passwordDraft.length < 8) {
      setError('Новый пароль должен быть не короче 8 символов');
      return;
    }

    setUpdatingUserId(user.id);
    setError(null);

    try {
      await changeUserPassword(user.id, { new_password: passwordDraft });
      setPasswordDraft('');
    } catch (e) {
      const err = getError(e);
      setError(`Не удалось изменить пароль: ${err.message}`);
    } finally {
      setUpdatingUserId(null);
    }
  };

  const deactivateSession = async (session: AdminUserSession) => {
    setDeactivatingSessionId(session.id);
    setError(null);

    try {
      await deactivateUserSession(session.id);
      setSessions((prev) => prev.filter((item) => item.id !== session.id));
    } catch (e) {
      const err = getError(e);
      setError(`Не удалось деактивировать сессию: ${err.message}`);
    } finally {
      setDeactivatingSessionId(null);
    }
  };

  return (
    <section className="theme-card space-y-4 rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-xl font-semibold text-[rgb(var(--text-primary))]">Пользователи системы</h2>
        <button className="interactive-chip rounded-lg border border-[rgb(var(--border))] px-3 py-1.5 text-sm font-medium text-[rgb(var(--text-primary))]" onClick={() => void loadUsers()}>
          Обновить
        </button>
      </div>

      {loading && <p className="text-sm text-[rgb(var(--text-secondary))]">Загрузка пользователей...</p>}
      {error && <p className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">{error}</p>}

      {users.length > 0 && (
        <div className="overflow-x-auto rounded-xl border border-[rgb(var(--border))]">
          <table className="min-w-full border-collapse text-sm">
            <thead className="bg-[rgb(var(--bg-main))]"><tr className="text-left text-[rgb(var(--text-secondary))]"><th className="px-3 py-2">Email</th><th className="px-3 py-2">Имя</th><th className="px-3 py-2">Аватар URL</th><th className="px-3 py-2">Роль</th><th className="px-3 py-2">Пароль</th><th className="px-3 py-2">Действия</th></tr></thead>
            <tbody>
              {users.map((user) => {
                const isEditing = editingUserId === user.id;
                return (
                  <tr key={user.id} className="border-t border-[rgb(var(--border))] align-top hover:bg-[rgb(var(--bg-main))]">
                    <td className="px-3 py-2">{user.email}</td>
                    <td className="px-3 py-2">
                      {isEditing ? <input className="w-52 rounded border px-2 py-1" value={user.name || ''} onChange={(e) => setUsers((prev) => prev.map((u) => u.id === user.id ? { ...u, name: e.target.value } : u))} /> : (user.name || '—')}
                    </td>
                    <td className="px-3 py-2">
                      {isEditing ? <input className="w-64 rounded border px-2 py-1" value={user.avatar_url || ''} onChange={(e) => setUsers((prev) => prev.map((u) => u.id === user.id ? { ...u, avatar_url: e.target.value || null } : u))} /> : (user.avatar_url || '—')}
                    </td>
                    <td className="px-3 py-2">
                      {isEditing ? (
                        <select className="rounded border px-2 py-1" value={user.role} onChange={(e) => setUsers((prev) => prev.map((u) => u.id === user.id ? { ...u, role: e.target.value } : u))}>
                          <option value="user">user</option>
                          <option value="admin">admin</option>
                        </select>
                      ) : user.role}
                    </td>
                    <td className="px-3 py-2">
                      <div className="flex gap-2">
                        <input className="w-44 rounded border px-2 py-1" type="password" placeholder="Новый пароль" value={editingUserId === user.id ? passwordDraft : ''} onChange={(e) => setPasswordDraft(e.target.value)} />
                        <button className="interactive-chip rounded-lg border border-[rgb(var(--border))] px-2 py-1 text-xs" disabled={updatingUserId === user.id} onClick={() => void savePassword(user)}>Сменить</button>
                      </div>
                    </td>
                    <td className="px-3 py-2">
                      {isEditing ? (
                        <div className="flex gap-2">
                          <button className="interactive-chip rounded-lg border border-[rgb(var(--border))] px-2 py-1 text-xs" disabled={updatingUserId === user.id} onClick={() => void saveUser(user)}>Сохранить</button>
                          <button className="interactive-chip rounded-lg border border-[rgb(var(--border))] px-2 py-1 text-xs" onClick={() => setEditingUserId(null)}>Отмена</button>
                        </div>
                      ) : (
                        <button className="interactive-chip rounded-lg border border-[rgb(var(--border))] px-2 py-1 text-xs" onClick={() => { setEditingUserId(user.id); setPasswordDraft(''); }}>Редактировать</button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      <div className="space-y-3 rounded-xl border border-[rgb(var(--border))] p-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h3 className="text-lg font-semibold text-[rgb(var(--text-primary))]">Активные пользовательские сессии</h3>
          <button className="interactive-chip rounded-lg border border-[rgb(var(--border))] px-3 py-1.5 text-sm font-medium text-[rgb(var(--text-primary))]" onClick={() => void loadSessions()}>Обновить сессии</button>
        </div>
        {sessionsLoading && <p className="text-sm text-[rgb(var(--text-secondary))]">Загрузка сессий...</p>}
        {!sessionsLoading && sessions.length === 0 && <p className="text-sm text-[rgb(var(--text-secondary))]">Активных сессий нет.</p>}
        {sessions.length > 0 && <div className="overflow-x-auto rounded-xl border border-[rgb(var(--border))]"><table className="min-w-full border-collapse text-sm"><thead className="bg-[rgb(var(--bg-main))]"><tr className="text-left text-[rgb(var(--text-secondary))]"><th className="px-3 py-2">Session ID</th><th className="px-3 py-2">User</th><th className="px-3 py-2">Role</th><th className="px-3 py-2">IP</th><th className="px-3 py-2">Last seen</th><th className="px-3 py-2">Действия</th></tr></thead><tbody>{sessions.map((session) => (<tr key={session.id} className="border-t border-[rgb(var(--border))] align-top hover:bg-[rgb(var(--bg-main))]"><td className="max-w-48 truncate px-3 py-2" title={session.id}>{session.id}</td><td className="px-3 py-2">{session.email}</td><td className="px-3 py-2">{session.role}</td><td className="px-3 py-2">{session.ip || '—'}</td><td className="px-3 py-2">{session.last_seen ? new Date(session.last_seen).toLocaleString() : '—'}</td><td className="px-3 py-2"><button className="interactive-chip rounded-lg border border-red-200 bg-red-50 px-2 py-1 text-xs font-medium text-red-700 disabled:opacity-60" disabled={deactivatingSessionId === session.id} onClick={() => void deactivateSession(session)}>{deactivatingSessionId === session.id ? 'Деактивируем...' : 'Деактивировать'}</button></td></tr>))}</tbody></table></div>}
      </div>
    </section>
  );
};
