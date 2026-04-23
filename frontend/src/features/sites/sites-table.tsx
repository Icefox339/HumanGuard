import { FormEvent, useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { activateSite, createSite, deleteSite, getSites, Site, suspendSite } from '@/api/sites';
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
  const [createData, setCreateData] = useState({ name: '', domain: '', origin_server: '' });

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
      await createSite({ user_id: user.id, ...createData });
      setCreateData({ name: '', domain: '', origin_server: '' });
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
      <form className="space-y-3 rounded border border-slate-200 bg-white p-4" onSubmit={onCreate}>
        <h2 className="text-lg font-semibold">Добавить сайт</h2>
        <input className="w-full rounded border px-3 py-2" placeholder="Название" value={createData.name} onChange={(e) => setCreateData((p) => ({ ...p, name: e.target.value }))} required />
        <input className="w-full rounded border px-3 py-2" placeholder="Домен" value={createData.domain} onChange={(e) => setCreateData((p) => ({ ...p, domain: e.target.value }))} required />
        <input className="w-full rounded border px-3 py-2" placeholder="Origin server (например http://localhost:3000)" value={createData.origin_server} onChange={(e) => setCreateData((p) => ({ ...p, origin_server: e.target.value }))} required />
        <button className="rounded bg-slate-900 px-4 py-2 text-white">Создать сайт</button>
      </form>

      <section className="rounded border border-slate-200 bg-white p-4">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Сайты</h2>
          <button className="rounded border px-3 py-1 text-sm" onClick={() => void loadSites()}>
            Обновить
          </button>
        </div>

        {loading && <p className="text-sm text-slate-600">Загрузка...</p>}
        {error && <p className="mb-2 text-sm text-red-600">{error}</p>}

        {!loading && sites.length === 0 && <p className="text-sm text-slate-600">Сайтов пока нет.</p>}

        <div className="space-y-3">
          {sites.map((site) => (
            <article key={site.id} className="rounded border border-slate-200 p-3">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="font-medium">{site.name}</p>
                  <p className="text-sm text-slate-600">{site.domain}</p>
                  <p className="text-xs text-slate-500">status: {site.status}</p>
                </div>
                <div className="flex gap-2">
                  <button className="rounded border px-3 py-1 text-sm" onClick={() => void onAction(() => activateSite(site.id))}>Activate</button>
                  <button className="rounded border px-3 py-1 text-sm" onClick={() => void onAction(() => suspendSite(site.id))}>Suspend</button>
                  <button className="rounded bg-red-700 px-3 py-1 text-sm text-white" onClick={() => void onAction(() => deleteSite(site.id))}>Delete</button>
                </div>
              </div>
            </article>
          ))}
        </div>
      </section>
    </section>
  );
};
