type SpinnerProps = {
  className?: string;
  label?: string;
};

export const Spinner = ({ className = '', label = 'Loading' }: SpinnerProps) => (
  <span className={`inline-flex items-center gap-2 text-sm text-slate-600 ${className}`} role="status" aria-live="polite">
    <span className="h-4 w-4 animate-spin rounded-full border-2 border-slate-300 border-t-slate-800" aria-hidden="true" />
    <span>{label}</span>
  </span>
);
