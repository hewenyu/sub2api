import React from 'react';
import { cn } from '@/utils/cn';

export interface BadgeProps {
  children: React.ReactNode;
  variant?:
    | 'default'
    | 'success'
    | 'warning'
    | 'error'
    | 'info'
    | 'active'
    | 'inactive'
    | 'rate-limited'
    | 'unauthorized'
    | 'overloaded'
    | 'pro'
    | 'max'
    | 'codex-type';
  className?: string;
}

export const Badge: React.FC<BadgeProps> = ({ children, variant = 'default', className }) => {
  const variants = {
    default: 'bg-gray-100 text-gray-800',
    success: 'bg-emerald-100 text-emerald-800',
    warning: 'bg-yellow-100 text-yellow-800',
    error: 'bg-red-100 text-red-800',
    info: 'bg-blue-100 text-blue-800',
    active: 'bg-emerald-100 text-emerald-800',
    inactive: 'bg-red-100 text-red-800',
    'rate-limited': 'bg-yellow-100 text-yellow-800',
    unauthorized: 'bg-red-100 text-red-800',
    overloaded: 'bg-pink-100 text-pink-800',
    pro: 'bg-blue-100 text-blue-800',
    max: 'bg-purple-100 text-purple-800',
    'codex-type': 'bg-green-100 text-green-800',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold',
        variants[variant],
        className
      )}
    >
      {children}
    </span>
  );
};
