export const BuildVersion = () => (
  <footer className="fixed bottom-0 left-0 right-0 border-t border-slate-200 bg-white/95 px-4 py-2 text-center text-xs text-slate-500 backdrop-blur">
    Build version: {import.meta.env.VITE_BUILD_VERSION}
  </footer>
);
