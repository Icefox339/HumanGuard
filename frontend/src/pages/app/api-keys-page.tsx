import { FormEvent, useEffect, useState } from 'react';
import { createApiKey, CreatedApiKey, listApiKeys, ApiKey, revokeApiKey } from '@/api/api-keys';

export const ApiKeysPage = () => {
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [name, setName] = useState('');
  const [expiresInDays, setExpiresInDays] = useState('');
  const [createdKey, setCreatedKey] = useState<CreatedApiKey | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await listApiKeys();
      setKeys(data);
    } catch {
      setError('Не удалось загрузить API ключи.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const onCreate = async (event: FormEvent) => {
    event.preventDefault();
    if (!name.trim()) return;

    setSaving(true);
    setError(null);
    try {
      const expires = Number(expiresInDays);
      const data = await createApiKey({
        name: name.trim(),
        ...(expiresInDays && Number.isFinite(expires) && expires > 0 ? { expires_in_days: expires } : {})
      });
      setCreatedKey(data);
      setName('');
      setExpiresInDays('');
      await load();
    } catch {
      setError('Не удалось создать API ключ.');
    } finally {
      setSaving(false);
    }
  };

  const onDelete = async (id: string) => {
    if (!window.confirm('Удалить (отозвать) ключ?')) return;
    setError(null);
    try {
      await revokeApiKey(id);
      await load();
    } catch {
      setError('Не удалось удалить (отозвать) ключ.');
    }
  };

  return (
    <section className="space-y-4">
      <header className="theme-card rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm">
        <h2 className="text-lg font-semibold text-[rgb(var(--text-primary))]">API токены</h2>
        <p className="mt-1 text-sm text-[rgb(var(--text-secondary))]">Создавайте, просматривайте и удаляйте токены доступа.</p>

        <form className="mt-4 flex flex-wrap items-end gap-3" onSubmit={(e) => void onCreate(e)}>
          <div>
            <label className="block text-sm text-[rgb(var(--text-secondary))]">Название</label>
            <input className="form-input rounded-lg px-3 py-2" value={name} onChange={(e) => setName(e.target.value)} placeholder="CI token" />
          </div>
          <div>
            <label className="block text-sm text-[rgb(var(--text-secondary))]">Срок (дней, опц.)</label>
            <input className="form-input rounded-lg px-3 py-2" value={expiresInDays} onChange={(e) => setExpiresInDays(e.target.value)} placeholder="30" inputMode="numeric" />
          </div>
          <button className="interactive-chip theme-button px-4 py-2 disabled:opacity-60" type="submit" disabled={saving || !name.trim()}>
            {saving ? 'Создание…' : 'Создать токен'}
          </button>
        </form>

        {createdKey && (
          <div className="mt-4 rounded-lg border border-emerald-500/40 bg-emerald-500/10 p-3">
            <p className="text-sm font-medium text-[rgb(var(--text-primary))]">Новый токен (показывается один раз):</p>
            <code className="mt-1 block break-all text-sm text-emerald-300">{createdKey.key}</code>
          </div>
        )}

        {error && <p className="mt-3 field-error">{error}</p>}
      </header>

      <section className="theme-card rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm">
        <div className="mb-3 flex items-center justify-between">
          <h3 className="text-base font-semibold text-[rgb(var(--text-primary))]">Список токенов</h3>
          <button className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm" onClick={() => void load()} type="button">Обновить</button>
        </div>

        {loading && <p className="text-sm text-[rgb(var(--text-secondary))]">Загрузка...</p>}
        {!loading && keys.length === 0 && <p className="text-sm text-[rgb(var(--text-secondary))]">Токенов пока нет.</p>}

        <div className="responsive-table-wrap rounded-xl border border-[rgb(var(--border))]">
          <table className="min-w-full border-collapse text-sm">
            <thead className="bg-[rgb(var(--bg-main))]">
              <tr className="text-left text-[rgb(var(--text-secondary))]">
                <th className="px-3 py-2">Название</th>
                <th className="px-3 py-2">Префикс</th>
                <th className="px-3 py-2">Создан</th>
                <th className="px-3 py-2">Истекает</th>
                <th className="px-3 py-2">Последнее использование</th>
                <th className="px-3 py-2">Статус</th>
                <th className="px-3 py-2">Токен</th>
                <th className="px-3 py-2">Действия</th>
              </tr>
            </thead>
            <tbody>
              {keys.map((key) => (
                <tr key={key.id} className="border-t border-[rgb(var(--border))] align-top hover:bg-[rgb(var(--bg-main))]">
                  <td className="px-3 py-2 font-medium text-[rgb(var(--text-primary))]">{key.name}</td>
                  <td className="px-3 py-2">{key.prefix}</td>
                  <td className="px-3 py-2">{new Date(key.created_at).toLocaleString()}</td>
                  <td className="px-3 py-2">{key.expires_at ? new Date(key.expires_at).toLocaleString() : '—'}</td>
                  <td className="px-3 py-2">{key.last_used_at ? new Date(key.last_used_at).toLocaleString() : '—'}</td>
                  <td className="px-3 py-2">{key.revoked ? 'Отозван' : 'Активен'}</td>
                  <td className="px-3 py-2">
                    <code className="block max-w-xs break-all text-xs text-emerald-300">{key.key}</code>
                  </td>
                  <td className="px-3 py-2">
                    <button className="interactive-chip rounded border border-red-400/40 px-3 py-1 text-sm text-red-300" onClick={() => void onDelete(key.id)} disabled={key.revoked}>
                      {key.revoked ? 'Отозван' : 'Удалить'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </section>
  );
};
