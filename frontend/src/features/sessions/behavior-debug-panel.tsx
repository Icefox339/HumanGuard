import { useEffect, useState } from 'react';

type DebugMetrics = {
  clicks: number;
  keypresses: number;
  scrollEvents: number;
  scrollDistance: number;
  mouseMoves: number;
  viewport: { width: number; height: number };
  startedAt: number;
  lastActivityAt: number;
};

const createInitialMetrics = (): DebugMetrics => ({
  clicks: 0,
  keypresses: 0,
  scrollEvents: 0,
  scrollDistance: 0,
  mouseMoves: 0,
  viewport: { width: window.innerWidth, height: window.innerHeight },
  startedAt: Date.now(),
  lastActivityAt: Date.now()
});

const formatTime = (value: number) => new Date(value).toLocaleTimeString();

export const BehaviorDebugPanel = () => {
  const [metrics, setMetrics] = useState<DebugMetrics>(() => createInitialMetrics());
  const [open, setOpen] = useState(false);

  useEffect(() => {
    let lastScrollY = window.scrollY;

    const markActivity = () => {
      setMetrics((prev) => ({ ...prev, lastActivityAt: Date.now() }));
    };

    const onClick = () => {
      setMetrics((prev) => ({ ...prev, clicks: prev.clicks + 1 }));
      markActivity();
    };

    const onKeyDown = () => {
      setMetrics((prev) => ({ ...prev, keypresses: prev.keypresses + 1 }));
      markActivity();
    };

    const onScroll = () => {
      const currentY = window.scrollY;
      const delta = Math.abs(currentY - lastScrollY);
      lastScrollY = currentY;
      setMetrics((prev) => ({
        ...prev,
        scrollEvents: prev.scrollEvents + 1,
        scrollDistance: prev.scrollDistance + delta
      }));
      markActivity();
    };

    const onMouseMove = () => {
      setMetrics((prev) => ({ ...prev, mouseMoves: prev.mouseMoves + 1 }));
      markActivity();
    };

    const onResize = () => {
      setMetrics((prev) => ({
        ...prev,
        viewport: { width: window.innerWidth, height: window.innerHeight }
      }));
    };

    window.addEventListener('click', onClick, { passive: true });
    window.addEventListener('keydown', onKeyDown, { passive: true });
    window.addEventListener('scroll', onScroll, { passive: true });
    window.addEventListener('mousemove', onMouseMove, { passive: true });
    window.addEventListener('resize', onResize, { passive: true });

    return () => {
      window.removeEventListener('click', onClick);
      window.removeEventListener('keydown', onKeyDown);
      window.removeEventListener('scroll', onScroll);
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('resize', onResize);
    };
  }, []);

  return (
    <div className="theme-card rounded-2xl p-5">
      <div className="mb-3 flex items-center justify-between gap-3">
        <h2 className="text-lg font-semibold text-[rgb(var(--text-primary))]">Behavior debug</h2>
        <button
          type="button"
          className="interactive-chip theme-button"
          onClick={() => setOpen((prev) => !prev)}
        >
          {open ? 'Скрыть' : 'Показать'}
        </button>
      </div>

      {open ? (
        <div className="grid grid-cols-1 gap-2 text-sm text-[rgb(var(--text-secondary))] sm:grid-cols-2">
          <p>Clicks: <span className="font-medium text-[rgb(var(--text-primary))]">{metrics.clicks}</span></p>
          <p>Keypresses: <span className="font-medium text-[rgb(var(--text-primary))]">{metrics.keypresses}</span></p>
          <p>Scroll events: <span className="font-medium text-[rgb(var(--text-primary))]">{metrics.scrollEvents}</span></p>
          <p>Scroll distance: <span className="font-medium text-[rgb(var(--text-primary))]">{metrics.scrollDistance}px</span></p>
          <p>Mouse moves: <span className="font-medium text-[rgb(var(--text-primary))]">{metrics.mouseMoves}</span></p>
          <p>Viewport: <span className="font-medium text-[rgb(var(--text-primary))]">{metrics.viewport.width}×{metrics.viewport.height}</span></p>
          <p>Started: <span className="font-medium text-[rgb(var(--text-primary))]">{formatTime(metrics.startedAt)}</span></p>
          <p>Last activity: <span className="font-medium text-[rgb(var(--text-primary))]">{formatTime(metrics.lastActivityAt)}</span></p>
        </div>
      ) : null}
    </div>
  );
};
