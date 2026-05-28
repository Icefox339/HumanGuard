import { NavLink } from 'react-router-dom';
import { useAuthStore } from '@/app/store/auth-store';

const mainLinks = [
  ['Панель', '/dashboard'],
  ['Сайты', '/sites'],
  ['Файлы', '/files'],
  ['Профиль', '/profile'],
  ['Ключи API', '/api-keys']
];

type SidebarProps = {
  isOpen: boolean;
  onClose: () => void;
};

export const Sidebar = ({ isOpen, onClose }: SidebarProps) => {
  const user = useAuthStore((s) => s.user);
  const isAdmin = user?.role === 'admin';

  return (
    <>
      {isOpen && (
        <button
          className="fixed inset-0 z-30 bg-slate-900/35 sm:hidden"
          onClick={onClose}
          aria-label="Закрыть меню"
          type="button"
        />
      )}
      <aside
        className={`theme-surface fixed inset-y-0 left-0 z-40 h-screen w-[85vw] max-w-72 overflow-y-auto border-r theme-border p-4 transition-transform duration-200 sm:static sm:z-auto sm:h-auto sm:w-64 sm:shrink-0 ${isOpen ? 'translate-x-0' : '-translate-x-full sm:translate-x-0'}`}
      >
        <h1 className="mb-6 flex items-center gap-2 text-xl font-semibold text-[rgb(var(--text-primary))]">
          <span className="brand-icon" aria-hidden="true">
            <svg viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M10 2.2 16 4.6v4.3c0 4.1-2.7 7-6 8.9-3.3-1.9-6-4.8-6-8.9V4.6l6-2.4Z" fill="currentColor" />
              <circle cx="10" cy="9.4" r="2.2" fill="white" fillOpacity="0.95" />
            </svg>
          </span>
          HumanGuard
        </h1>
        <nav className="flex flex-col gap-1">
          {mainLinks.map(([label, href]) => (
            <NavLink
              key={href}
              to={href}
              onClick={onClose}
              className={({ isActive }) => (isActive ? 'nav-link nav-link--active' : 'nav-link')}
            >
              {label}
            </NavLink>
          ))}

          {isAdmin && (
            <>
              <p className="mt-4 px-2 text-xs font-semibold uppercase tracking-wide text-[rgb(var(--text-secondary))]">
                Администрирование
              </p>
              <NavLink
                to="/admin/users"
                onClick={onClose}
                className={({ isActive }) => (isActive ? 'nav-link nav-link--active' : 'nav-link')}
              >Пользователи</NavLink>
              <NavLink
                to="/admin/tokens"
                onClick={onClose}
                className={({ isActive }) => (isActive ? 'nav-link nav-link--active' : 'nav-link')}
              >Менеджер токенов</NavLink>
            </>
          )}
        </nav>
      </aside>
    </>
  );
};
