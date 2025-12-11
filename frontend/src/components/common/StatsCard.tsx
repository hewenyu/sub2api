import React from 'react';
import { cn } from '@/utils/cn';

interface StatsCardProps {
  label: string;
  value: string | number;
  icon: React.ReactNode;
  change?: string;
  changeType?: 'positive' | 'negative' | 'neutral';
  accentColor?: 'blue' | 'purple' | 'green' | 'orange';
}

export const StatsCard: React.FC<StatsCardProps> = ({
  label,
  value,
  icon,
  change,
  changeType = 'neutral',
  accentColor = 'blue',
}) => {
  const accentColors = {
    blue: 'border-blue-500',
    purple: 'border-purple-500',
    green: 'border-green-500',
    orange: 'border-orange-500',
  };

  const iconBgColors = {
    blue: 'bg-blue-100',
    purple: 'bg-purple-100',
    green: 'bg-green-100',
    orange: 'bg-orange-100',
  };

  const iconTextColors = {
    blue: 'text-blue-600',
    purple: 'text-purple-600',
    green: 'text-green-600',
    orange: 'text-orange-600',
  };

  const changeColors = {
    positive: 'text-green-600',
    negative: 'text-red-600',
    neutral: 'text-gray-600',
  };

  return (
    <div className={cn('bg-white rounded-lg shadow p-6 border-l-4', accentColors[accentColor])}>
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <p className="text-gray-600 text-sm font-semibold">{label}</p>
          <p className="text-3xl font-bold text-gray-900 mt-2">{value}</p>
          {change && (
            <p className={cn('text-sm mt-2', changeColors[changeType])}>
              {changeType === 'positive' && '↑ '}
              {changeType === 'negative' && '↓ '}
              {change}
            </p>
          )}
        </div>
        <div
          className={cn(
            'w-12 h-12 rounded-lg flex items-center justify-center',
            iconBgColors[accentColor]
          )}
        >
          <div className={cn('w-6 h-6', iconTextColors[accentColor])}>{icon}</div>
        </div>
      </div>
    </div>
  );
};
