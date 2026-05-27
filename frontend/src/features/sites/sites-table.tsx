import { FormEvent, useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { activateSite, createSite, deleteSite, getSites, Site, suspendSite } from '@/api/sites';
import { updateSiteSettings } from '@/api/settings';
import { useAuthStore } from '@/app/store/auth-store';

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
};

export const SitesTable = () => {
  const user = useAuthStore((s) => s.user);
  const [sites, setSites] = useState<Site[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createData, setCreateData] = useState({ name: '', domain: '', origin_server: '', settings: '' });

  const loadSites = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getSites();
      setSites(data);
    } catch (e) {
      setError(parseError(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadSites();
  }, []);

  const onCreate = async (e: FormEvent) => {
    e.preventDefault();
    if (!user?.id) {
      setError('Нужен залогиненный пользователь, чтобы создать сайт.');
      return;
    }

    try {
      setError(null);
      const createdSite = await createSite({
        user_id: user.id,
        name: createData.name,
        domain: createData.domain,
        origin_server: createData.origin_server
      });

      const trimmedSettings = createData.settings.trim();
      if (trimmedSettings) {
        const parsedSettings = JSON.parse(trimmedSettings) as Record<string, unknown>;
        await updateSiteSettings(createdSite.id, parsedSettings);
      }

      setCreateData({ name: '', domain: '', origin_server: '', settings: '' });
      await loadSites();
    } catch (e) {
      setError(parseError(e));
    }
  };

  const onAction = async (action: () => Promise<unknown>) => {
    try {
      setError(null);
      await action();
      await loadSites();
    } catch (e) {
      setError(parseError(e));
    }
  };

  return (
    <section className="space-y-4">
      <form className="theme-card space-y-3 rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm" onSubmit={onCreate}>
        <h2 className="text-lg font-semibold text-[rgb(var(--text-primary))]">Добавить сайт</h2>
        <input className="form-input w-full rounded-lg px-3 py-2" placeholder="Название" value={createData.name} onChange={(e) => setCreateData((p) => ({ ...p, name: e.target.value }))} required />
        <input className="form-input w-full rounded-lg px-3 py-2" placeholder="Домен" value={createData.domain} onChange={(e) => setCreateData((p) => ({ ...p, domain: e.target.value }))} required />
        <input className="form-input w-full rounded-lg px-3 py-2" placeholder="Origin server (например http://localhost:3000)" value={createData.origin_server} onChange={(e) => setCreateData((p) => ({ ...p, origin_server: e.target.value }))} required />
        <textarea
          className="form-input min-h-28 w-full rounded-lg px-3 py-2"
          placeholder={'Настройки сайта (JSON), например:\n{"strict_mode": true, "country_block": ["RU"]}'}
          value={createData.settings}
          onChange={(e) => setCreateData((p) => ({ ...p, settings: e.target.value }))}
        />
        <button className="interactive-chip theme-button px-4 py-2">Создать сайт</button>
      </form>

      <section className="theme-card rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-[rgb(var(--text-primary))]">Сайты</h2>
          <button className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm text-[rgb(var(--text-primary))]" onClick={() => void loadSites()}>
            Обновить
          </button>
        </div>

        {loading && <p className="text-sm text-[rgb(var(--text-secondary))]">Загрузка...</p>}
        {error && <p className="mb-2 field-error">{error}</p>}

        {!loading && sites.length === 0 && <p className="text-sm text-[rgb(var(--text-secondary))]">Сайтов пока нет.</p>}

        <div className="space-y-3">
          {sites.map((site) => (
            <article key={site.id} className="rounded border border-[rgb(var(--border))] p-3">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="font-medium text-[rgb(var(--text-primary))]">{site.name}</p>
                  <p className="text-sm text-[rgb(var(--text-secondary))]">{site.domain}</p>
                  <p className="text-xs text-[rgb(var(--text-secondary))]">status: {site.status}</p>
                </div>
                <div className="flex gap-2">
                  <button className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm text-[rgb(var(--text-primary))]" onClick={() => void onAction(() => activateSite(site.id))}>Activate</button>
                  <button className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm text-[rgb(var(--text-primary))]" onClick={() => void onAction(() => suspendSite(site.id))}>Suspend</button>
                  <button className="interactive-chip rounded bg-red-700 px-3 py-1 text-sm text-white" onClick={() => void onAction(() => deleteSite(site.id))}>Delete</button>
                </div>
              </div>
            </article>
          ))}
        </div>
      </section>
    </section>
  );
};
