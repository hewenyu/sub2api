import React from 'react';
import { cn } from '@/utils/cn';

export interface CardProps {
  title?: string;
  description?: string;
  children: React.ReactNode;
  actions?: React.ReactNode;
  className?: string;
}

export const Card: React.FC<CardProps> = ({
  title,
  description,
  children,
  actions,
  className,
}) => {
  return (
    <div className={cn('rounded-md border border-gray-200 bg-white shadow-md', className)}>
      {(title || description || actions) && (
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <div>
            {title && <h3 className="text-lg font-semibold text-gray-900">{title}</h3>}
            {description && <p className="mt-1 text-sm text-gray-600">{description}</p>}
          </div>
          {actions && <div className="flex items-center gap-2">{actions}</div>}
        </div>
      )}
      <div className="px-6 py-6">{children}</div>
    </div>
  );
};
