import { Link, isRouteErrorResponse, useRouteError } from 'react-router-dom';

const getErrorMessage = (error: unknown) => {
  if (isRouteErrorResponse(error)) {
    if (error.status === 404) {
      return 'Страница не найдена. Возможно, ссылка устарела или была удалена.';
    }

    return `Ошибка ${error.status}: ${error.statusText || 'что-то пошло не так'}`;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return 'Произошла непредвиденная ошибка. Попробуйте перезагрузить страницу.';
};

export const RouteErrorPage = () => {
  const error = useRouteError();

  return (
    <section className="error-scene mx-auto flex min-h-[70vh] max-w-2xl flex-col items-center justify-center px-4 text-center">
      <div className="error-orb" aria-hidden="true" />
      <p className="error-code">Oops!</p>
      <h1 className="text-3xl font-semibold text-slate-900">Не удалось открыть страницу</h1>
      <p className="mt-3 max-w-lg text-sm leading-6 text-slate-600">{getErrorMessage(error)}</p>
      <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
        <Link to="/dashboard" className="interactive-chip rounded-lg bg-slate-900 px-5 py-2.5 text-sm font-medium text-white">
          На дашборд
        </Link>
        <button
          type="button"
          className="interactive-chip rounded-lg border border-slate-300 bg-white px-5 py-2.5 text-sm font-medium text-slate-700"
          onClick={() => window.location.reload()}
        >
          Перезагрузить
        </button>
      </div>
    </section>
  );
};
