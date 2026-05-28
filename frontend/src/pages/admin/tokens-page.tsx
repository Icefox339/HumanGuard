import { useEffect, useState } from 'react';
import { AdminApiKey, listAllApiKeys, revokeApiKey } from '@/api/api-keys';

export const TokensPage = () => {
  const [tokens, setTokens] = useState<AdminApiKey[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await listAllApiKeys();
      setTokens(data);
    } catch {
      setError('Не удалось загрузить токены пользователей.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const onDelete = async (id: string) => {
    if (!window.confirm('Удалить токен пользователя?')) return;
    setError(null);
    try {
      await revokeApiKey(id);
      await load();
    } catch {
      setError('Не удалось удалить токен пользователя.');
    }
  };

  return (
    <section className="theme-card space-y-4 rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-xl font-semibold text-[rgb(var(--text-primary))]">Менеджер токенов</h2>
        <button className="interactive-chip rounded-lg border border-[rgb(var(--border))] px-3 py-1.5 text-sm" type="button" onClick={() => void load()}>
          Обновить
        </button>
      </div>

      {loading && <p className="text-sm text-[rgb(var(--text-secondary))]">Загрузка токенов...</p>}
      {error && <p className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">{error}</p>}
      {!loading && tokens.length === 0 && !error && <p className="text-sm text-[rgb(var(--text-secondary))]">Токенов не найдено.</p>}

      {tokens.length > 0 && (
        <div className="responsive-table-wrap rounded-xl border border-[rgb(var(--border))]">
          <table className="min-w-full border-collapse text-sm">
            <thead className="bg-[rgb(var(--bg-main))]">
              <tr className="text-left text-[rgb(var(--text-secondary))]">
                <th className="px-3 py-2">Пользователь</th>
                <th className="px-3 py-2">Email</th>
                <th className="px-3 py-2">Название</th>
                <th className="px-3 py-2">Префикс</th>
                <th className="px-3 py-2">Создан</th>
                <th className="px-3 py-2">Статус</th>
                <th className="px-3 py-2">Действия</th>
              </tr>
            </thead>
            <tbody>
              {tokens.map((token) => (
                <tr key={token.id} className="border-t border-[rgb(var(--border))] align-top hover:bg-[rgb(var(--bg-main))]">
                  <td className="px-3 py-2">{token.user_name || '—'}</td>
                  <td className="px-3 py-2">{token.user_email || '—'}</td>
                  <td className="px-3 py-2">{token.name}</td>
                  <td className="px-3 py-2">{token.prefix}</td>
                  <td className="px-3 py-2">{new Date(token.created_at).toLocaleString()}</td>
                  <td className="px-3 py-2">{token.revoked ? 'Отозван' : 'Активен'}</td>
                  <td className="px-3 py-2">
                    <button
                      className="interactive-chip rounded border border-red-400/40 px-3 py-1 text-sm text-red-300"
                      onClick={() => void onDelete(token.id)}
                      disabled={token.revoked}
                    >
                      {token.revoked ? 'Отозван' : 'Удалить'}
                    </button>
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
