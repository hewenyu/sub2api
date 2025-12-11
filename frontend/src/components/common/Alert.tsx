import React from 'react';
import { AlertCircle, CheckCircle, Info, XCircle, X } from 'lucide-react';
import { cn } from '@/utils/cn';

export interface AlertProps {
  variant?: 'success' | 'warning' | 'error' | 'info';
  title?: string;
  children: React.ReactNode;
  onClose?: () => void;
  className?: string;
}

export const Alert: React.FC<AlertProps> = ({
  variant = 'info',
  title,
  children,
  onClose,
  className,
}) => {
  const variants = {
    success: {
      container: 'bg-green-50 border-green-200 text-green-800',
      icon: CheckCircle,
      iconColor: 'text-green-600',
    },
    warning: {
      container: 'bg-yellow-50 border-yellow-200 text-yellow-800',
      icon: AlertCircle,
      iconColor: 'text-yellow-600',
    },
    error: {
      container: 'bg-red-50 border-red-200 text-red-800',
      icon: XCircle,
      iconColor: 'text-red-600',
    },
    info: {
      container: 'bg-blue-50 border-blue-200 text-blue-800',
      icon: Info,
      iconColor: 'text-blue-600',
    },
  };

  const { container, icon: Icon, iconColor } = variants[variant];

  return (
    <div className={cn('rounded-lg border p-4', container, className)}>
      <div className="flex items-start">
        <Icon className={cn('h-5 w-5 flex-shrink-0', iconColor)} />
        <div className="ml-3 flex-1">
          {title && <h3 className="text-sm font-medium">{title}</h3>}
          <div className={cn('text-sm', title && 'mt-1')}>{children}</div>
        </div>
        {onClose && (
          <button
            onClick={onClose}
            className={cn('ml-3 flex-shrink-0 rounded-md p-1 hover:bg-black/5', iconColor)}
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>
    </div>
  );
};
