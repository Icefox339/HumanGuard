import { useState } from 'react';
import { Header } from '@/components/layout/header';
import { Sidebar } from '@/components/layout/sidebar';
import { BuildVersion } from '@/components/layout/build-version';
import { RouteTransition } from '@/components/layout/route-transition';
import { NavigationProgress } from '@/components/layout/navigation-progress';

export const AppLayout = () => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

  return (
    <div className="app-shell flex min-h-screen text-[rgb(var(--text-primary))]">
      <Sidebar isOpen={isSidebarOpen} onClose={() => setIsSidebarOpen(false)} />
      <div className="flex flex-1 flex-col">
        <Header onToggleSidebar={() => setIsSidebarOpen((prev) => !prev)} />
        <NavigationProgress />
        <main className="p-4 pb-20 sm:p-6">
          <RouteTransition />
        </main>
      </div>
      <BuildVersion />
    </div>
  );
};
