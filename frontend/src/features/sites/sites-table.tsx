import { FormEvent, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { AxiosError } from 'axios';
import { activateSite, createSite, deleteSite, getSites, Site, suspendSite } from '@/api/sites';
import { updateSiteSettings } from '@/api/settings';
import { addSiteBlacklistIP, getSiteBlacklist, removeSiteBlacklistIP, type BlacklistEntry } from '@/api/blacklist';
import { useAuthStore } from '@/app/store/auth-store';

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
};

const isValidIPv4 = (value: string) => /^(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}$/.test(value);

const parseSiteSettings = (rawSettings: string): Record<string, unknown> | null => {
  const trimmedSettings = rawSettings.trim();
  if (!trimmedSettings) {
    return null;
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(trimmedSettings);
  } catch {
    throw new Error('Настройки сайта должны быть валидным JSON.');
  }

  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error('Настройки сайта должны быть JSON-объектом.');
  }

  const allowedTopLevelKeys = new Set(['collector', 'analyzer', 'reaction']);
  const unsupportedKeys = Object.keys(parsed).filter((key) => !allowedTopLevelKeys.has(key));
  if (unsupportedKeys.length > 0) {
    throw new Error(`Неподдерживаемые ключи настроек: ${unsupportedKeys.join(', ')}.`);
  }

  return parsed as Record<string, unknown>;
};

export const SitesTable = () => {
  const user = useAuthStore((s) => s.user);
  const [sites, setSites] = useState<Site[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createData, setCreateData] = useState({ name: '', domain: '', origin_server: '', settings: '' });
  const [blacklistBySite, setBlacklistBySite] = useState<Record<string, BlacklistEntry[]>>({});
  const [blacklistForms, setBlacklistForms] = useState<Record<string, { ip: string; reason: string }>>({});

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
      const parsedSettings = parseSiteSettings(createData.settings);

      const createdSite = await createSite({
        user_id: user.id,
        name: createData.name,
        domain: createData.domain,
        origin_server: createData.origin_server
      });

      if (parsedSettings) {
        await updateSiteSettings(createdSite.id, parsedSettings);
      }

      setCreateData({ name: '', domain: '', origin_server: '', settings: '' });
      await loadSites();
    } catch (e) {
      setError(parseError(e));
    }
  };

  const loadBlacklist = async (siteId: string) => {
    try {
      const entries = await getSiteBlacklist(siteId);
      setBlacklistBySite((prev) => ({ ...prev, [siteId]: entries }));
    } catch (e) {
      setError(parseError(e));
    }
  };

  const onAddBlacklist = async (siteId: string) => {
    const draft = blacklistForms[siteId] ?? { ip: '', reason: '' };
    if (!draft.ip.trim()) {
      setError('Укажите IP для добавления в blacklist.');
      return;
    }

    if (!isValidIPv4(draft.ip.trim())) {
      setError('Введите корректный IPv4 адрес (например, 192.168.0.10).');
      return;
    }

    try {
      setError(null);
      await addSiteBlacklistIP(siteId, { ip: draft.ip.trim(), reason: draft.reason.trim() || undefined });
      setBlacklistForms((prev) => ({ ...prev, [siteId]: { ip: '', reason: '' } }));
      await loadBlacklist(siteId);
    } catch (e) {
      setError(parseError(e));
    }
  };

  const onDeleteBlacklist = async (siteId: string, ip: string) => {
    try {
      setError(null);
      await removeSiteBlacklistIP(siteId, ip);
      await loadBlacklist(siteId);
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
          placeholder={'Настройки сайта (JSON), например:\n{"collector":{"enabled":true},"analyzer":{"enabled":true},"reaction":{"enabled":true}}'}
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
                  <div className="mt-2 flex gap-2">
                    <Link className="text-xs underline text-[rgb(var(--accent))]" to={`/sites/${site.id}/settings`}>Settings</Link>
                  </div>
                </div>
                <div className="flex gap-2">
                  <button className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm text-[rgb(var(--text-primary))]" onClick={() => void onAction(() => activateSite(site.id))}>Activate</button>
                  <button className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm text-[rgb(var(--text-primary))]" onClick={() => void onAction(() => suspendSite(site.id))}>Suspend</button>
                  <button className="interactive-chip rounded bg-red-700 px-3 py-1 text-sm text-white" onClick={() => void onAction(() => deleteSite(site.id))}>Delete</button>
                </div>
              </div>

              <div className="mt-3 space-y-2 rounded border border-[rgb(var(--border))] p-2">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <p className="text-sm font-medium">Blacklist</p>
                  <button className="interactive-chip rounded border border-[rgb(var(--border))] px-2 py-1 text-xs" type="button" onClick={() => void loadBlacklist(site.id)}>Обновить blacklist</button>
                </div>
                <div className="grid gap-2 sm:grid-cols-3">
                  <input className="form-input rounded px-2 py-1 text-sm" placeholder="IP (например 1.2.3.4)" value={(blacklistForms[site.id]?.ip ?? '')} onChange={(e) => setBlacklistForms((prev) => ({ ...prev, [site.id]: { ip: e.target.value, reason: prev[site.id]?.reason ?? '' } }))} />
                  <input className="form-input rounded px-2 py-1 text-sm" placeholder="Причина (опц.)" value={(blacklistForms[site.id]?.reason ?? '')} onChange={(e) => setBlacklistForms((prev) => ({ ...prev, [site.id]: { ip: prev[site.id]?.ip ?? '', reason: e.target.value } }))} />
                  <button className="interactive-chip theme-button px-3 py-1 text-sm" type="button" onClick={() => void onAddBlacklist(site.id)}>Добавить IP</button>
                </div>
                {(blacklistBySite[site.id] ?? []).length === 0 ? <p className="text-xs text-[rgb(var(--text-secondary))]">Blacklist пуст.</p> : (
                  <div className="space-y-1">
                    {(blacklistBySite[site.id] ?? []).map((entry) => (
                      <div key={`${entry.site_id}:${entry.ip}`} className="flex items-center justify-between gap-2 rounded border border-[rgb(var(--border))] px-2 py-1 text-xs">
                        <span>{entry.ip} {entry.reason ? `— ${entry.reason}` : ''}</span>
                        <button className="interactive-chip rounded border border-red-400/60 px-2 py-0.5 text-red-300" type="button" onClick={() => void onDeleteBlacklist(site.id, entry.ip)}>Удалить</button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </article>
          ))}
        </div>
      </section>
    </section>
  );
};
