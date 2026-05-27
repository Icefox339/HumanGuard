import { useState } from 'react';
import { PROJECT_VERSION } from '@/lib/runtime-config';

export const BuildVersion = () => {
  const [isVisible, setIsVisible] = useState(true);

  return (
    <div className="fixed bottom-3 right-3 z-50 flex items-end gap-2">
      {isVisible && (
        <footer className="rounded-md border border-slate-200 bg-white/95 px-3 py-2 text-xs text-slate-500 shadow-sm backdrop-blur">
          Build version: {PROJECT_VERSION}
        </footer>
      )}
      <button
        className="interactive-chip rounded-md border border-slate-200 bg-white/95 px-3 py-2 text-xs font-medium text-slate-600 shadow-sm"
        onClick={() => setIsVisible((prev) => !prev)}
        type="button"
      >
        {isVisible ? 'Hide build' : 'Show build'}
      </button>
    </div>
  );
};
