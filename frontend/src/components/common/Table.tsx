import React from 'react';
import { ChevronUp, ChevronDown } from 'lucide-react';
import { cn } from '@/utils/cn';

export interface Column<T> {
  key: string;
  title: string;
  sortable?: boolean;
  render?: (value: unknown, record: T, index: number) => React.ReactNode;
  width?: string;
}

export interface TableProps<T> {
  columns: Column<T>[];
  data: T[];
  loading?: boolean;
  onSort?: (key: string, direction: 'asc' | 'desc') => void;
  sortKey?: string;
  sortDirection?: 'asc' | 'desc';
  className?: string;
}

export function Table<T extends Record<string, unknown>>({
  columns,
  data,
  loading = false,
  onSort,
  sortKey,
  sortDirection,
  className,
}: TableProps<T>) {
  const handleSort = (key: string) => {
    if (!onSort) return;

    const newDirection =
      sortKey === key && sortDirection === 'asc' ? 'desc' : 'asc';
    onSort(key, newDirection);
  };

  return (
    <div className={cn('overflow-hidden rounded-lg border border-gray-200', className)}>
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-100">
            <tr>
              {columns.map((column) => (
                <th
                  key={column.key}
                  className={cn(
                    'px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-700',
                    column.sortable && 'cursor-pointer select-none hover:bg-gray-200'
                  )}
                  style={{ width: column.width }}
                  onClick={() => column.sortable && handleSort(column.key)}
                >
                  <div className="flex items-center gap-2">
                    {column.title}
                    {column.sortable && (
                      <div className="flex flex-col">
                        <ChevronUp
                          className={cn(
                            'h-3 w-3',
                            sortKey === column.key && sortDirection === 'asc'
                              ? 'text-primary'
                              : 'text-gray-400'
                          )}
                        />
                        <ChevronDown
                          className={cn(
                            'h-3 w-3 -mt-1',
                            sortKey === column.key && sortDirection === 'desc'
                              ? 'text-primary'
                              : 'text-gray-400'
                          )}
                        />
                      </div>
                    )}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 bg-white">
            {loading ? (
              <tr>
                <td
                  colSpan={columns.length}
                  className="h-[200px] bg-gray-50 px-4 text-center text-sm text-gray-500"
                >
                  Loading...
                </td>
              </tr>
            ) : data.length === 0 ? (
              <tr>
                <td
                  colSpan={columns.length}
                  className="h-[200px] bg-gray-50 px-4 text-center text-sm text-gray-500"
                >
                  No data available
                </td>
              </tr>
            ) : (
              data.map((record, index) => (
                <tr
                  key={index}
                  className="hover:bg-gray-50 transition-colors"
                >
                  {columns.map((column) => (
                    <td
                      key={column.key}
                      className="whitespace-nowrap px-4 py-4 text-sm text-gray-900"
                    >
                      {column.render
                        ? column.render(record[column.key], record, index)
                        : String(record[column.key] ?? '-')}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
