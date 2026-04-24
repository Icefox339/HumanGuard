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
    <section className="auth-card w-full space-y-4 rounded-2xl p-6">
      <h1 className="text-2xl font-semibold text-[rgb(var(--text-primary))]">Настройка 2FA</h1>
      <p className="theme-text-muted text-sm">
        Отсканируй QR-код в Google Authenticator / Authy, затем используй код из приложения при логине.
      </p>

      <div className="theme-info flex justify-center rounded-xl p-4">
        <img src={toQrImageUrl(state.qr_code_url)} alt="2FA QR code" className="h-64 w-64" />
      </div>

      <p className="text-sm">
        <span className="font-medium">Email:</span> {state.email ?? '—'}
      </p>
      <p className="text-sm">
        <span className="font-medium">Секрет (backup):</span> <code>{state.totp_secret}</code>
      </p>
      <p className="theme-text-muted break-all text-xs">OTPAUTH URL: {state.qr_code_url}</p>

      <Link
        to="/auth/login"
        state={{ message: '2FA настроен. Введи email, пароль и код из приложения.' }}
        className="interactive-chip theme-button inline-block px-4 py-2"
      >
        Перейти ко входу
      </Link>
    </section>
  );
};
