import { Link, Navigate, useLocation } from 'react-router-dom';

type SetupState = {
  email?: string;
  totp_secret?: string;
  qr_code_url?: string;
};

const toQrImageUrl = (otpAuthUrl: string) =>
  `https://api.qrserver.com/v1/create-qr-code/?size=260x260&data=${encodeURIComponent(otpAuthUrl)}`;

export const TwoFaSetupPage = () => {
  const location = useLocation();
  const state = (location.state ?? {}) as SetupState;

  if (!state.qr_code_url || !state.totp_secret) {
    return <Navigate to="/auth/register" replace />;
  }

  return (
    <section className="w-full space-y-4 rounded border border-slate-200 bg-white p-6">
      <h1 className="text-2xl font-semibold">Настройка 2FA</h1>
      <p className="text-sm text-slate-600">
        Отсканируй QR-код в Google Authenticator / Authy, затем используй код из приложения при логине.
      </p>

      <div className="flex justify-center rounded border border-slate-200 bg-white p-4">
        <img src={toQrImageUrl(state.qr_code_url)} alt="2FA QR code" className="h-64 w-64" />
      </div>

      <p className="text-sm">
        <span className="font-medium">Email:</span> {state.email ?? '—'}
      </p>
      <p className="text-sm">
        <span className="font-medium">Секрет (backup):</span> <code>{state.totp_secret}</code>
      </p>
      <p className="text-xs text-slate-500 break-all">OTPAUTH URL: {state.qr_code_url}</p>

      <Link
        to="/auth/login"
        state={{ message: '2FA настроен. Введи email, пароль и код из приложения.' }}
        className="inline-block rounded bg-slate-900 px-4 py-2 text-white"
      >
        Перейти ко входу
      </Link>
    </section>
  );
};
