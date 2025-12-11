import React from 'react';
import { Edit, Trash2, Power, PowerOff, RefreshCw } from 'lucide-react';
import { Table } from '@/components/common/Table';
import type { Column } from '@/components/common/Table';
import { Button } from '@/components/common/Button';
import { Badge } from '@/components/common/Badge';
import type { ClaudeAccount } from '@/types/account';

interface ClaudeAccountListProps {
  data: ClaudeAccount[];
  loading?: boolean;
  onEdit: (account: ClaudeAccount) => void;
  onDelete: (id: number) => void;
  onToggle: (id: number, isActive: boolean) => void;
  onRefreshToken: (id: number) => void;
}

export const ClaudeAccountList: React.FC<ClaudeAccountListProps> = ({
  data,
  loading,
  onEdit,
  onDelete,
  onToggle,
  onRefreshToken,
}) => {
  const formatDate = (dateString: string | null): string => {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusBadge = (account: ClaudeAccount) => {
    if (!account.is_active) {
      return <Badge variant="inactive">Inactive</Badge>;
    }
    if (account.rate_limited_until && new Date(account.rate_limited_until) > new Date()) {
      return <Badge variant="rate-limited">Rate Limited</Badge>;
    }
    if (account.overload_until && new Date(account.overload_until) > new Date()) {
      return <Badge variant="overloaded">Overloaded</Badge>;
    }
    return <Badge variant="active">Active</Badge>;
  };

  const columns: Column<ClaudeAccount>[] = [
    {
      key: 'email',
      title: 'Email',
      width: '25%',
      render: (value) => {
        const email = String(value);
        return (
          <div className="flex items-center">
            <div className="w-8 h-8 bg-purple-100 rounded-full flex items-center justify-center mr-3">
              <span className="text-purple-600 font-semibold text-sm">
                {email?.[0]?.toUpperCase() || 'C'}
              </span>
            </div>
            <span className="font-medium text-gray-900">{email}</span>
          </div>
        );
      },
    },
    {
      key: 'id',
      title: 'Status',
      width: '15%',
      render: (_value, record) => getStatusBadge(record),
    },
    {
      key: 'subscription_level',
      title: 'Subscription Level',
      width: '15%',
      render: (value) => {
        const level = String(value);
        return (
          <Badge variant={level === 'max' ? 'max' : level === 'pro' ? 'pro' : 'default'}>
            {level.toUpperCase()}
          </Badge>
        );
      },
    },
    {
      key: 'priority',
      title: 'Priority',
      width: '10%',
      render: (value) => <span className="text-gray-900">{String(value)}</span>,
    },
    {
      key: 'last_used_at',
      title: 'Last Used',
      width: '20%',
      render: (value) => (
        <span className="text-gray-600 text-sm">{formatDate(value as string | null)}</span>
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
            onClick={() => onRefreshToken(Number(value))}
            className="p-2"
            title="Refresh Token"
          >
            <RefreshCw className="w-4 h-4 text-blue-600" />
          </Button>
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

  return <Table<ClaudeAccount> columns={columns} data={data} loading={loading} />;
};
