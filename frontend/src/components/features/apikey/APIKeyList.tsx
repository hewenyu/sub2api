import React from 'react';
import { Copy, Edit, Trash2, Power, PowerOff } from 'lucide-react';
import { Table } from '@/components/common/Table';
import type { Column } from '@/components/common/Table';
import { Button } from '@/components/common/Button';
import { Badge } from '@/components/common/Badge';
import { formatCost } from '@/utils/format';
import type { APIKey } from '@/types/apikey';

interface APIKeyListProps {
  data: APIKey[];
  loading?: boolean;
  onEdit: (apiKey: APIKey) => void;
  onDelete: (id: number) => void;
  onToggle: (id: number, isActive: boolean) => void;
}

export const APIKeyList: React.FC<APIKeyListProps> = ({
  data,
  loading,
  onEdit,
  onDelete,
  onToggle,
}) => {
  const handleCopy = async (prefix: string) => {
    try {
      await navigator.clipboard.writeText(prefix);
      console.log('API Key prefix copied to clipboard');
    } catch (error) {
      console.error('Failed to copy:', error);
    }
  };

  const formatDate = (timestamp: number): string => {
    if (!timestamp) return '-';
    const date = new Date(timestamp * 1000);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const columns: Column<APIKey>[] = [
    {
      key: 'name',
      title: 'Name',
      width: '20%',
      render: (value) => <span className="font-medium text-gray-900">{String(value)}</span>,
    },
    {
      key: 'key_prefix',
      title: 'Key Prefix',
      width: '20%',
      render: (value) => (
        <div className="flex items-center space-x-2">
          <code className="bg-gray-100 px-2 py-1 rounded text-xs font-mono">
            {String(value)}...
          </code>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleCopy(String(value))}
            className="p-1"
          >
            <Copy className="w-3 h-3" />
          </Button>
        </div>
      ),
    },
    {
      key: 'is_active',
      title: 'Status',
      width: '10%',
      render: (value) => (
        <Badge variant={value ? 'success' : 'default'}>{value ? 'Active' : 'Inactive'}</Badge>
      ),
    },
    {
      key: 'total_requests',
      title: 'Requests',
      width: '10%',
      render: (value) => (
        <span className="text-gray-900">
          {typeof value === 'number' ? value.toLocaleString() : '0'}
        </span>
      ),
    },
    {
      key: 'total_cost',
      title: 'Total Cost',
      width: '10%',
      render: (value) => (
        <span className="text-gray-900 font-medium">
          {formatCost(typeof value === 'number' ? value : 0)}
        </span>
      ),
    },
    {
      key: 'created_at',
      title: 'Created',
      width: '15%',
      render: (value) => (
        <span className="text-gray-600 text-sm">
          {formatDate(typeof value === 'number' ? value : 0)}
        </span>
      ),
    },
    {
      key: 'id',
      title: 'Actions',
      width: '15%',
      render: (value, record) => (
        <div className="flex items-center space-x-1">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onToggle(Number(value), !record.is_active)}
            className="p-2"
            title={record.is_active ? 'Deactivate' : 'Activate'}
          >
            {record.is_active ? (
              <PowerOff className="w-4 h-4 text-orange-600" />
            ) : (
              <Power className="w-4 h-4 text-green-600" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onEdit(record)}
            className="p-2"
            title="Edit"
          >
            <Edit className="w-4 h-4 text-blue-600" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onDelete(Number(value))}
            className="p-2"
            title="Delete"
          >
            <Trash2 className="w-4 h-4 text-red-600" />
          </Button>
        </div>
      ),
    },
  ];

  return <Table<APIKey> columns={columns} data={data} loading={loading} />;
};
