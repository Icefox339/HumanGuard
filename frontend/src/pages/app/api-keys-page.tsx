import { FormEvent, useEffect, useState } from 'react';
import { createApiKey, CreatedApiKey, listApiKeys, ApiKey, revokeApiKey } from '@/api/api-keys';

const CREATED_KEYS_STORAGE_KEY = 'created_api_keys_by_id';

const readCreatedKeys = (): Record<string, string> => {
  try {
    const raw = window.localStorage.getItem(CREATED_KEYS_STORAGE_KEY);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as unknown;
    if (!parsed || typeof parsed !== 'object') return {};
    return Object.entries(parsed).reduce<Record<string, string>>((acc, [key, value]) => {
      if (typeof value === 'string') {
        acc[key] = value;
      }
      return acc;
    }, {});
  } catch {
    return {};
  }
};

const saveCreatedKeys = (tokensById: Record<string, string>) => {
  window.localStorage.setItem(CREATED_KEYS_STORAGE_KEY, JSON.stringify(tokensById));
};

export const ApiKeysPage = () => {
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [name, setName] = useState('');
  const [expiresInDays, setExpiresInDays] = useState('');
  const [createdKey, setCreatedKey] = useState<CreatedApiKey | null>(null);
  const [createdKeysById, setCreatedKeysById] = useState<Record<string, string>>({});
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
    setCreatedKeysById(readCreatedKeys());
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
      setCreatedKeysById((prev) => {
        const next = { ...prev, [data.id]: data.key };
        saveCreatedKeys(next);
        return next;
      });
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

        <div className="space-y-3">
          {keys.map((key) => (
            <article key={key.id} className="rounded border border-[rgb(var(--border))] p-3">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="font-medium text-[rgb(var(--text-primary))]">{key.name}</p>
                  <p className="text-sm text-[rgb(var(--text-secondary))]">Префикс: {key.prefix}</p>
                  {createdKeysById[key.id] && (
                    <>
                      <p className="mt-1 text-xs text-emerald-300">Сохранённый токен (из этой сессии):</p>
                      <code className="block break-all text-xs text-emerald-300">{createdKeysById[key.id]}</code>
                    </>
                  )}
                  <p className="text-xs text-[rgb(var(--text-secondary))]">Создан: {new Date(key.created_at).toLocaleString()}</p>
                </div>
                <button className="interactive-chip rounded border border-red-400/40 px-3 py-1 text-sm text-red-300" onClick={() => void onDelete(key.id)} disabled={key.revoked}>
                  {key.revoked ? 'Отозван' : 'Удалить'}
                </button>
              </div>
            </article>
          ))}
        </div>
      </section>
    </section>
  );
};
