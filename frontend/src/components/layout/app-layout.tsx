import { Outlet } from 'react-router-dom';
import { Header } from '@/components/layout/header';
import { Sidebar } from '@/components/layout/sidebar';
import { BuildVersion } from '@/components/layout/build-version';

export const AppLayout = () => (
  <div className="flex min-h-screen bg-slate-50 text-slate-900">
    <Sidebar />
    <div className="flex flex-1 flex-col">
      <Header />
      <main className="p-6 pb-14">
        <Outlet />
      </main>
    </div>
    <BuildVersion />
  </div>
);
