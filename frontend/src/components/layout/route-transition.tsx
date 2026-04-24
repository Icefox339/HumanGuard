import { Outlet, useLocation } from 'react-router-dom';

type RouteTransitionProps = {
  className?: string;
};

export const RouteTransition = ({ className }: RouteTransitionProps) => {
  const location = useLocation();

  return (
    <div key={location.pathname} className={className ? `route-transition ${className}` : 'route-transition'}>
      <Outlet />
    </div>
  );
};
