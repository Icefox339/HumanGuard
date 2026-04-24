import { Header } from '@/components/layout/header';
import { Sidebar } from '@/components/layout/sidebar';
import { BuildVersion } from '@/components/layout/build-version';
import { RouteTransition } from '@/components/layout/route-transition';
import { NavigationProgress } from '@/components/layout/navigation-progress';

export const AppLayout = () => (
  <div className="app-shell flex min-h-screen text-[rgb(var(--text-primary))]">
    <Sidebar />
    <div className="flex flex-1 flex-col">
      <Header />
      <NavigationProgress />
      <main className="p-6 pb-14">
        <RouteTransition />
      </main>
    </div>
    <BuildVersion />
  </div>
);
