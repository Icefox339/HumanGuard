import type { ReactNode } from 'react';

type ErrorAlertProps = {
  title?: string;
  message: ReactNode;
};

export const ErrorAlert = ({ title = 'Ошибка', message }: ErrorAlertProps) => (
  <div className="route-transition rounded-xl border border-rose-200 bg-rose-50/70 px-4 py-3 text-rose-800" role="alert">
    <p className="text-sm font-semibold">{title}</p>
    <div className="mt-1 text-sm text-rose-700">{message}</div>
  </div>
);
