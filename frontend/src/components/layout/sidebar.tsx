import { NavLink } from 'react-router-dom';
import { useAuthStore } from '@/app/store/auth-store';

const links = [
  ['Dashboard', '/dashboard'],
  ['Sites', '/sites'],
  ['Files', '/files'],
  ['Profile', '/profile']
];

export const Sidebar = () => {
  const user = useAuthStore((s) => s.user);
  const isAdmin = user?.role === 'admin';

  return (
    <aside className="theme-surface w-64 border-r theme-border p-4">
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
        {links.map(([label, href]) => (
          <NavLink
            key={href}
            to={href}
            className={({ isActive }) =>
              isActive
                ? 'nav-link nav-link--active'
                : 'nav-link'
            }
          >
            {label}
          </NavLink>
        ))}
      </nav>

      {isAdmin && (
        <div className="mt-6 border-t border-[rgb(var(--border))] pt-4">
          <p className="mb-2 px-2 text-xs font-semibold uppercase tracking-wide text-[rgb(var(--text-secondary))]">Администрирование</p>
          <NavLink to="/admin/users" className={({ isActive }) => (isActive ? 'nav-link nav-link--active' : 'nav-link')}>
            Users
          </NavLink>
        </div>
      )}
    </aside>
  );
};
