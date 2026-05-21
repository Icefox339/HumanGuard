export type BehaviorTrackerOptions = {
  sessionId: string;
  siteId?: string;
  apiBaseUrl: string;
  flushIntervalMs?: number;
  maxMouseSamples?: number;
};

type MousePoint = { x: number; y: number; t: number };

type BehaviorPayload = {
  session_id: string;
  metrics: Record<string, unknown>;
};

export type BehaviorTracker = {
  flush: () => Promise<void>;
  stop: () => Promise<void>;
};

export const startBehaviorTracker = (options: BehaviorTrackerOptions): BehaviorTracker => {
  const {
    sessionId,
    siteId,
    apiBaseUrl,
    flushIntervalMs = 5000,
    maxMouseSamples = 40
  } = options;

  const metrics = {
    clicks: 0,
    keypresses: 0,
    scrollEvents: 0,
    scrollDistance: 0,
    mouseMoves: 0,
    mouseSamples: [] as MousePoint[],
    viewport: {
      width: window.innerWidth,
      height: window.innerHeight
    },
    startedAt: Date.now(),
    lastActivityAt: Date.now()
  };

  let lastScrollY = window.scrollY;
  let destroyed = false;

  const markActivity = () => {
    metrics.lastActivityAt = Date.now();
  };

  const onClick = () => {
    metrics.clicks += 1;
    markActivity();
  };

  const onKeyDown = () => {
    metrics.keypresses += 1;
    markActivity();
  };

  const onScroll = () => {
    const currentY = window.scrollY;
    metrics.scrollEvents += 1;
    metrics.scrollDistance += Math.abs(currentY - lastScrollY);
    lastScrollY = currentY;
    markActivity();
  };

  const onMouseMove = (event: MouseEvent) => {
    metrics.mouseMoves += 1;
    if (metrics.mouseSamples.length < maxMouseSamples) {
      metrics.mouseSamples.push({ x: event.clientX, y: event.clientY, t: Date.now() });
    }
    markActivity();
  };

  const onResize = () => {
    metrics.viewport = {
      width: window.innerWidth,
      height: window.innerHeight
    };
  };

  const buildPayload = (): BehaviorPayload => ({
    session_id: sessionId,
    metrics: {
      clicks: metrics.clicks,
      keypresses: metrics.keypresses,
      scroll_events: metrics.scrollEvents,
      scroll_distance: metrics.scrollDistance,
      mouse_moves: metrics.mouseMoves,
      mouse_samples: metrics.mouseSamples,
      viewport: metrics.viewport,
      started_at: metrics.startedAt,
      last_activity_at: metrics.lastActivityAt,
      sent_at: Date.now()
    }
  });

  const flush = async () => {
    if (destroyed) {
      return;
    }

    const payload = buildPayload();
    const endpoint = `${apiBaseUrl.replace(/\/$/, '')}/api/behavior/${sessionId}`;

    await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(siteId ? { 'X-Site-ID': siteId } : {})
      },
      body: JSON.stringify(payload),
      keepalive: true
    });

    metrics.mouseSamples = [];
  };

  const flushOnHide = () => {
    if (destroyed) {
      return;
    }

    const endpoint = `${apiBaseUrl.replace(/\/$/, '')}/api/behavior/${sessionId}`;
    const payload = JSON.stringify(buildPayload());

    if (navigator.sendBeacon) {
      const blob = new Blob([payload], { type: 'application/json' });
      navigator.sendBeacon(endpoint, blob);
      return;
    }

    void flush();
  };

  window.addEventListener('click', onClick, { passive: true });
  window.addEventListener('keydown', onKeyDown, { passive: true });
  window.addEventListener('scroll', onScroll, { passive: true });
  window.addEventListener('mousemove', onMouseMove, { passive: true });
  window.addEventListener('resize', onResize, { passive: true });
  document.addEventListener('visibilitychange', flushOnHide);
  window.addEventListener('beforeunload', flushOnHide);

  const timer = window.setInterval(() => {
    void flush();
  }, flushIntervalMs);

  return {
    flush,
    stop: async () => {
      if (destroyed) {
        return;
      }

      destroyed = true;
      window.clearInterval(timer);
      window.removeEventListener('click', onClick);
      window.removeEventListener('keydown', onKeyDown);
      window.removeEventListener('scroll', onScroll);
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('resize', onResize);
      document.removeEventListener('visibilitychange', flushOnHide);
      window.removeEventListener('beforeunload', flushOnHide);
      await flush();
    }
  };
};
