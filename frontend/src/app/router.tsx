import { createBrowserRouter, Navigate } from 'react-router-dom';
import { AppLayout } from '@/components/layout/app-layout';
import { AuthLayout } from '@/components/layout/auth-layout';
import { ProtectedRoute } from '@/components/layout/protected-route';
import { LoginPage } from '@/pages/auth/login-page';
import { RegisterPage } from '@/pages/auth/register-page';
import { DashboardPage } from '@/pages/app/dashboard-page';
import { ProfilePage } from '@/pages/app/profile-page';
import { SitesPage } from '@/pages/app/sites-page';
import { SiteCreatePage } from '@/pages/app/site-create-page';
import { SiteDetailsPage } from '@/pages/app/site-details-page';
import { SiteSettingsPage } from '@/pages/app/site-settings-page';
import { SiteSessionsPage } from '@/pages/app/site-sessions-page';
import { SuspiciousSessionsPage } from '@/pages/app/suspicious-sessions-page';
import { SiteStatsPage } from '@/pages/app/site-stats-page';
import { NotFoundPage } from '@/pages/app/not-found-page';
import { UsersPage } from '@/pages/admin/users-page';

export const router = createBrowserRouter([
  {
    path: '/auth',
    element: <AuthLayout />,
    children: [
      { path: 'login', element: <LoginPage /> },
      { path: 'register', element: <RegisterPage /> }
    ]
  },
  {
    path: '/',
    element: (
      <ProtectedRoute>
        <AppLayout />
      </ProtectedRoute>
    ),
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: 'dashboard', element: <DashboardPage /> },
      { path: 'profile', element: <ProfilePage /> },
      { path: 'sites', element: <SitesPage /> },
      { path: 'sites/new', element: <SiteCreatePage /> },
      { path: 'sites/:siteId', element: <SiteDetailsPage /> },
      { path: 'sites/:siteId/settings', element: <SiteSettingsPage /> },
      { path: 'sites/:siteId/sessions', element: <SiteSessionsPage /> },
      { path: 'sites/:siteId/suspicious', element: <SuspiciousSessionsPage /> },
      { path: 'sites/:siteId/stats', element: <SiteStatsPage /> },
      { path: 'admin/users', element: <UsersPage /> },
      { path: '*', element: <NotFoundPage /> }
    ]
  }
]);
