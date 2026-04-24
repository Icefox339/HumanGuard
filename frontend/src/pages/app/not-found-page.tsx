import { Link } from 'react-router-dom';

export const NotFoundPage = () => (
  <section className="error-scene mx-auto flex min-h-[70vh] max-w-2xl flex-col items-center justify-center px-4 text-center">
    <div className="error-orb" aria-hidden="true" />
    <p className="error-code">404</p>
    <h1 className="text-3xl font-semibold text-slate-900">Кажется, эта страница потерялась</h1>
    <p className="mt-3 max-w-lg text-sm leading-6 text-slate-600">
      Ничего страшного — вернитесь в панель управления. Мы добавили мягкую анимацию, чтобы даже ошибки выглядели аккуратно.
    </p>
    <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
      <Link to="/dashboard" className="interactive-chip rounded-lg bg-slate-900 px-5 py-2.5 text-sm font-medium text-white">
        Открыть дашборд
      </Link>
      <Link to="/sites" className="interactive-chip rounded-lg border border-slate-300 bg-white px-5 py-2.5 text-sm font-medium text-slate-700">
        К сайтам
      </Link>
    </div>
  </section>
);
