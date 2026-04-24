import { BuildVersion } from '@/components/layout/build-version';
import { RouteTransition } from '@/components/layout/route-transition';
import { NavigationProgress } from '@/components/layout/navigation-progress';

export const AuthLayout = () => (
  <>
    <NavigationProgress />
    <div className="auth-gradient-shell">
      <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-4 pb-14">
        <RouteTransition className="w-full" />
      </main>
    </div>
    <BuildVersion />
  </>
);
