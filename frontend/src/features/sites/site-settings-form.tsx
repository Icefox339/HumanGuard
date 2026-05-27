import { FormEvent, useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { useParams } from 'react-router-dom';
import { getSiteSettings, updateSiteSettings } from '@/api/settings';

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
};

export const SiteSettingsForm = () => {
  const { siteId } = useParams();
  const [rawSettings, setRawSettings] = useState('');
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadSettings = async () => {
    if (!siteId) {
      setError('Site ID не найден в URL.');
      return;
    }

    setLoading(true);
    setError(null);
    setStatus(null);

    try {
      const data = await getSiteSettings(siteId);
      setRawSettings(JSON.stringify(data, null, 2));
    } catch (e) {
      setError(parseError(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadSettings();
  }, [siteId]);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!siteId) return;

    setSaving(true);
    setError(null);
    setStatus(null);

    try {
      const parsed = JSON.parse(rawSettings) as Record<string, unknown>;
      await updateSiteSettings(siteId, parsed);
      setStatus('Настройки успешно сохранены.');
      await loadSettings();
    } catch (e) {
      if (e instanceof SyntaxError) {
        setError('Настройки должны быть валидным JSON.');
      } else {
        setError(parseError(e));
      }
    } finally {
      setSaving(false);
    }
  };

  return (
    <form className="theme-card space-y-3 rounded-2xl border border-[rgb(var(--border))] p-5 text-[rgb(var(--text-primary))] shadow-sm" onSubmit={onSubmit}>
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-lg font-semibold">Настройки сайта</h2>
        <button className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm" type="button" onClick={() => void loadSettings()}>
          Обновить
        </button>
      </div>

      {loading ? <p className="text-sm text-[rgb(var(--text-secondary))]">Загрузка настроек...</p> : (
        <textarea className="form-input min-h-[420px] w-full rounded-lg px-3 py-2 font-mono text-sm" value={rawSettings} onChange={(e) => setRawSettings(e.target.value)} />
      )}

      {status && <p className="text-sm text-emerald-700">{status}</p>}
      {error && <p className="field-error">{error}</p>}

      <button className="interactive-chip theme-button px-4 py-2" type="submit" disabled={saving || loading || !rawSettings.trim()}>
        {saving ? 'Сохраняем...' : 'Сохранить настройки'}
      </button>
    </form>
  );
};
