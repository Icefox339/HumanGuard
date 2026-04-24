import { Outlet } from 'react-router-dom';
import { BuildVersion } from '@/components/layout/build-version';

export const AuthLayout = () => (
  <>
    <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-4 pb-14">
      <Outlet />
    </main>
    <BuildVersion />
  </>
);
