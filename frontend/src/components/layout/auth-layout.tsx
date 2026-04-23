import { Outlet } from 'react-router-dom';

export const AuthLayout = () => (
  <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-4">
    <Outlet />
  </main>
);
