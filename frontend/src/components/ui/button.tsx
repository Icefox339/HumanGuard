import type { ButtonHTMLAttributes, PropsWithChildren } from 'react';

type ButtonVariant = 'primary' | 'secondary' | 'ghost';

type ButtonProps = PropsWithChildren<ButtonHTMLAttributes<HTMLButtonElement>> & {
  variant?: ButtonVariant;
};

const variants: Record<ButtonVariant, string> = {
  primary: 'bg-slate-900 text-white hover:bg-slate-700',
  secondary: 'border border-slate-300 bg-white text-slate-800 hover:bg-slate-100',
  ghost: 'bg-transparent text-slate-700 hover:bg-slate-100'
};

export const Button = ({ children, className = '', type = 'button', variant = 'primary', ...props }: ButtonProps) => (
  <button
    type={type}
    className={`interactive-chip inline-flex items-center justify-center rounded-lg px-4 py-2 text-sm font-medium ${variants[variant]} ${className}`}
    {...props}
  >
    {children}
  </button>
);
