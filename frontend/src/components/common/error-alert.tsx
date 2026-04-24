import type { ReactNode } from 'react';

type ErrorAlertProps = {
  title?: string;
  message: ReactNode;
};

export const ErrorAlert = ({ title = 'Ошибка', message }: ErrorAlertProps) => (
  <div className="theme-error route-transition rounded-xl px-4 py-3" role="alert">
    <p className="text-sm font-semibold">{title}</p>
    <div className="theme-error-subtle mt-1 text-sm">{message}</div>
  </div>
);
