import { NavLink } from 'react-router-dom';

const links = [
  ['Dashboard', '/dashboard'],
  ['Sites', '/sites'],
  ['Profile', '/profile'],
  ['Users', '/admin/users']
];

export const Sidebar = () => (
  <aside className="w-64 border-r border-slate-200 bg-white p-4">
    <h1 className="mb-6 text-xl font-semibold">HumanGuard</h1>
    <nav className="flex flex-col gap-1">
      {links.map(([label, href]) => (
        <NavLink key={href} to={href} className="rounded px-3 py-2 text-sm hover:bg-slate-100">
          {label}
        </NavLink>
      ))}
    </nav>
  </aside>
);
