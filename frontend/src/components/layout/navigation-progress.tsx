import { useEffect, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';

const PROGRESS_VISIBLE_MS = 550;

export const NavigationProgress = () => {
  const location = useLocation();
  const [isAnimating, setIsAnimating] = useState(false);
  const isFirstRender = useRef(true);

  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      return;
    }

    setIsAnimating(true);
    const timer = window.setTimeout(() => {
      setIsAnimating(false);
    }, PROGRESS_VISIBLE_MS);

    return () => window.clearTimeout(timer);
  }, [location.pathname, location.search, location.hash]);

  return <div className={isAnimating ? 'navigation-progress navigation-progress--active' : 'navigation-progress'} aria-hidden={!isAnimating} />;
};
