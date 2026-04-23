import { NavLink, Outlet } from 'react-router-dom';

export const AuthLayout = () => (
  <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-4">
    <div className="w-full space-y-4">
      <div className="grid grid-cols-2 rounded-lg border border-slate-200 bg-white p-1 text-sm">
        <NavLink
          to="/auth/login"
          className={({ isActive }) =>
            `rounded-md px-3 py-2 text-center ${isActive ? 'bg-slate-900 text-white' : 'text-slate-600'}`
          }
        >
          Sign in
        </NavLink>
        <NavLink
          to="/auth/register"
          className={({ isActive }) =>
            `rounded-md px-3 py-2 text-center ${isActive ? 'bg-slate-900 text-white' : 'text-slate-600'}`
          }
        >
          Create user
        </NavLink>
      </div>
      <Outlet />
    </div>
  </main>
);
